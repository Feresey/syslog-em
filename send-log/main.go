package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// timestamp format
const timeFormatLayout string = "01/02/2006-15:04:05.000000"

// Читает из сканера, парсит время, возвращает событие и время
func parseEvent(fileScanner *bufio.Scanner, timeRegexp *regexp.Regexp) (time.Time, []byte, error) {
	logLine := fileScanner.Bytes()
	if bytes.Equal(logLine, []byte("")) {
		return time.Time{}, []byte{}, fmt.Errorf("Empty string")
	}
	match := timeRegexp.FindIndex(logLine)
	if len(match) == 0 {
		return time.Time{}, []byte{}, fmt.Errorf("Can't find timestamp")
	}
	timeString := logLine[match[0]:match[1]]
	logLine = logLine[match[0]:]
	eventTime, _ := time.Parse(timeFormatLayout, string(timeString))
	logLine = append(logLine, byte('\n'))
	return eventTime, logLine, nil
}

// подменяет время события
func replaceTimestamp(out []byte) {
	systemTime := time.Now().Format(timeFormatLayout)
	copy(out[:], systemTime)
}

func processFile(file io.Reader, writer net.Conn) error {
	fileScanner := bufio.NewScanner(file)
	if !fileScanner.Scan() {
		return nil
	}

	timeRegexp := regexp.MustCompile(`(?m)\d{2}/\d{2}/\d{4}-\d{2}:\d{2}:\d{2}.\d{6}`)

	prevEventTime, logLine, err := parseEvent(fileScanner, timeRegexp)
	if err != nil {
		return err
	}

	replaceTimestamp(logLine)
	fmt.Print(string(logLine))

	_, err = writer.Write(logLine)
	if err != nil {
		return err
	}

	for fileScanner.Scan() {
		currentEventTime, logLine, err := parseEvent(fileScanner, timeRegexp)
		if err != nil {
			if err == fmt.Errorf("Empty string") {

			} else {
				return err
			}
		}

		timer := time.NewTimer(currentEventTime.Sub(prevEventTime))
		<-timer.C

		replaceTimestamp(logLine)
		fmt.Print(string(logLine))

		_, err = writer.Write(logLine)
		if err != nil {
			return err
		}

	}

	return nil
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("./send-log file host:port")
	}
	inputFileName, hostAddress := os.Args[1], os.Args[2]
	inputFileExt := filepath.Ext(inputFileName)

	fmt.Printf("File: %s, ext: %s, Host: %s\n", inputFileName, inputFileExt, hostAddress)

	logConn, err := net.Dial("tcp", hostAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer logConn.Close()

	switch inputFileExt {
	case ".zip":
		zipReader, err := zip.OpenReader(inputFileName)
		if err != nil {
			log.Fatal(err)
		}
		defer zipReader.Close()

		for _, inputFile := range zipReader.File {
			inputFileReader, err := inputFile.Open()
			if err != nil {
				log.Fatal(err)
			}
			log.Println("File:", inputFile.FileInfo().Name())
			err = processFile(inputFileReader, logConn)
			if err != nil {
				log.Fatal(err)
			}
			inputFileReader.Close()
		}

	default:
		inputFile, err := os.Open(inputFileName)
		if err != nil {
			log.Fatal(err)
		}
		defer inputFile.Close()

		inputFileStat, err := inputFile.Stat()
		if err != nil {
			log.Fatal(err)
		}
		if inputFileStat.IsDir() {
			log.Fatalln(inputFileName, " - is a dir")
		} else {
			err := processFile(inputFile, logConn)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
