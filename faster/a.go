package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

var tags = []string{"TRACE", "INFO", "DEBUG", "ERROR", "WARN"}

// BufferSize size
const (
	BufferSize = 16
	rFC3339    = "2006-01-02T15:04:05"
)

func getTimeInput(startTime string, stopTime string) (time.Time, time.Time) {
	start, errStart := time.Parse(rFC3339, startTime)
	stop, errStop := time.Parse(rFC3339, stopTime)

	if errStart != nil || errStop != nil {
		log.Fatal("Could not parse the time provided as input")
	}
	return start, stop

}

func shouldReturn(startTime time.Time, stopTime time.Time, logLine string) bool {
	logParts := strings.SplitN(logLine, ",", 2)
	logTimeString := strings.Replace(logParts[0], " ", "T", 1)
	creationTime, err := time.Parse(rFC3339, logTimeString)
	if err != nil {
		log.Fatal("Error while parsing date in the log file", err)
	}
	if creationTime.After(startTime) && creationTime.Before(stopTime) {
		return true
	}
	return false
}

func main() {

	//f, err := os.Open("info.log")
	path := flag.String("path", "info.log", "The path of the log to be analyzed")
	level := flag.String("level", "ERROR", "Log level to look for. Default options are: TRACE, INFO, DEBUG, and ERROR")
	startTime := flag.String("startTime", "2020-01-01T18:00:58", "Start time for the log file if format YYYY-MM-DD'T'HH:MM:SS")
	stopTime := flag.String("stopTime", "2025-08-19T18:00:58", "Stop time for the log file if format YYYY-MM-DD'T'HH:MM:SS")

	flag.Parse()

	start, stop := getTimeInput(*startTime, *stopTime)
	//fmt.Println(start.String())
	f, err := os.Open(*path)

	if err != nil {
		log.Fatal("File does not exist")
		//panic("File does not exist")
	}
	defer f.Close()
	r := bufio.NewReader(f)
	var cont bool = false
	for {
		s, err := r.ReadString('\n')
		if err != nil {
			break
		}

		//time.Parse
		if strings.Contains(s, *level) {
			if shouldReturn(start, stop, s) {
				fmt.Println(s)
				cont = true
			} else {
				cont = false
			}
		} else {
			for _, value := range tags {
				if strings.Contains(s, value) {
					cont = false
					break
				}
			}
			if cont {
				fmt.Println(s)
			}
		}

	}

}
