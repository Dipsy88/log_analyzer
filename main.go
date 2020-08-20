package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	//f, err := os.Open("info.log")
	path := flag.String("path", "info.log", "The path of the log to be analyzed")
	level := flag.String("level", "ERROR", "Log level to look for. Default options are: TRACE, INFO, DEBUG, and ERROR")

	flag.Parse()
	f, err := os.Open(*path)

	if err != nil {
		log.Fatal("File does not exist")
		//panic("File does not exist")
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		s, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if strings.Contains(s, *level) {
			fmt.Println(s)
		}
	}
}
