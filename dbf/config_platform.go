package dbf

import (
	"runtime"
)

var (
	basePlatform  string
	extExecutable string
)

func initConfigPlatform() {
	// Set base platform
	switch runtime.GOOS { // Set OS
	case "linux":
		basePlatform = "linux"
		extExecutable = ".bin"
	case "windows":
		basePlatform = "win"
		extExecutable = ".exe"
	case "darwin":
		basePlatform = "mac"
		extExecutable = ".app"
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
	confDbf.Platform = make([]StructConfDbfPlatform, 3)

	// CPU
	confDbf.Platform[0] = StructConfDbfPlatform{
		Id:        "cpu_" + basePlatform,
		Active:    1,
		Benchmark: 0,
	}

	// GPU AMD
	confDbf.Platform[1] = StructConfDbfPlatform{
		Id:        "gpu_" + basePlatform + "_amd",
		Active:    0,
		Benchmark: 0,
	}

	// GPU Nvidia
	confDbf.Platform[2] = StructConfDbfPlatform{
		Id:        "gpu_" + basePlatform + "_nv",
		Active:    0,
		Benchmark: 0,
	}
}
