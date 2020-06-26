// Package deno provides backends for Deno using deno.land/x
package deno

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	// "net/url"
	// "os"
	// "regexp"

	// "github.com/hashicorp/go-version"

	"github.com/agnivade/levenshtein"
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

// denoPatterns are FilenamePatterns for Deno files.
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
	Name           string
	Type           string `json:"type"`
	Owner          string `json:"owner"`
	Repo           string `json:"repo"`
	Description    string `json:"desc"`
	DefaultVersion string `json:"default_version"`
	Path           string `json:"path"`
	Package        string `json:"package"`
}

type PackageMatch struct {
	Package             DenoXPackage
	LevenshteinDistance int
}
type byLeven []PackageMatch

func (s byLeven) Len() int {
	return len(s)
}
func (s byLeven) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byLeven) Less(i, j int) bool {
	return s[i].LevenshteinDistance < s[j].LevenshteinDistance
}

func queryDenoX(denoPackages map[string]DenoXPackage, query string, count int) []DenoXPackage {
	var pkgs []DenoXPackage
	var matches []PackageMatch

	if len(query) == 0 {
		return pkgs
	}

	for pkgName := range denoPackages {
		matches = append(matches, PackageMatch{
			Package:             denoPackages[pkgName],
			LevenshteinDistance: levenshtein.ComputeDistance(pkgName, query),
		})
	}

	sort.Sort(byLeven(matches))

	for i, match := range matches {
		if i > 10 || match.LevenshteinDistance > 2 {
			break
		}
		pkgs = append(pkgs, match.Package)
	}
	return pkgs
}

func searchDenoPackage(query string) []api.PkgInfo {
	// Download packages
	pkgString := getJson(DENO_X_DATABASE)
	var denoPackages map[string]DenoXPackage
	if err := json.Unmarshal(pkgString, &denoPackages); err != nil {
		util.Die("Error downloading deno packages: %s", err)
	}
	for pkgName := range denoPackages {
		p := denoPackages[pkgName]
		p.Name = pkgName
		denoPackages[pkgName] = p
	}

	denoPkgs := queryDenoX(denoPackages, query, 10)

	var packages []api.PkgInfo
	for _, denoPkg := range denoPkgs {
		pkg := api.PkgInfo{
			Name:        denoPkg.Name,
			Description: denoPkg.Description,
			Author:      denoPkg.Owner,
		}
		if denoPkg.Type == "github" {
			pkg.SourceCodeURL = fmt.Sprintf("https://github.com/%s/%s", denoPkg.Owner, denoPkg.Repo)
			pkg.BugTrackerURL = fmt.Sprintf("https://github.com/%s/%s/issues", denoPkg.Owner, denoPkg.Repo)
		}

		packages = append(packages, pkg)
	}

	return packages
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
