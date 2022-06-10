package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

func main() {
	buildpackTomlPath := filepath.Join("..", "..", "buildpack.toml")
	config := parseBuildpackToml(buildpackTomlPath)

	metadataPath := os.Args[1]
	tarballPath := os.Args[2]
	id := os.Args[3]
	version := os.Args[4]
	targetOS := os.Args[5]

	metadataContents, err := os.ReadFile(metadataPath)
	if err != nil {
		panic(err)
	}

	tarballChecksum, err := fs.NewChecksumCalculator().Sum(tarballPath)
	if err != nil {
		panic(err)
	}

	var metadata cargo.ConfigMetadataDependency
	err = json.Unmarshal(metadataContents, &metadata)
	if err != nil {
		panic(fmt.Errorf("failed to parse metadata file: %w", err))
	}

	metadata.ID = id
	metadata.URI = fmt.Sprintf("%s-%s-%s.tgz", id, version, targetOS)
	metadata.SHA256 = tarballChecksum
	metadata.Stacks = []string{targetOS}
	config.Metadata.Dependencies = append(config.Metadata.Dependencies, metadata)

	file, err := os.OpenFile(buildpackTomlPath, os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		panic(fmt.Errorf("failed to open buildpack config file: %w", err))
	}
	defer file.Close()

	err = cargo.EncodeConfig(file, config)
	if err != nil {
		panic(fmt.Errorf("failed to write buildpack config: %w", err))
	}

	fmt.Println("Updating buildpack.toml with new version: ", metadata.Version)
}

func parseBuildpackToml(buildpackTomlPath string) cargo.Config {
	configParser := cargo.NewBuildpackParser()
	config, err := configParser.Parse(buildpackTomlPath)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %s: %s", buildpackTomlPath, err))
	}
	return config
}
