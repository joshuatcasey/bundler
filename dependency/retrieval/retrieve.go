package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/joshuatcasey/bundler/libdependency/common"
)

type BundlerRelease struct {
	Version string `json:"number"`
	Date    string `json:"created_at"`
	SHA     string `json:"sha"`
}

func main() {
	buildpackTomlPath := os.Args[1]
	output := os.Args[2]

	fmt.Printf("buildpackTomlPath=%s\n", buildpackTomlPath)
	fmt.Printf("output=%s\n", output)

	var id = "bundler"
	config := common.ParseBuildpackToml(buildpackTomlPath)

	rubyGemVersions := getRubyGemVersions()
	common.GetNewVersions(id, "Bundler", config, rubyGemVersions, output)
}

func getRubyGemVersions() []*semver.Version {
	resp, err := http.Get("https://rubygems.org/api/v1/versions/bundler.json")
	if err != nil {
		panic("could not get release json")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("coudln't read response body")
	}

	var bundlerReleases []BundlerRelease
	err = json.Unmarshal(body, &bundlerReleases)
	if err != nil {
		panic("could not unmarshal response")
	}

	var rubyGemVersions []*semver.Version
	for _, br := range bundlerReleases {
		version, err := semver.NewVersion(br.Version)
		if err != nil {
			continue
		}
		rubyGemVersions = append(rubyGemVersions, version)
	}

	return rubyGemVersions
}
