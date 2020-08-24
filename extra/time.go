package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	layoutISO = "2006-01-02"
	layoutUS  = "January 2, 2006"
	rFC3339   = "2006-01-02T15:04:05"
	layout    = "2006-01-02T15:04:05.000Z"
)

func main() {
	s := "2020-08-19 17:59:44,127 [][3400][C411T15][cb2222ca-2bae-49ab-bee0-dbe5ad62ac54][WARN] - MOCK: LPSIncidentReceiver_V2 mock is on, not actually calling service"
	logParts := strings.SplitN(s, ",", 2)
	// fmt.Println(logParts[0])
	logCreationTimeString := strings.Replace(logParts[0], " ", "T", 1)

	// fmt.Println(logCreationTimeString)

	// date := "2009-01-02T15:04:05"
	//date := "2020-08-19T18:00:58"
	end := "2021-08-19T18:00:58"
	start := "2019-08-19T18:00:58"
	t, _ := time.Parse(rFC3339, logCreationTimeString)

	startTime, _ := time.Parse(rFC3339, start)
	endTime, _ := time.Parse(rFC3339, end)

	fmt.Println(t) // 1999-12-31 00:00:00 +0000 UTC
	fmt.Println(t.Format(rFC3339))

	if t.After(startTime) && t.Before(endTime) {
		fmt.Println("text")
	}

	layout := "2006-01-02 15:04:05 -0700"
	dateString := "2018-12-17 12:55:50 +0300"
	t, err := time.Parse(layout, dateString)
	if err != nil {
		fmt.Println("Error while parsing date :", err)
	}
	fmt.Println(t.Format("2006-01-02 15:04:05"))
}
