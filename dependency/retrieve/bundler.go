package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/joshuatcasey/libdependency/retrieve"
	"github.com/joshuatcasey/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/vacation"
)

func generateMetadata(hasVersion versionology.HasVersion) (cargo.ConfigMetadataDependency, error) {
	version := hasVersion.GetVersion().String()
	depURL := fmt.Sprintf("https://rubygems.org/downloads/bundler-%s.gem", version)

	release, ok := hasVersion.(PrettyBundlerRelease)
	if !ok {
		return cargo.ConfigMetadataDependency{}, fmt.Errorf("wrong type")
	}

	return cargo.ConfigMetadataDependency{
		Version:         version,
		ID:              "bundler",
		Name:            "Bundler",
		Source:          depURL,
		SourceSHA256:    release.SHA,
		DeprecationDate: nil,
		CPE:             fmt.Sprintf("cpe:2.3:a:bundler:bundler:%s:*:*:*:*:ruby:*:*", version),
		PURL:            retrieve.GeneratePURL("bundler", hasVersion.GetVersion().String(), release.SHA, depURL),
		Licenses:        retrieve.LookupLicenses(depURL, bundlerDecompress),
	}, nil
}

type RawBundlerRelease struct {
	Version string `json:"number"`
	Date    string `json:"created_at"`
	SHA     string `json:"sha"`
}

type PrettyBundlerRelease struct {
	SemverVersion *semver.Version
	ReleaseDate   time.Time
	SHA           string
}

func (release PrettyBundlerRelease) GetVersion() *semver.Version {
	return release.SemverVersion
}

func getRubyGemVersions() ([]versionology.HasVersion, error) {
	resp, err := http.Get("https://rubygems.org/api/v1/versions/bundler.json")
	if err != nil {
		panic("could not get release json")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("coudln't read response body")
	}

	var bundlerReleases []RawBundlerRelease
	err = json.Unmarshal(body, &bundlerReleases)
	if err != nil {
		panic("could not unmarshal response")
	}

	var rubyGemVersions []versionology.HasVersion
	for _, br := range bundlerReleases {
		var pretty PrettyBundlerRelease
		pretty.SemverVersion, err = semver.NewVersion(br.Version)
		if err != nil {
			continue
		}

		pretty.ReleaseDate, err = time.Parse(time.RFC3339Nano, br.Date)
		if err != nil {
			continue
		}

		pretty.SHA = br.SHA

		rubyGemVersions = append(rubyGemVersions, pretty)
	}

	return rubyGemVersions, nil
}

// The bundler dependency comes as a .gem file (tar.gz mime type) with a
// data.tar.gz file inside that contains the license.
func bundlerDecompress(artifact io.Reader, destination string) error {
	archive := vacation.NewArchive(artifact)
	err := archive.Decompress(destination)
	if err != nil {
		return fmt.Errorf("failed to decompress source file: %w", err)
	}

	innerArtifact, _ := os.Open(filepath.Join(destination, "data.tar.gz"))
	innerArchive := vacation.NewArchive(innerArtifact)
	err = innerArchive.Decompress(destination)
	if err != nil {
		return fmt.Errorf("failed to decompress inner source file: %w", err)
	}

	return nil
}
