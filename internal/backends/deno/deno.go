// Package deno provides backends for Deno using deno.land/x
package deno

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	// "net/url"
	// "os"
	// "regexp"

	// "github.com/hashicorp/go-version"

	"github.com/replit/upm/internal/api"
	"github.com/replit/upm/internal/util"
)

const DENO_X_DATABASE = "https://raw.githubusercontent.com/denoland/deno_website2/master/database.json"

// lockJSON represents the relevant data in lock.json
type lockJSON struct {
	Dependencies map[string]struct {
		Version string `json:"version"`
	} `json:"dependencies"`
}

func getJson(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return bodyBytes
}

// denoPatterns are FilenamePatterns for DenoLandXBackend
var denoPatterns = []string{"*.js", "*.ts"}

// Implements ListLockfile for deno-deno.land/x backend
func listDenoLockFile() map[api.PkgName]api.PkgVersion {
	contentsB, err := ioutil.ReadFile("lock.json")
	if err != nil {
		util.Die("lock.json: %s", err)
	}
	var cfg lockJSON
	if err := json.Unmarshal(contentsB, &cfg); err != nil {
		util.Die("lock.json: %s", err)
	}
	pkgs := map[api.PkgName]api.PkgVersion{}
	for nameStr, data := range cfg.Dependencies {
		pkgs[api.PkgName(nameStr)] = api.PkgVersion(data.Version)
	}
	return pkgs
}

// DenoXPackage: Package metadata, as accessible on DENO_X_DATABASE
type DenoXPackage struct {
	Type           string `json:"type"`
	Owner          string `json:"owner"`
	Repo           string `json:"repo"`
	Description    string `json:"desc"`
	DefaultVersion string `json:"default_version"`
	Path           string `json:"path"`
	Package        string `json:"package"`
}

func searchDenoPackage(query string) []api.PkgInfo {

	pkgString := getJson(DENO_X_DATABASE)
	var i map[string]DenoXPackage
	json.Unmarshal(pkgString, &i)
	fmt.Println(i["oak"])

	var pkgs []api.PkgInfo
	return pkgs
}

// Implements Info for deno-deno.land/x backend
func denoPkgInfo(name api.PkgName) api.PkgInfo {
	return api.PkgInfo{
		Name:          "name",
		Description:   "desc",
		Version:       "version",
		HomepageURL:   "homepage",
		SourceCodeURL: "source code",
		BugTrackerURL: "bugs tracker",
		Author: util.AuthorInfo{
			Name:  "Author.Name",
			Email: "Author.Email",
			URL:   "Author.URL",
		}.String(),
		License: "License",
	}
}

// DenoLandXBackend is a UPM backend for deno that uses the list at deno.land/x
var DenoLandXBackend = api.LanguageBackend{
	Name:             "deno-deno.land/x",
	Specfile:         "deps.ts", // temporary
	Lockfile:         "lock.json",
	FilenamePatterns: denoPatterns,
	GetPackageDir: func() string {
		return ".deno"
	},
	Lock: func() {},
	Info: denoPkgInfo,
	ListSpecfile: func() map[api.PkgName]api.PkgSpec {
		pkgs := map[api.PkgName]api.PkgSpec{}
		return pkgs
	},
	ListLockfile: listDenoLockFile,
	Add:          func(pkgs map[api.PkgName]api.PkgSpec) {},
	Remove:       func(pkgs map[api.PkgName]bool) {},
	Search:       searchDenoPackage,
	Install:      func() {},
}
