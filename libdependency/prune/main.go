package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"golang.org/x/exp/slices"
)

func main() {
	config := parseBuildpackToml()

	// Get a map from constraints to dependencies
	constraintToDependencies := make(map[cargo.ConfigMetadataDependencyConstraint][]cargo.ConfigMetadataDependency)

	for _, dependency := range config.Metadata.Dependencies {
		dependencyVersionAsSemver := semver.MustParse(dependency.Version)
		for _, constraint := range config.Metadata.DependencyConstraints {
			constraintAsSemver, err := semver.NewConstraint(constraint.Constraint)
			if err != nil {
				panic(err)
			}

			if dependency.ID == constraint.ID && constraintAsSemver.Check(dependencyVersionAsSemver) {
				constraintToDependencies[constraint] = append(constraintToDependencies[constraint], dependency)
			}
		}
	}

	constraintToPatches := make(map[cargo.ConfigMetadataDependencyConstraint][]string)

	// We can have more than one dependency with the same version
	// so we have to figure out which versions are captured in the patches
	for constraint, dependencies := range constraintToDependencies {
		for _, dependency := range dependencies {
			if !slices.Contains(constraintToPatches[constraint], dependency.Version) {
				constraintToPatches[constraint] = append(constraintToPatches[constraint], dependency.Version)
			}
		}

		sort.Slice(constraintToPatches[constraint], func(i, j int) bool {
			iVersion := semver.MustParse(constraintToPatches[constraint][i])
			jVersion := semver.MustParse(constraintToPatches[constraint][j])
			return iVersion.LessThan(jVersion)
		})

		if constraint.Patches < len(constraintToPatches[constraint]) {
			constraintToPatches[constraint] = constraintToPatches[constraint][len(constraintToPatches[constraint])-constraint.Patches:]
		}
	}

	var patchesToKeep []string
	for _, versions := range constraintToPatches {
		patchesToKeep = append(patchesToKeep, versions...)
	}

	fmt.Println("patchesToKeep")
	fmt.Println(patchesToKeep)

	var dependenciesToKeep []cargo.ConfigMetadataDependency

	for _, dependency := range config.Metadata.Dependencies {
		if slices.Contains(patchesToKeep, dependency.Version) {
			dependenciesToKeep = append(dependenciesToKeep, dependency)
		}
	}

	// Sort the stacks within the dependency
	for _, dependency := range dependenciesToKeep {
		sort.Slice(dependency.Stacks, func(i, j int) bool {
			return dependency.Stacks[i] > dependency.Stacks[j]
		})
	}

	// Sort the dependencies by:
	// 1. ID
	// 2. Version
	// 3. len(Stacks)
	sort.Slice(dependenciesToKeep, func(i, j int) bool {
		dep1 := dependenciesToKeep[i]
		dep2 := dependenciesToKeep[j]
		if dep1.ID == dep2.ID {
			dep1Version := semver.MustParse(dep1.Version)
			dep2Version := semver.MustParse(dep2.Version)

			if dep1Version.Equal(dep2Version) {
				return len(dep1.Stacks) > len(dep2.Stacks)
			}

			return dep1Version.GreaterThan(dep2Version)
		}
		return dep1.ID > dep2.ID
	})

	fmt.Println("dependenciesToKeep")
	fmt.Println(dependenciesToKeep)

	config.Metadata.Dependencies = dependenciesToKeep

	file, err := os.OpenFile(filepath.Join("..", "..", "buildpack.toml"), os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		panic(fmt.Errorf("failed to open buildpack config file: %w", err))
	}
	defer file.Close()

	err = cargo.EncodeConfig(file, config)
	if err != nil {
		panic(fmt.Errorf("failed to write buildpack config: %w", err))
	}
}

func parseBuildpackToml() cargo.Config {
	buildpackTomlPath := filepath.Join("..", "..", "buildpack.toml")

	configParser := cargo.NewBuildpackParser()
	config, err := configParser.Parse(buildpackTomlPath)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %s: %s", buildpackTomlPath, err))
	}
	return config
}
