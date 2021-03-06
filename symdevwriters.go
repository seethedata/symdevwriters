// symdevwriters reads a BTP file and prints a list of devices in decending order of correlation to system cache usage.
// The goal of this analysis is to help identify devices that contribute to high cache utilization.

package main

import (
	"bufio"
	"fmt"
	"github.com/mcgrew/gostats"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// sortmap.go from https://gist.github.com/ikbear/4038654
type sortedMap struct {
	m map[string]float64
	s []string
}

func (sm *sortedMap) Len() int {
	return len(sm.m)
}

func (sm *sortedMap) Less(i, j int) bool {
	return sm.m[sm.s[i]] > sm.m[sm.s[j]]
}

func (sm *sortedMap) Swap(i, j int) {
	sm.s[i], sm.s[j] = sm.s[j], sm.s[i]
}

func sortedKeys(m map[string]float64) []string {
	sm := new(sortedMap)
	sm.m = m
	sm.s = make([]string, len(m))
	i := 0
	for key, _ := range m {
		sm.s[i] = key
		i++
	}
	sort.Sort(sm)
	return sm.s
}

func check(function string, e error) {
	if e != nil {
		log.Fatal(function, e)
	}
}

func LocateFile(exe string) string {
	progDirOld := `C:\Program Files (x86)\EMC\STPTools\` + exe
	progDirNew := `C:\Program Files\EMC\STPTools\` + exe
	fileLocation := ""

	if _, err := os.Stat(progDirNew); err == nil {
		fileLocation = progDirNew
	} else if _, err := os.Stat(progDirOld); err == nil {
		fileLocation = progDirOld
	} else {
		log.Fatal(exe + " is required, but is not found.\nLocations checked were:\n" + progDirNew + "\n" + progDirOld)
	}
	return fileLocation
}

func main() {
	stpToolsExe := LocateFile("StpRpt.exe")
	metricFile := "./filter.txt"

	cachePattern := regexp.MustCompile("number write pending tracks")
	writePattern := regexp.MustCompile("total writes per sec")
	cmdArgs := os.Args[1:]
	var btpFile string
	allFlag := "N"

	var devices = make(map[string][]float64)
	cacheData := make([]float64, 0)
	writeData := make([]float64, 0)
	
	if len(cmdArgs) < 1 {
		fmt.Println("USAGE: symdevwriters BTP_FILE_NAME [Y]")
		fmt.Println("By default, only the top 50 correlated devices will be shown. Setting \"Y\" as the second argument will show correlation values for all devices.")
		return
	} else if len(cmdArgs) == 2 {
		btpFile = cmdArgs[0]
		allFlag = cmdArgs[1]
	} else {
		btpFile = cmdArgs[0]
	}
	
	if _, err := os.Stat(btpFile); err != nil {
		fmt.Println("Unable to find file:", btpFile)
		return
	}
	
	f, err := os.Create(metricFile)
	check("Create filter file", err)
	defer func() {
		err=f.Close()
		check("Closing filter file",err)
		err=os.Remove(f.Name())
		check("Removing filter file",err)
	}()

	_, err = f.WriteString("System::number write pending tracks\n")
	check("Write first line to file", err)
	_, err = f.WriteString("Devices::total writes per sec")
	check("Write second line to file", err)
	
	fmt.Print("Reading BTP file ( "+btpFile+" )...")
	prep := exec.Command(stpToolsExe, "-f", btpFile, "-m", metricFile, "-std")
	stdout, err := prep.StdoutPipe()
	check("Reading BTP file. ", err)
	prep.Start()
	result := bufio.NewScanner(stdout)
	for result.Scan() {
		resultText := result.Text()
		if cachePattern.MatchString(resultText) {
			lineData := strings.Split(resultText, ",")
			for i := 2; i < len(lineData); i++ {
				num, err := strconv.ParseFloat(strings.TrimSpace(lineData[i]), 64)
				check("Parse Float", err)
				cacheData = append(cacheData, num)
			}
		} else if writePattern.MatchString(resultText) {
			lineData := strings.Split(resultText, ",")
			for i := 2; i < len(lineData); i++ {
				num, err := strconv.ParseFloat(strings.TrimSpace(lineData[i]), 64)
				check("Parse Float", err)
				writeData = append(writeData, num)
			}
			devices[lineData[0]] = writeData
			writeData = nil
		}
	}
	
	err=prep.Wait()
	check("Reading BTP file.",err)
	
	fmt.Println("Done.")
	
	correl := map[string]float64{}

	for device := range devices {
		dev := device
		if statistics.Max(devices[dev]) > 0 {
			correl[dev] = statistics.PearsonCorrelation(devices[dev], cacheData)
		}
	}
	if allFlag == "N" {
		fmt.Println("Top 50 correlated LUN writers are:")
	}
	fmt.Printf("%4s %10s %15s\n", "#", "Device", "Correl. Coeff.")
	for i, dev := range sortedKeys(correl) {
		fmt.Printf("%4d %10s %15.4f\n", i+1, dev, correl[dev])
		if i == 49 && allFlag == "N" {
			break
		}
	}
}
