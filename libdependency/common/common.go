package common

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"golang.org/x/exp/slices"
)

func ParseBuildpackToml(buildpackTomlPath string) cargo.Config {
	configParser := cargo.NewBuildpackParser()
	config, err := configParser.Parse(buildpackTomlPath)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %s: %s", buildpackTomlPath, err))
	}
	return config
}

type RetrievalOutput struct {
	Versions []string
	ID       string
	Name     string
}

func GetNewVersions(id, name string, config cargo.Config, allVersions []*semver.Version, output string) {
	buildpackVersions := GetBuildpackVersions(id, config)
	versionsFilteredByConstraints := FilterToConstraints(id, config, allVersions)
	versionsFilteredByPatches := FilterToPatches(versionsFilteredByConstraints, config, buildpackVersions)

	if len(versionsFilteredByPatches) < 1 {
		panic("No versions found")
	}

	retrievalOutput := RetrievalOutput{
		Versions: versionsFilteredByPatches,
		ID:       id,
		Name:     name,
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

func GetBuildpackVersions(id string, config cargo.Config) []string {
	var buildpackVersions []string
	for _, d := range config.Metadata.Dependencies {
		if d.ID != id {
			continue
		}
		buildpackVersions = append(buildpackVersions, d.Version)
	}
	return buildpackVersions
}

func FilterToPatches(versionsFilteredByConstraints map[string][]*semver.Version, config cargo.Config, buildpackVersions []string) []string {
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

func FilterToConstraints(id string, config cargo.Config, allVersions []*semver.Version) map[string][]*semver.Version {
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
	for _, version := range allVersions {
		for constraintId, constraint := range semverConstraints {
			if constraint.Check(version) {
				newVersions[constraintId] = append(newVersions[constraintId], version)
			}
		}
	}
	return newVersions
}
