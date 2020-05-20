package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"time"
)

// timestamp format
const timeFormatLayout string = "01/02/2006-15:04:05.000000"

// Читает из сканера, парсит время, возвращает событие и время
func parseEvent(fileScanner *bufio.Scanner, timeRegexp *regexp.Regexp) (time.Time, []byte) {
	logLine := fileScanner.Bytes()
	match := timeRegexp.FindIndex(logLine)
	timeString := logLine[match[0]:match[1]]
	logLine = logLine[match[0]:]
	eventTime, _ := time.Parse(timeFormatLayout, string(timeString))
	logLine = append(logLine, byte('\n'))
	return eventTime, logLine
}

// подменяет время события
func replaceTimestamp(out []byte) {
	systemTime := time.Now().Format(timeFormatLayout)
	copy(out[:], systemTime)
}

func processFile(file *os.File, writer net.Conn) error {
	fileScanner := bufio.NewScanner(file)
	if !fileScanner.Scan() {
		return nil
	}

	timeRegexp := regexp.MustCompile(`(?m)\d{2}/\d{2}/\d{4}-\d{2}:\d{2}:\d{2}.\d{6}`)

	prevEventTime, logLine := parseEvent(fileScanner, timeRegexp)

	replaceTimestamp(logLine)
	fmt.Print(string(logLine))

	_, err := writer.Write(logLine)
	if err != nil {
		log.Fatal(err)
	}

	for fileScanner.Scan() {
		currentEventTime, logLine := parseEvent(fileScanner, timeRegexp)

		timer := time.NewTimer(currentEventTime.Sub(prevEventTime))
		<-timer.C

		replaceTimestamp(logLine)
		fmt.Print(string(logLine))

		_, err := writer.Write(logLine)
		if err != nil {
			log.Fatal(err)
		}

	}

	return nil
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("./send-log file host:port")
	}
	inputFileName, hostAddress := os.Args[1], os.Args[2]
	fmt.Printf("File: %s Host: %s\n", inputFileName, hostAddress)

	logConn, err := net.Dial("tcp", hostAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer logConn.Close()

	inputFile, err := os.Open(inputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	processFile(inputFile, logConn)
}
