package slug

import (
	"archive/tar"
	"code.google.com/p/go-netrc/netrc"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/naaman/pf"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type ProcessTable struct {
	ProcessTypes map[string]string `json:"process_types"`
}

type Slug struct {
	Blob         map[string]string `json:"blob"`
	Commit       *string           `json:"commit"`
	CreatedAt    time.Time         `json:"created_at"`
	Id           string            `json:"id"`
	ProcessTypes map[string]string `json:"process_types"`
	UpdatedAt    time.Time         `json:"updated_at"`
	slugDir      string
	httpClient   *http.Client
	release      *Release
	tarFile      *os.File
	apiKey       string
	appName      string
}

type Release struct {
	Version int `json:"version"`
}

func NewSlug(slugDir string) *Slug {
	slugJson := &Slug{}
	slugJson.slugDir = slugDir

	client := &http.DefaultClient
	res, _ := client.Do(slugJson.createSlug())
	bod, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	json.Unmarshal(bod, &slugJson)
	slugJson.httpClient = *client
	slugJson.slugDir = slugDir
	return slugJson
}

func (s *Slug) Archive() {
	s.tarFile = tarGz(strings.TrimRight(s.slugDir, "/"))
}

func (s *Slug) Push() {
	_, err := s.httpClient.Do(s.putSlug())
	defer s.tarFile.Close()
	if err != nil {
		panic(err)
	}
}

func (s *Slug) Release() {
	res, _ := s.httpClient.Do(s.createRelease())
	bod, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	releaseJson := &Release{}
	json.Unmarshal(bod, &releaseJson)
	s.release = releaseJson
}

func (s *Slug) herokuReq(method string, resource string, body string) *http.Request {
	reqUrl := fmt.Sprintf("https://api.heroku.com/apps/%s/%s", s.appName, resource)
	req, _ := http.NewRequest(method, reqUrl, strings.NewReader(body))
	req.Header.Add("Accept", "application/vnd.heroku+json; version=3")
	req.SetBasicAuth("", s.apiKey)
	return req
}

func (s *Slug) herokuPost(resource string, body string) *http.Request {
	req := s.herokuReq("POST", resource, body)
	req.Header.Add("Content-Type", "application/json")
	return req
}

func (s *Slug) createSlug() *http.Request {
	procTable := new(ProcessTable)
	procTable.ProcessTypes = make(map[string]string)
	procfile := s.parseProcfile()
	for _, e := range procfile.Entries {
		procTable.ProcessTypes[e.Type] = e.Command
	}
	procTableJson, _ := json.Marshal(procTable)
	return s.herokuPost("slugs", string(procTableJson))
}

func (s *Slug) createRelease() *http.Request {
	slugJson := fmt.Sprintf(`{"slug":"%s"}`, s.Id)
	return s.herokuPost("releases", slugJson)
}

func (s *Slug) putSlug() *http.Request {
	tarFileStat, err := s.tarFile.Stat()
	tarFile, _ := os.Open(s.tarFile.Name())
	if err != nil {
		panic(err)
	}
	req, _ := http.NewRequest("PUT", s.Blob["put"], tarFile)
	req.ContentLength = tarFileStat.Size()
	return req
}

func (s *Slug) parseProcfile() *pf.Procfile {
	procfileFile, _ := os.Open(s.slugDir + "/Procfile")
	procfile, _ := pf.ParseProcfile(procfileFile)
	return procfile
}

func handleError(_e error) {
	if _e != nil {
		log.Fatal(_e)
		panic(_e)
	}
}

func targzWalk(dirPath string, tw *tar.Writer) {
	var walkfunc filepath.WalkFunc

	walkfunc = func(path string, fi os.FileInfo, err error) error {
		h, err := tar.FileInfoHeader(fi, "")
		handleError(err)

		h.Name = "./app/" + path

		if fi.Mode()&os.ModeSymlink != 0 {
			linkPath, err := os.Readlink(path)
			handleError(err)
			h.Linkname = linkPath
		}

		err = tw.WriteHeader(h)
		handleError(err)

		if fi.Mode()&os.ModeDir == 0 && fi.Mode()&os.ModeSymlink == 0 {
			fr, err := os.Open(path)
			handleError(err)
			defer fr.Close()

			_, err = io.Copy(tw, fr)
			handleError(err)
		}
		return nil
	}

	filepath.Walk(dirPath, walkfunc)
}

func tarGz(inPath string) *os.File {
	wd, _ := os.Getwd()
	os.Chdir(inPath)
	// file write
	tarFile, _ := ioutil.TempFile("", "slug")
	tarFileName := tarFile.Name() + ".tgz"
	os.Rename(tarFile.Name(), tarFileName)
	fw, err := os.Create(tarFileName)
	handleError(err)

	// gzip write
	gw := gzip.NewWriter(fw)
	defer gw.Close()

	// tar write
	tw := tar.NewWriter(gw)
	defer tw.Close()

	targzWalk(".", tw)

	os.Chdir(wd)
	return fw
}

func netrcApiKey() string {
	if u, err := user.Current(); err == nil {
		netrcPath := u.HomeDir + "/.netrc"
		if _, err := os.Stat(netrcPath); err == nil {
			key, _ := netrc.FindMachine(netrcPath, "api.heroku.com")
			if key.Password != "" {
				return key.Password
			}
		}
	}
	return ""
}
