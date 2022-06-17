package main

import (
	"github.com/joshuatcasey/libdependency/retrieve"
)

func main() {
	// To debug, uncomment these lines
	//os.Args = []string{
	//	"foobar",
	//	"--buildpack-toml-path", "/Users/caseyj/git/paketo-buildpacks/bundler/buildpack.toml",
	//	"--output-file", "/var/folders/r4/8zvzyhwj405_393mznhtxb9r0000gn/T/tmp.hU7kMOQp",
	//	"--id", "bundler",
	//	"--name", "Bundler",
	//}
	retrieve.NewMetadata(getRubyGemVersions, generateMetadata)
}
