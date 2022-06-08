package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type BundlerRelease struct {
	Version string `json:"number"`
}

func main() {
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

	var versions []string
	for _, br := range bundlerReleases {
		versions = append(versions, br.Version)
	}

	fmt.Println(versions)
}
