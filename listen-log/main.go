package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
)

func handleLogger(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Print(conn.LocalAddr(), " ", err)
			return
		}
		fmt.Print(conn.LocalAddr(), " ", msg)
	}
}

func main() {
	var port int

	flag.IntVar(&port, "p", 45009, "tcp post to listen")
	flag.Parse()

	input, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	defer input.Close()
	for {
		conn, err := input.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleLogger(conn)
	}
}
