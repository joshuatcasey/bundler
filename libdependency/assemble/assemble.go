package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/fs"
)

type Artifact struct {
	tarballPath   string
	uri           string
	tarballSHA256 string
	os            string
	version       string
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

	metadataGlob := filepath.Join(artifactPath, "metadata-*.json")
	if metadataFiles, err := filepath.Glob(metadataGlob); err != nil {
		panic(err)
	} else if len(metadataFiles) < 1 {
		panic(fmt.Errorf("no metadata files found: %s", metadataGlob))
	} else {
		fmt.Printf("Found metadata files:\n")
		for _, metadata := range metadataFiles {
			fmt.Printf("- %s\n", filepath.Base(metadata))
		}
	}

	artifacts := findArtifacts(artifactPath, id)

	bytes, err := json.Marshal(artifacts)
	if err != nil {
		panic("cannot marshal")
	}
	fmt.Println("Found artifacts:")
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
				uri: "<UNKNOWN>",
			}
			for _, file := range files {
				fullpath := filepath.Join(tarball, file.Name())
				fmt.Printf("  - %s\n", file.Name())

				if isTarball(file) {
					artifact.tarballPath = fullpath
					continue
				}

				bytes, err := os.ReadFile(fullpath)
				if err != nil {
					panic(err)
				}

				if isOS(file) {
					artifact.os = string(bytes)
				} else if isVersion(file) {
					artifact.version = string(bytes)
				} else if isSHA256(file) {
					artifact.tarballSHA256 = string(bytes)
				}
			}

			calculatedSHA256, err := fs.NewChecksumCalculator().Sum(artifact.tarballPath)
			calculatedSHA256 = calculatedSHA256[:64]
			if !strings.HasPrefix(artifact.tarballSHA256, calculatedSHA256) {
				fmt.Printf("SHA256 does not match! Expected=%s, Calculated=%s\n", artifact.tarballSHA256, calculatedSHA256)
				panic("SHA256 does not match!")
			}
			artifact.tarballSHA256 = calculatedSHA256

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
