package common

import (
	"time"

	"github.com/paketo-buildpacks/packit/v2/cargo"
)

type DepVersion struct {
	cargo.ConfigMetadataDependency
	ReleaseDate *time.Time `json:"release_date,omitempty"`
}
