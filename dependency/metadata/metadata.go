package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"github.com/package-url/packageurl-go"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/vacation"
)

type DepVersion struct {
	cargo.ConfigMetadataDependency
	ReleaseDate *time.Time `json:"release_date,omitempty"`
}

func main() {
	version := os.Args[1]
	output := os.Args[2]

	fmt.Printf("version=%s\n", version)
	fmt.Printf("output=%s\n", output)

	dependencyVersion := getDependencyVersion(version)
	bytes, err := json.Marshal(dependencyVersion)
	if err != nil {
		panic(fmt.Errorf("cannot marshal: %w", err))
	}

	err = os.WriteFile(output, bytes, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("cannot write to %s: %w", output, err))
	}

	fmt.Println(string(bytes))
}

func getDependencyVersion(version string) DepVersion {
	bundlerReleases := getRubyGemVersions()

	depURL := fmt.Sprintf("https://rubygems.org/downloads/bundler-%s.gem", version)
	for _, release := range bundlerReleases {
		if release.Version.String() == version {
			return DepVersion{
				ConfigMetadataDependency: cargo.ConfigMetadataDependency{
					Version:         version,
					Source:          depURL,
					SourceSHA256:    release.SHA,
					DeprecationDate: nil,
					CPE:             fmt.Sprintf("cpe:2.3:a:bundler:bundler:%s:*:*:*:*:ruby:*:*", version),
					PURL:            generatePURL("bundler", version, release.SHA, depURL),
					Licenses:        lookupLicenses(depURL),
				},
				ReleaseDate: &release.ReleaseDate,
			}
		}
	}

	return DepVersion{}
}

type RawBundlerRelease struct {
	Version string `json:"number"`
	Date    string `json:"created_at"`
	SHA     string `json:"sha"`
}

type PrettyBundlerRelease struct {
	Version     *semver.Version
	ReleaseDate time.Time
	SHA         string
}

func getRubyGemVersions() []PrettyBundlerRelease {
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

	var rubyGemVersions []PrettyBundlerRelease
	for _, br := range bundlerReleases {
		var pretty PrettyBundlerRelease
		pretty.Version, err = semver.NewVersion(br.Version)
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

	return rubyGemVersions
}

func lookupLicenses(sourceURL string) []interface{} {
	// getting the dependency artifact from sourceURL
	resp, err := http.Get(sourceURL)
	if err != nil {
		panic(fmt.Errorf("failed to query url: %w", err))
	}
	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf("failed to query url %s with: status code %d", sourceURL, resp.StatusCode))
	}

	// decompressing the dependency artifact
	tempDir, err := os.MkdirTemp("", "destination")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	err = bundlerDecompress(resp.Body, tempDir)
	if err != nil {
		panic(err)
	}

	// scanning artifact for license file
	filer, err := filer.FromDirectory(tempDir)
	if err != nil {
		panic(fmt.Errorf("failed to setup a licensedb filer: %w", err))
	}

	licenses, err := licensedb.Detect(filer)
	// if no licenses are found, just return an empty slice.
	if err != nil {
		if err.Error() != "no license file was found" {
			panic(fmt.Errorf("failed to detect licenses: %w", err))
		}
		return []interface{}{}
	}

	// Only return the license IDs, in alphabetical order
	var licenseIDs []string
	for key := range licenses {
		licenseIDs = append(licenseIDs, key)
	}
	sort.Strings(licenseIDs)

	var licenseIDsAsInterface []interface{}
	for _, licenseID := range licenseIDs {
		licenseIDsAsInterface = append(licenseIDsAsInterface, licenseID)
	}

	return licenseIDsAsInterface
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

func generatePURL(name, version, sourceSHA, source string) string {
	purl := packageurl.NewPackageURL(
		packageurl.TypeGeneric,
		"",
		name,
		version,
		packageurl.QualifiersFromMap(map[string]string{
			"checksum":     sourceSHA,
			"download_url": source,
		}),
		"",
	)

	// Unescape the path to remove the added `%2F` and other encodings added to
	// the URL by packageurl-go
	// If the unescaping fails, we should still return the path URL with the
	// encodings, packageurl-go has examples with both the encodings and without,
	// we prefer to avoid the encodings when we can for convenience.
	purlString, err := url.PathUnescape(purl.ToString())
	if err != nil {
		return purl.ToString()
	}

	return purlString
}
