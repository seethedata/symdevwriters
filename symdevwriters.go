package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

func check(function string, e error) {
	if e != nil {
		log.Fatal(function, e)
	}
}

func main() {
	stpToolsExe := "C:\\Program Files (x86)\\EMC\\STPTools\\stprpt.exe"
	btpFile := "btp.btp"
	metricFile := "filter.txt"
	cachePattern := regexp.MustCompile("number write pending tracks")
	writePattern := regexp.MustCompile("total writes per sec")
	var devices = make(map[string][]string)
	var cacheData []string
	var writeData []string

	prep := exec.Command(stpToolsExe, "-f", btpFile, "-m", metricFile, "-std")
	stdout, err := prep.StdoutPipe()
	check("Reading BTP file. ", err)
	prep.Start()
	result := bufio.NewScanner(stdout)
	for result.Scan() {
		resultText := result.Text()
		if cachePattern.MatchString(resultText) {
			cacheData = strings.Split(resultText, ",")
		} else if writePattern.MatchString(resultText) {
			writeData = strings.Split(resultText, ",")
			devices[writeData[0]] = writeData
		}
	}

	for device := range devices {
		for value := range devices[device] {
			if value > 1 {
				fmt.Println(device, devices[device][value],cacheData[value])
			}
		}
	}
}
