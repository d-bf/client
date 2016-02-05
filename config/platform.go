package config

import (
	"github.com/d-bf/client/dbf"
	"runtime"
)

var basePlatform string

func init() {
	// Set base platform
	switch runtime.GOOS { // Set OS
	case "linux":
		basePlatform = "linux"
	case "windows":
		basePlatform = "win"
	case "darwin":
		basePlatform = "mac"
	default:
		dbf.Log.Printf("The operating system '%s' is not supported!\n", runtime.GOOS)
		panic(1)
	}
	switch runtime.GOARCH { // Set Arch
	case "386":
		basePlatform += "_32"
	case "amd64":
		basePlatform += "_64"
		//	case "arm":
		//		basePlatform += "_arm"
		//	case "arm64":
		//		basePlatform += "_arm64"
	default:
		dbf.Log.Printf("The architecture '%s' is not supported in operating system '%s'!\n", runtime.GOARCH, runtime.GOOS)
		panic(1)
	}
}

/*
type dbfConfPlatform struct {
	Id        string `json:"id"`
	Active    int    `json:"active"`
	Benchmark int    `json:"benchmark"`
}
*/

func createPlatforms() *[]DbfConfPlatform {
	platform := make([]DbfConfPlatform, 3)

	// CPU
	platform[0] = DbfConfPlatform{
		Id:        "cpu_" + basePlatform,
		Active:    1,
		Benchmark: 0,
	}

	// GPU AMD
	platform[1] = DbfConfPlatform{
		Id:        "gpu_" + basePlatform + "_amd",
		Active:    0,
		Benchmark: 0,
	}

	// GPU Nvidia
	platform[2] = DbfConfPlatform{
		Id:        "gpu_" + basePlatform + "_nv",
		Active:    0,
		Benchmark: 0,
	}

	return &platform
}
