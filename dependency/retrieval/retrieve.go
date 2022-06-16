package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/joshuatcasey/bundler/libdependency/common"
	"github.com/paketo-buildpacks/packit/v2/cargo"
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

	retrievalOutput := common.RetrievalOutput{
		Versions: versionsFilteredByPatches,
		ID:       id,
		Name:     "Bundler",
	}

	bytes, err := json.Marshal(retrievalOutput)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(output, bytes, os.ModePerm)
	if err != nil {
		panic(err)
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
