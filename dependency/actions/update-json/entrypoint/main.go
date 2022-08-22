package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

func main() {
	var config struct {
		Version string
		Target  string
		SHA256  string
		URI     string
		File    string
	}

	flag.StringVar(&config.Version, "version", "", "Dependency version")
	flag.StringVar(&config.Target, "target", "", "Dependency target name")
	flag.StringVar(&config.SHA256, "sha256", "", "Dependency SHA256 to add")
	flag.StringVar(&config.URI, "uri", "", "Dependency URI to add")
	flag.StringVar(&config.File, "file", "", "Dependency metadata.json file to modify")
	flag.Parse()

	if config.Version == "" {
		fail(errors.New(`missing required input "version"`))
	}

	if config.Target == "" {
		fail(errors.New(`missing required input "target"`))
	}
	if config.SHA256 == "" {
		fail(errors.New(`missing required input "SHA256"`))
	}
	if config.URI == "" {
		fail(errors.New(`missing required input "uri"`))
	}
	if config.File == "" {
		fail(errors.New(`missing required input "file"`))
	}

	entries := []*cargo.ConfigMetadataDependency{}
	file, err := os.OpenFile(config.File, os.O_RDWR, os.ModePerm)
	if err != nil {
		fail(err)
	}

	err = json.NewDecoder(file).Decode(&entries)
	if err != nil {
		fail(err)
	}

	// Find the dependency of interest and update the SHA256
	for _, dependency := range entries {
		if dependency.Target == config.Target && dependency.Version == config.Version {
			dependency.SHA256 = config.SHA256
			dependency.URI = config.URI
		}
	}

	_ = file.Truncate(0)
	_, _ = file.Seek(0, 0)

	// Write it back to the file
	err = json.NewEncoder(file).Encode(entries)
	if err != nil {
		fail(err)
	}
	defer file.Close()

	fmt.Println("Success!")
}

func fail(err error) {
	fmt.Printf("Error: %s", err)
	os.Exit(1)
}
