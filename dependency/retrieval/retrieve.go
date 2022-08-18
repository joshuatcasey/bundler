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
	"github.com/anchore/packageurl-go"
	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"github.com/joshuatcasey/bundler/libdependency/common"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/vacation"
	"golang.org/x/exp/slices"
)

type BundlerRelease struct {
	Version string `json:"number"`
	Date    string `json:"created_at"`
	SHA     string `json:"sha"`
}

var id = "bundler"

func main() {
	buildpackTomlPath := os.Args[1]
	output := os.Args[2]

	fmt.Printf("buildpackTomlPath=%s\n", buildpackTomlPath)
	fmt.Printf("output=%s\n", output)

	config := common.ParseBuildpackToml(buildpackTomlPath)

	buildpackVersions := getBuildpackVersions(config)
	rubyGemVersions := getRubyGemVersions()
	versionsFilteredByConstraints := filterToConstraints(config, rubyGemVersions)
	versionsFilteredByPatches := filterToPatches(versionsFilteredByConstraints, config, buildpackVersions)

	if len(versionsFilteredByPatches) < 1 {
		panic("No versions found")
	}

	fmt.Printf("generating metadata for %v", versionsFilteredByPatches)
	allDependencies := []cargo.ConfigMetadataDependency{}
	for _, version := range versionsFilteredByPatches {
		dependencyVersions := getDependencyVersion(version)
		allDependencies = append(allDependencies, dependencyVersions...)
	}

	bytes, err := json.Marshal(allDependencies)
	if err != nil {
		panic(fmt.Errorf("cannot marshal: %w", err))
	}
	err = os.WriteFile(output, bytes, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("cannot write to %s: %w", output, err))
	}

	fmt.Println(string(bytes))
}

func filterToPatches(versionsFilteredByConstraints map[string][]*semver.Version, config cargo.Config, buildpackVersions []string) []string {
	var versionsToAdd []*semver.Version
	for constraintId, versions := range versionsFilteredByConstraints {
		var buildpackConstraint cargo.ConfigMetadataDependencyConstraint
		for _, constraint := range config.Metadata.DependencyConstraints {
			if constraint.ID == constraintId {
				buildpackConstraint = constraint
			}
		}

		sort.Slice(versions, func(i, j int) bool {
			return versions[i].LessThan(versions[j])
		})

		// if there are more requested patches than matching dependencies, just
		// return all matching dependencies.
		if buildpackConstraint.Patches > len(versions) {
			continue
		}

		// Buildpack.toml dependencies are usually in order from lowest to highest
		// version. We want to return the the n largest matching dependencies in the
		// same order, n being the constraint.Patches field from the buildpack.toml.
		// Here, we are returning the n highest matching Dependencies.
		versionsToAdd = append(versionsToAdd, versions[len(versions)-buildpackConstraint.Patches:]...)
	}

	var versionsAsStrings []string
	for _, version := range versionsToAdd {
		versionAsString := version.String()
		if !slices.Contains(buildpackVersions, versionAsString) {
			versionsAsStrings = append(versionsAsStrings, versionAsString)
		}
	}

	return versionsAsStrings
}

func filterToConstraints(config cargo.Config, rubyGemVersions []*semver.Version) map[string][]*semver.Version {
	semverConstraints := make(map[string]*semver.Constraints)
	for _, constraint := range config.Metadata.DependencyConstraints {
		if constraint.ID != id {
			continue
		}

		semverConstraint, err := semver.NewConstraint(constraint.Constraint)
		if err != nil {
			panic(err)
		}
		semverConstraints[constraint.ID] = semverConstraint
	}

	newVersions := make(map[string][]*semver.Version)
	for _, version := range rubyGemVersions {
		for constraintId, constraint := range semverConstraints {
			if constraint.Check(version) {
				newVersions[constraintId] = append(newVersions[constraintId], version)
			}
		}
	}
	return newVersions
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

func getBuildpackVersions(config cargo.Config) []string {
	var buildpackVersions []string
	for _, d := range config.Metadata.Dependencies {
		if d.ID != id {
			continue
		}
		buildpackVersions = append(buildpackVersions, d.Version)
	}
	return buildpackVersions
}

func getDependencyVersion(version string) []cargo.ConfigMetadataDependency {
	bundlerReleases := getPrettyRubyGemVersions()
	targets := map[string][]string{"bionic": []string{"io.buildpacks.stacks.bionic"}, "jammy": []string{"io.buildpacks.stacks.jammy"}}
	dependencies := []cargo.ConfigMetadataDependency{}

	depURL := fmt.Sprintf("https://rubygems.org/downloads/bundler-%s.gem", version)
	for _, release := range bundlerReleases {
		if release.Version.String() == version {
			for target, stacks := range targets {
				dependencies = append(dependencies,
					cargo.ConfigMetadataDependency{
						Version:         version,
						ID:              "bundler",
						Name:            "Bundler",
						Source:          depURL,
						SourceSHA256:    release.SHA,
						DeprecationDate: nil,
						CPE:             fmt.Sprintf("cpe:2.3:a:bundler:bundler:%s:*:*:*:*:ruby:*:*", version),
						PURL:            generatePURL("bundler", version, release.SHA, depURL),
						Licenses:        lookupLicenses(depURL),
						Stacks:          stacks,
						Target:          target,
					})
			}
		}
	}
	return dependencies
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

func getPrettyRubyGemVersions() []PrettyBundlerRelease {
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
