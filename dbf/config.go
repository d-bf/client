package dbf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	_BENCH_TYPE_CPU     = 0
	_BENCH_TYPE_GPU_AMD = 1
	_BENCH_TYPE_GPU_NV  = 2

	_VENDOR_TYPE_GENERATOR = "generator"
	_VENDOR_TYPE_CRACKER   = "cracker"
)

var (
	confDbf      *ConfigDbf
	pathCurrent  string
	pathConfDir  string
	pathConfFile string
	pathVendor   string
	pathCrack    string
)

func InitConfig() {
	initConfigPlatform()

	// Set current path
	var err error
	pathCurrent, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err == nil {
		pathCurrent += string(os.PathSeparator)
	} else {
		Log.Printf("%s\n", err)
		panic(1)
	}

	pathConfDir = pathCurrent + "config" + string(os.PathSeparator)
	pathConfFile = pathConfDir + "dbf.json"
	pathVendor = pathCurrent + "vendor" + string(os.PathSeparator)
	pathCrack = pathCurrent + "crack" + string(os.PathSeparator)

	check()
}

// checkDir checks if dir exists and is accessible, otherwise tries to create it
func checkDir(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) { // Does not exist, so create it
			if err = os.MkdirAll(path, 0775); err != nil {
				Log.Printf("%s\n", err) // Error in creating
				return err
			}
		} else {
			Log.Printf("%s\n", err) // Error in accessing
			return err
		}
	}

	return nil
}

func check() {
	err := checkDir(pathConfDir)
	if err != nil {
		Log.Printf("%s\n", err)
		panic(1)
	}

	// Check config file
	if _, err := os.Stat(pathConfFile); err != nil {
		if os.IsNotExist(err) { // Does not exist, so create it
			// Create initial config file
			err = createConfDbf()
			if err == nil {
				fmt.Printf("Please enter server's URL in url_api in config file: %s\n", pathConfFile)
				panic(0)
			} else { // Can't create
				Log.Printf("%s\n", err)
				panic(1)
			}
		} else {
			Log.Printf("%s\n", err) // Error in accessing
			panic(1)
		}
	} else { // Sync config file
		err := checkDir(pathVendor)
		if err != nil {
			Log.Printf("%s\n", err)
			panic(1)
		}

		err = checkDir(pathCrack)
		if err != nil {
			Log.Printf("%s\n", err)
			panic(1)
		}

		confDbf = readConfDbf()

		initServer()

		// Check default vendor files and update benchmarks
		for i, platform := range *confDbf.Platform {
			if strings.HasPrefix(platform.Id, "cpu") { // CPU
				if platform.Active != 0 { // Is active
					(*confDbf.Platform)[i].Benchmark = getBench(_BENCH_TYPE_CPU, &platform.Id)
				}
			} else if strings.HasSuffix(platform.Id, "_amd") { // GPU AMD
				if platform.Active != 0 { // Is active
					(*confDbf.Platform)[i].Benchmark = getBench(_BENCH_TYPE_GPU_AMD, &platform.Id)
				}
			} else if strings.HasSuffix(platform.Id, "_nv") { // GPU Nvidia
				if platform.Active != 0 { // Is active
					(*confDbf.Platform)[i].Benchmark = getBench(_BENCH_TYPE_GPU_NV, &platform.Id)
				}
			}
		}

		// Update config file
		saveConfDbf(confDbf)
	}
}

func getBench(benchType int, platformId *string) int {
	// Check vendor file
	var vendorBench string
	switch benchType {
	case _BENCH_TYPE_CPU:
		vendorBench = "hashcat"
	case _BENCH_TYPE_GPU_AMD:
		vendorBench = "oclHashcat"
	case _BENCH_TYPE_GPU_NV:
		vendorBench = "cudaHashcat"
	default:
		return 0
	}

	pathVendorBench := pathVendor + "cracker" + string(os.PathSeparator) + vendorBench + string(os.PathSeparator) + *platformId + string(os.PathSeparator)
	err := checkDir(pathVendorBench)
	if err != nil {
		Log.Printf("%s\n", err)
		panic(1)
	}
	pathVendorBench += *platformId

	if _, err := os.Stat(pathVendorBench); os.IsNotExist(err) {
		if getVendor(_VENDOR_TYPE_CRACKER, &vendorBench, platformId, &pathVendorBench) == false {
			return 0
		}
	}

	// Preform benchmark
	return 1
}
