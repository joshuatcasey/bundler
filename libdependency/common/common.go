package common

import (
	"fmt"
	"time"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type DepVersion struct {
	cargo.ConfigMetadataDependency
	ReleaseDate *time.Time `json:"release_date,omitempty"`
}

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
