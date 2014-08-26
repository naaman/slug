package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"

	"code.google.com/p/go-netrc/netrc"
	"github.com/naaman/slug"
)

var (
	fDir       = flag.String("dir", "", "-dir /path/to/src")
	fTar       = flag.String("tar", "", "-tar /path/to/tarball")
	fAppName   = flag.String("app", "", "-app exampleapp")
	fRelease   = flag.Bool("release", false, "-release true")
	fApiKey    = flag.String("key", netrcApiKey(), "-key 123ABC")
	workingDir string
)

func init() {
	flag.Parse()
	if *fDir != "" {
		workingDir = *fDir
	} else {
		workingDir, _ = os.Getwd()
	}
	if *fAppName == "" {
		flag.Usage()
		os.Exit(1)
	}

}

func main() {
	fmt.Printf("Initializing slug for %s...", workingDir)
	s := slug.NewSlug(*fApiKey, *fAppName, workingDir)
	fmt.Printf("done\nArchiving %s...", workingDir)
	var tarFile *os.File
	if *fTar != "" {
		var err error
		tarFile, err = os.Open(*fTar)
		if err != nil {
			log.Println(err)
			return
		}
		s.SetArchive(tarFile)
	} else {
		tarFile = s.Archive()
	}

	fmt.Printf("done\nPushing %s...", tarFile.Name())
	err := s.Push()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("done\n")
	if *fRelease {
		fmt.Printf("Releasing...")
		release := s.Release()
		fmt.Printf("done (v%d)", release.Version)
	}
}

func netrcApiKey() string {
	if u, err := user.Current(); err == nil {
		netrcPath := u.HomeDir + "/.netrc"
		if _, err := os.Stat(netrcPath); err == nil {
			key, err := netrc.FindMachine(netrcPath, "api.heroku.com")
			if err != nil {
				return ""
			}
			if key.Password != "" {
				return key.Password
			}
		}
	}
	return ""
}
