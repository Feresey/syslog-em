package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"flag"
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

type emptyStringError struct {
}

func (emptyStringError) Error() string {
	return "Empty string"
}

// Читает из сканера, парсит время, возвращает событие и время
func parseEvent(fileScanner *bufio.Scanner, timeRegexp *regexp.Regexp) (time.Time, []byte, error) {
	logLine := fileScanner.Bytes()
	if bytes.Equal(logLine, []byte("")) {
		return time.Time{}, []byte{}, emptyStringError{}
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
	copy(out, systemTime)
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
			var emptyError emptyStringError
			if errors.Is(err, emptyError) {
				continue
			} else {
				fmt.Println("awd", err)
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
	var (
		filename string
		host     string
	)

	flag.StringVar(&filename, "f", "", "file to use")
	flag.StringVar(&host, "h", "", "host:port to send logs")
	flag.Parse()

	if len(host) == 0 || len(filename) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	inputFileExt := filepath.Ext(filename)

	fmt.Printf("File: %s, ext: %s, Host: %s\n", filename, inputFileExt, host)

	logConn, err := net.Dial("tcp", host)
	if err != nil {
		log.Fatal(err)
	}
	defer logConn.Close()
	log.Printf("Connection to %s established.", host)

	switch inputFileExt {
	case ".zip":
		err = runZip(logConn, filename)
	default:
		err = runFile(logConn, filename)
	}
	if err != nil {
		log.Fatal("Process file(s): ", err)
	}
}

func runZip(conn net.Conn, filename string) error {
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, inputFile := range zipReader.File {
		inputFileReader, err := inputFile.Open()
		if err != nil {
			return err
		}

		log.Println("File:", inputFile.FileInfo().Name())
		err = processFile(inputFileReader, conn)
		_ = inputFileReader.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func runFile(conn net.Conn, filename string) error {
	inputFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	inputFileStat, err := inputFile.Stat()
	if err != nil {
		return err
	}

	if inputFileStat.IsDir() {
		log.Println(filename, " - is a dir. skipping...")
	} else {
		err := processFile(inputFile, conn)
		if err != nil {
			return err
		}
	}

	return nil
}
