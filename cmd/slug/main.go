package main

import (
	"code.google.com/p/go-netrc/netrc"
	"flag"
	"fmt"
	"github.com/naaman/slug"
	"os"
	"os/user"
)

var (
	fDir       = flag.String("dir", "", "-dir /path/to/src")
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
	s.Archive()
	fmt.Printf("done\nPushing %s...", s.TarFile.Name())
	s.Push()
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
			key, _ := netrc.FindMachine(netrcPath, "api.heroku.com")
			if key.Password != "" {
				return key.Password
			}
		}
	}
	return ""
}
