package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit/v2/fs"
)

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

			for _, file := range files {
				fmt.Printf("  - %s, isDir=%t, isTarball=%t, isSHA256=%t\n",
					file.Name(),
					file.IsDir(),
					isTarball(file),
					isSHA256(file))
			}
		}
	}
}

func isTarball(file os.FileInfo) bool {
	return strings.HasSuffix(file.Name(), ".tgz")
}

func isSHA256(file os.FileInfo) bool {
	return strings.HasSuffix(file.Name(), ".sha256")
}
