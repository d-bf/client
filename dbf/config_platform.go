package dbf

import (
	"runtime"
)

var basePlatform string

func initConfigPlatform() {
	// Set base platform
	switch runtime.GOOS { // Set OS
	case "linux":
		basePlatform = "linux"
	case "windows":
		basePlatform = "win"
	case "darwin":
		basePlatform = "mac"
	default:
		Log.Printf("The operating system '%s' is not supported!\n", runtime.GOOS)
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
		Log.Printf("The architecture '%s' is not supported in operating system '%s'!\n", runtime.GOARCH, runtime.GOOS)
		panic(1)
	}
}

func createPlatform() {
	confDbf.Platform = make([]ConfDbfPlatform, 3)

	// CPU
	confDbf.Platform[0] = ConfDbfPlatform{
		Id:        "cpu_" + basePlatform,
		Active:    1,
		Benchmark: 0,
	}

	// GPU AMD
	confDbf.Platform[1] = ConfDbfPlatform{
		Id:        "gpu_" + basePlatform + "_amd",
		Active:    0,
		Benchmark: 0,
	}

	// GPU Nvidia
	confDbf.Platform[2] = ConfDbfPlatform{
		Id:        "gpu_" + basePlatform + "_nv",
		Active:    0,
		Benchmark: 0,
	}
}
