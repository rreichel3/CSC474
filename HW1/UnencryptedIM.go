package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var PORT = "3939"

func main() {
	clientPtr := flag.String("c", "", "Address of server to connect to")
	serverPtr := flag.Bool("s", false, "Flag to start server")
	flag.Parse()
	var c net.Conn
	if *serverPtr {
		c = serverStart()
	} else if *clientPtr != "" {
		c = clientConnect(*clientPtr)
	} else {
		log.Fatal("Incorrect command line arguments")
	}

	readCh := make(chan []byte)
	writeCh := make(chan []byte)
	errorCh := make(chan error)
	go asyncRead(c, readCh, errorCh)
	go asyncWrite(c, writeCh, errorCh)
	for {
		select {
		case data := <-readCh:
			fmt.Print(string(data))
		case data := <-writeCh:
			_, err := c.Write(data)
			if err != nil {
				errorCh <- err
			}
		case err := <-errorCh:
			log.Fatal(err)
		}
	}
}

func asyncWrite(conn net.Conn, writeCh chan []byte, errCh chan error) {
	scanner := bufio.NewScanner(os.Stdin)
	newline := "\n"
	for scanner.Scan() {
		msg := scanner.Bytes()
		msg = append(msg, newline...)
		writeCh <- msg
	}

}

func asyncRead(conn net.Conn, readCh chan []byte, errCh chan error) {
	for {
		data := make([]byte, 512)
		_, err := conn.Read(data)
		if err != nil {
			errCh <- err
			return
		}
		readCh <- data
	}
}

func clientConnect(address string) net.Conn {
	conn, err := net.Dial("tcp", address+":"+PORT)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func serverStart() net.Conn {
	ln, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		log.Fatal(err)
	}
	// This should be in a loop but since we're waiting for a single p2p connection, we're not looping
	// (Also it should be handled asynchrously if its a "real" server)
	conn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}
	return conn
}
