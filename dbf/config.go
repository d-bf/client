package dbf

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	PATH_SEPARATOR = string(os.PathSeparator)

	_PATH_CONF_DIR  = 1
	_PATH_CONF_FILE = 2
	_PATH_VENDOR    = 3
	_PATH_TASK      = 4
	_PATH_CRACK     = 5

	_BENCH_TYPE_CPU     = 0
	_BENCH_TYPE_GPU_AMD = 1
	_BENCH_TYPE_GPU_NV  = 2

	_VENDOR_TYPE_GENERATOR = "generator"
	_VENDOR_TYPE_CRACKER   = "cracker"
)

var (
	PathData          string
	confDbf           StructConfDbf
	activePlatStr     string
	regexpBenchValue  *regexp.Regexp
	regexpBenchFloat  *regexp.Regexp
	regexpBenchSuffix *regexp.Regexp
	wgBench           sync.WaitGroup
)

type StructActivePlatform struct {
	Id        string `json:"id"`
	Benchmark uint64 `json:"benchmark"`
}

func InitConfig() {
	initConfigPlatform()

	check()
}

func getPath(path int) string {
	switch path {
	case _PATH_CONF_DIR:
		return PathData + "config" + PATH_SEPARATOR
	case _PATH_CONF_FILE:
		return PathData + "config" + PATH_SEPARATOR + "dbf.json"
	case _PATH_CRACK:
		return PathData + "crack" + PATH_SEPARATOR
	case _PATH_TASK:
		return PathData + "task" + PATH_SEPARATOR
	case _PATH_VENDOR:
		return PathData + "vendor" + PATH_SEPARATOR
	default:
		Log.Printf("Undefined path id '%d' in getPath()!\n", path)
		return ""
	}
}

func check() {
	err := checkDir(getPath(_PATH_CONF_DIR))
	if err != nil {
		Log.Printf("%s\n", err)
		panic(1)
	}

	// Check config file
	if _, err := os.Stat(getPath(_PATH_CONF_FILE)); err != nil {
		if os.IsNotExist(err) { // Does not exist, so create it

			fmt.Println("Creating initial config file...")

			// Create initial config file
			err = createConfDbf()
			if err == nil {
				fmt.Printf("Please enter server's URL in url_api in config file: %s\n", getPath(_PATH_CONF_FILE))
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
		fmt.Println("Synchronizing config file...")

		err := checkDir(getPath(_PATH_VENDOR))
		if err != nil {
			Log.Printf("%s\n", err)
			panic(1)
		}

		err = checkDir(getPath(_PATH_TASK))
		if err != nil {
			Log.Printf("%s\n", err)
			panic(1)
		}

		err = checkDir(getPath(_PATH_CRACK))
		if err != nil {
			Log.Printf("%s\n", err)
			panic(1)
		}

		setConfDbf()

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
		var activePlat []StructActivePlatform
		for i, platform := range confDbf.Platform {
			if platform.Active != 0 { // Is active
				if strings.HasPrefix(platform.Id, "cpu") { // CPU
					wgBench.Add(1)
					go func(i int, platformId string) {
						confDbf.Platform[i].Benchmark = getBench(_BENCH_TYPE_CPU, &platformId)
						wgBench.Done()
					}(i, platform.Id)
				} else if strings.HasSuffix(platform.Id, "_amd") { // GPU AMD
					wgBench.Add(1)
					go func(i int, platformId string) {
						confDbf.Platform[i].Benchmark = getBench(_BENCH_TYPE_GPU_AMD, &platformId)
						wgBench.Done()
					}(i, platform.Id)
				} else if strings.HasSuffix(platform.Id, "_nv") { // GPU Nvidia
					wgBench.Add(1)
					go func(i int, platformId string) {
						confDbf.Platform[i].Benchmark = getBench(_BENCH_TYPE_GPU_NV, &platformId)
						wgBench.Done()
					}(i, platform.Id)
				}

				if confDbf.Platform[i].Benchmark > 0 {
					activePlat = append(activePlat, StructActivePlatform{
						Id:        platform.Id,
						Benchmark: confDbf.Platform[i].Benchmark,
					})
				}
			}
		}

		wgBench.Wait() // Wait for all benchmarks to finish

		activePlatByte, err := json.Marshal(activePlat)
		if err != nil {
			Log.Printf("%s\n", err)
		}
		activePlatStr = string(activePlatByte)

		fmt.Println("Updating config file...")

		// Update config file
		saveConfDbf()
	}
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

	vendorBenchPath := getPath(_PATH_VENDOR) + _VENDOR_TYPE_CRACKER + PATH_SEPARATOR + vendorBench + PATH_SEPARATOR + *platformId + PATH_SEPARATOR + _VENDOR_TYPE_CRACKER + extExecutable

	if checkVendor(_VENDOR_TYPE_CRACKER, &vendorBench, platformId, &vendorBenchPath) == false {
		return 0
	}

	// Preform benchmark
	fmt.Printf("Preforming benchmark (%s)...\n", *platformId)

	cmd := exec.Command(vendorBenchPath, "-b", "-m 0")

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

func checkVendor(vendorType string, vendorName *string, platformId *string, vendorPath *string) bool {
	if _, err := os.Stat(*vendorPath); err != nil {
		if os.IsNotExist(err) { // Does not exist, so get it

			fmt.Printf("Downloading %s: %s (%s)...\n", vendorType, *vendorName, *platformId)

			return getVendor(&vendorType, vendorName, platformId, vendorPath)
		} else {
			Log.Printf("%s\n", err) // Error in accessing
			return false
		}
	}

	return true
}
