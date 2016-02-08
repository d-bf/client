package dbf

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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
	confDbf           *ConfigDbf
	pathCurrent       string
	pathConfDir       string
	pathConfFile      string
	pathVendor        string
	pathCrack         string
	regexpBenchValue  *regexp.Regexp
	regexpBenchFloat  *regexp.Regexp
	regexpBenchSuffix *regexp.Regexp
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

		// Init regexp
		regexpBenchValue, err = regexp.Compile("\\d*\\.?\\d*[A-Za-z]?")
		if err != nil {
			Log.Printf("%s\n", err)
			panic(1)
		}
		regexpBenchFloat, err = regexp.Compile("\\d*\\.?\\d*")
		if err != nil {
			Log.Printf("%s\n", err)
			panic(1)
		}
		regexpBenchSuffix, err = regexp.Compile("[A-Za-z]$")
		if err != nil {
			Log.Printf("%s\n", err)
			panic(1)
		}

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

func getBench(benchType int, platformId *string) uint64 {
	// Check vendor file
	var vendorBench, benchLinePrefix string
	switch benchType {
	case _BENCH_TYPE_CPU:
		vendorBench = "hashcat"
		benchLinePrefix = "Speed/sec:"
	case _BENCH_TYPE_GPU_AMD:
		vendorBench = "oclHashcat"
		benchLinePrefix = "Speed.GPU.#"
	case _BENCH_TYPE_GPU_NV:
		vendorBench = "cudaHashcat"
		benchLinePrefix = "Speed.GPU.#"
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
	cmd := exec.Command(pathVendorBench, "-b", "-m 0")

	cmdOut, err := cmd.StdoutPipe()
	if err != nil {
		Log.Printf("%s\n", err)
		return 0
	}

	err = cmd.Start()
	if err != nil {
		Log.Printf("%s\n", err)
		return 0
	}

	var bench, totalBench uint64
	numOfBench := 0

	cmdScanner := bufio.NewScanner(cmdOut)
	cmdScanner.Split(bufio.ScanLines)
	for cmdScanner.Scan() {
		// Get benchmark from line
		if strings.HasPrefix(cmdScanner.Text(), benchLinePrefix) {
			bench = getBenchValue(regexpBenchValue.FindString(strings.Replace(strings.SplitN(cmdScanner.Text(), ":", 2)[1], " ", "", -1)))
			if bench > 0 {
				totalBench += bench
				numOfBench++
			}
		}
	}
	err = cmdScanner.Err()
	if err != nil {
		Log.Printf("%s\n", err)
	}

	err = cmd.Wait()
	if err != nil {
		Log.Printf("%s\n", err)
	}

	if numOfBench > 0 {
		return uint64(math.Floor(float64(totalBench/uint64(numOfBench)) + 0.5))
	}

	return 0
}

func getBenchValue(benchValue string) uint64 {
	benchFloat, err := strconv.ParseFloat(regexpBenchFloat.FindString(benchValue), 64)
	if err != nil {
		Log.Printf("%s\n", err)
		return 0
	}

	benchSuffix := strings.ToUpper(regexpBenchSuffix.FindString(benchValue))

	// Should return bench in Mega

	if len(benchSuffix) == 0 {
		benchFloat /= 1048576 // Byte to Mega
	}

	switch benchSuffix {
	case "K":
		benchFloat /= 1024 // Kilo to Mega
	case "G":
		benchFloat *= 1024 // Giga to Mega
	case "T":
		benchFloat *= 1048576 // Tera to Mega
	case "P":
		benchFloat *= 1073741824 // Peta to Mega
	case "E":
		benchFloat *= 1099511627776 // Exa to Mega
	case "Z":
		benchFloat *= 1125899906842624 // Zetta to Mega
	case "Y":
		benchFloat *= 1152921504606846976 // Yotta to Mega
	}

	return uint64(math.Floor(benchFloat + 0.5))
}
