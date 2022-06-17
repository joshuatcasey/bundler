package main

import (
	"github.com/joshuatcasey/libdependency/retrieve"
)

func main() {
	retrieve.NewMetadata("bundler", getRubyGemVersions, generateMetadata, "bionic", "jammy")
}
