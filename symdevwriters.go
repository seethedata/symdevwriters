package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	_"strconv"
	"github.com/mcgrew/gostats"
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
	var devices = make(map[string][]float64)
	var cacheData []float64
	cacheData=make([]float64,5)
	var writeData []float64
	writeData=make([]float64,5)

	prep := exec.Command(stpToolsExe, "-f", btpFile, "-m", metricFile, "-std")
	stdout, err := prep.StdoutPipe()
	check("Reading BTP file. ", err)
	prep.Start()
	result := bufio.NewScanner(stdout)
	for result.Scan() {
		resultText := result.Text()
		if cachePattern.MatchString(resultText) {
			lineData:=strings.Split(resultText,",")
			
			for i:=2;i<len(lineData);i++ {
			 fmt.Println("Yeah " ,i,lineData[i])//cacheData[i]=strconv.ParseFloat(lineData[i],64) //strconv.ParseFloat(lineData[i],64)
			}
		} else if writePattern.MatchString(resultText) {
			lineData:=strings.Split(resultText,",")
			for i:=2;i<len(lineData);i++ {
				writeData[i]=12.0//strconv.ParseFloat(lineData[i],64)
			}
			devices[lineData[0]] = writeData
		}
	}
	for device := range devices {
				fmt.Println(device, statistics.PearsonCorrelation(devices[device],cacheData))
			}
}
