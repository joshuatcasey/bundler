package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joshuatcasey/bundler/libdependency/common"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

type Artifact struct {
	TarballPath   string
	Uri           string
	TarballSHA256 string
	Os            string
	Version       string
	Metadata      *common.DepVersion
}

func main() {
	id := os.Args[1]
	artifactPath := os.Args[2]

	fmt.Printf("id=%s\n", id)
	fmt.Printf("artifactPath=%s\n", artifactPath)

	if exists, err := fs.Exists(artifactPath); err != nil {
		panic(err)
	} else if !exists {
		panic(fmt.Errorf("directory %s does not exist", artifactPath))
	} else if fs.IsEmptyDir(artifactPath) {
		panic(fmt.Errorf("directory %s is empty", artifactPath))
	}

	versionToMetadata := getMetadata(artifactPath)
	fmt.Println("Found metadata:")
	printAsJson(versionToMetadata)

	artifacts := findArtifacts(artifactPath, id)

	for i := range artifacts {
		artifacts[i].Metadata = versionToMetadata[artifacts[i].Version]
	}

	fmt.Println("Found artifacts:")
	printAsJson(artifacts)
}

func getMetadata(artifactPath string) map[string]*common.DepVersion {
	versionToMetadata := make(map[string]*common.DepVersion)
	metadataGlob := filepath.Join(artifactPath, "metadata-*.json")
	if metadataFiles, err := filepath.Glob(metadataGlob); err != nil {
		panic(err)
	} else if len(metadataFiles) < 1 {
		panic(fmt.Errorf("no metadata files found: %s", metadataGlob))
	} else {
		fmt.Printf("Found metadata files:\n")
		for _, metadata := range metadataFiles {
			fmt.Printf("- %s\n", filepath.Base(metadata))

			version := strings.TrimPrefix(filepath.Base(metadata), "metadata-")
			version = strings.TrimSuffix(version, ".json")

			var depVersion common.DepVersion

			metadataContents, err := os.ReadFile(filepath.Join(metadata, filepath.Base(metadata)))
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal(metadataContents, &depVersion)
			if err != nil {
				panic(fmt.Errorf("failed to parse metadata file: %w", err))
			}

			versionToMetadata[version] = &depVersion
		}
	}
	return versionToMetadata
}

func printAsJson(item interface{}) {
	bytes, err := json.Marshal(item)
	if err != nil {
		panic("cannot marshal")
	}
	fmt.Println(string(bytes))
}

func findArtifacts(artifactPath string, id string) []Artifact {
	var artifacts []Artifact

	tarballGlob := filepath.Join(artifactPath, fmt.Sprintf("%s-*", id))
	if tarballs, err := filepath.Glob(tarballGlob); err != nil {
		panic(err)
	} else if len(tarballs) < 1 {
		panic(fmt.Errorf("no compiled artifact folders found: %s", tarballGlob))
	} else {
		fmt.Printf("Found compiled artifact folders:\n")
		for _, tarball := range tarballs {
			fmt.Printf("- %s\n", filepath.Base(tarball))

			dir, err := os.Open(tarball)
			if err != nil {
				panic(err)
			}

			files, err := dir.Readdir(0)
			if err != nil {
				panic(err)
			}

			artifact := Artifact{
				Uri: "<UNKNOWN>",
			}
			for _, file := range files {
				fullpath := filepath.Join(tarball, file.Name())
				fmt.Printf("  - %s\n", file.Name())

				if isTarball(file) {
					artifact.TarballPath = fullpath
					continue
				}

				bytes, err := os.ReadFile(fullpath)
				if err != nil {
					panic(err)
				}

				if isOS(file) {
					artifact.Os = strings.TrimSpace(string(bytes))
				} else if isVersion(file) {
					artifact.Version = strings.TrimSpace(string(bytes))
				} else if isSHA256(file) {
					artifact.TarballSHA256 = strings.TrimSpace(string(bytes))
				}
			}

			calculatedSHA256, err := fs.NewChecksumCalculator().Sum(artifact.TarballPath)
			calculatedSHA256 = calculatedSHA256[:64]
			if !strings.HasPrefix(artifact.TarballSHA256, calculatedSHA256) {
				fmt.Printf("SHA256 does not match! Expected=%s, Calculated=%s\n", artifact.TarballSHA256, calculatedSHA256)
				panic("SHA256 does not match!")
			}
			artifact.TarballSHA256 = calculatedSHA256

			if err != nil {
				panic(err)
			}

			artifacts = append(artifacts, artifact)
		}
	}
	return artifacts
}

func isTarball(file os.FileInfo) bool {
	return strings.HasSuffix(file.Name(), ".tgz")
}

func isSHA256(file os.FileInfo) bool {
	return strings.HasSuffix(file.Name(), ".sha256")
}

func isOS(file os.FileInfo) bool {
	return strings.HasSuffix(file.Name(), ".os")
}

func isVersion(file os.FileInfo) bool {
	return strings.HasSuffix(file.Name(), ".version")
}
