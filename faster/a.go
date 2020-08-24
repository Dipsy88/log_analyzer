package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

var tags = []string{"TRACE", "INFO", "DEBUG", "ERROR", "WARN"}

// BufferSize size
const (
	BufferSize = 16
	rFC3339    = "2006-01-02T15:04:05"
)

type chunk struct {
	bufsize int
	offset  int64
}

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
		fmt.Println("Error while parsing date in the log file", err)
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
	fmt.Println(start.String())
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

// ProcessChunk is chunk
func ProcessChunk(chunk []byte, linesPool *sync.Pool, stringPool *sync.Pool, start time.Time, end time.Time) {

	var wg2 sync.WaitGroup

	logs := stringPool.Get().(string)
	logs = string(chunk)

	linesPool.Put(chunk)

	logsSlice := strings.Split(logs, "\n")

	stringPool.Put(logs)

	chunkSize := 300
	n := len(logsSlice)
	noOfThread := n / chunkSize

	if n%chunkSize != 0 {
		noOfThread++
	}

	for i := 0; i < (noOfThread); i++ {

		wg2.Add(1)
		go func(s int, e int) {
			defer wg2.Done() //to avaoid deadlocks
			for i := s; i < e; i++ {
				text := logsSlice[i]
				if len(text) == 0 {
					continue
				}
				logSlice := strings.SplitN(text, ",", 2)
				logCreationTimeString := logSlice[0]

				logCreationTime, err := time.Parse("2006-01-02T15:04:05.0000Z", logCreationTimeString)
				if err != nil {
					fmt.Printf("\n Could not able to parse the time :%s for log : %v", logCreationTimeString, text)
					return
				}

				if logCreationTime.After(start) && logCreationTime.Before(end) {
					//fmt.Println(text)
				}
			}

		}(i*chunkSize, int(math.Min(float64((i+1)*chunkSize), float64(len(logsSlice)))))
	}

	wg2.Wait()
	logsSlice = nil
}
