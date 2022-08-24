package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/joshuatcasey/bundler/libdependency/common"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"golang.org/x/exp/slices"
)

// Copy of cargo config structure, with the addition of the Target field
type Dependency struct {
	cargo.ConfigMetadataDependency
	Target string `toml:"target"          json:"target,omitempty"`
}

// Dependency-specific function that returns all dependencies from upstream as []semver.Version
type UpstreamVersionFunc func() []*semver.Version

// Dependency-specific function that given a version string, generates the
// metadata for the dependency according to
// https://github.com/paketo-buildpacks/rfcs/blob/main/text/dependencies/rfcs/0004-dependency-management-phase-one.md#1-version-retrieval-make-retrieve
// This may output multiple dependencies for one version, that contain different `target` fields
type GenerateMetadataFunc func(version string) []Dependency

func Retrieve(id, buildpackTomlPath, output string, upstreamVersionFunc UpstreamVersionFunc, generateMetadataFunc GenerateMetadataFunc) {
	newVersions := upstreamVersionFunc()
	config := common.ParseBuildpackToml(buildpackTomlPath)
	buildpackVersions := GetBuildpackVersions(id, config)
	versionsFilteredByConstraints := FilterToConstraints(id, config, newVersions)
	versionsFilteredByPatches := FilterToPatches(versionsFilteredByConstraints, config, buildpackVersions)

	if len(versionsFilteredByPatches) < 1 {
		panic("No versions found")
	}

	fmt.Printf("Generating metadata for %v\n", versionsFilteredByPatches)
	allDependencies := []Dependency{}

	for _, version := range versionsFilteredByPatches {
		dependencyVersions := generateMetadataFunc(version)
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
	fmt.Println("done!")
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

func FilterToConstraints(id string, config cargo.Config, versions []*semver.Version) map[string][]*semver.Version {
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
	for _, version := range versions {
		for constraintID, constraint := range semverConstraints {
			if constraint.Check(version) {
				newVersions[constraintID] = append(newVersions[constraintID], version)
			}
		}
	}
	return newVersions
}

func FilterToPatches(versions map[string][]*semver.Version, config cargo.Config, buildpackVersions []string) []string {
	var versionsToAdd []*semver.Version
	for constraintID, versions := range versions {
		var buildpackConstraint cargo.ConfigMetadataDependencyConstraint
		for _, constraint := range config.Metadata.DependencyConstraints {
			if constraint.ID == constraintID {
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
