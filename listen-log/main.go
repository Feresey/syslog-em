package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
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
	if len(os.Args) != 2 {
		log.Fatal("./send-log :port")
	}
	port := os.Args[1]
	input, err := net.Listen("tcp", port)
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
