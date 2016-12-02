package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"crypto/sha256"
	"crypto/rand"
	"crypto/aes"
	"crypto/hmac"
	"crypto/cipher"
)

var PORT = "3939"

func main() {
	clientPtr := flag.String("c", "", "Address of server to connect to")
	serverPtr := flag.Bool("s", false, "Flag to start server")
	confPtr   := flag.String("confkey", "", "Key to use for message confidentiality")
	authPtr   := flag.String("authkey", "", "Key to use for message authenticity")
	//helpPtr   := flag.Bool("h", false, "Flag for Help")
	flag.Parse()
	var c net.Conn
	if len(*confPtr) == 0 || len(*authPtr) == 0 {
		log.Fatal("Must provide both a confkey and authkey argument")
	}
	if *serverPtr && *clientPtr == "" {
		c = serverStart()
	} else if *clientPtr != "" && !*serverPtr {
		c = clientConnect(*clientPtr)
	} else {
		log.Fatal("Incorrect command line arguments")
	}
	tmp := sha256.Sum256([]byte(*confPtr))
	confHash := tmp[:16]
	tmp  = sha256.Sum256([]byte(*authPtr))
	// yay slice magics
	authHash := tmp[:]
	readCh := make(chan []byte)
	writeCh := make(chan []byte)
	errorCh := make(chan error)
	go asyncRead(c, readCh, errorCh, confHash, authHash)
	go asyncWrite(c, writeCh, errorCh, confHash, authHash)
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

func asyncWrite(conn net.Conn, writeCh chan []byte, errCh chan error, confkey, authkey []byte) {
	scanner := bufio.NewScanner(os.Stdin)
	newline := "\n"
	// Crypto Objects
	block, err := aes.NewCipher(confkey)
	blocksize := block.BlockSize()
	hmacHash   := hmac.New(sha256.New, authkey)
	if err != nil {
		log.Fatal("Couldn't create an AES Cipher for some reason")
	}
	for scanner.Scan() {
		msg := scanner.Bytes()
		IV  := make([]byte, blocksize)
		_, err := rand.Read(IV)
		if err != nil {
			log.Fatal("Something bad happened generating an IV")
		}
		msg = append(msg, newline...)
		bufferAmt := blocksize-(len(msg)%blocksize)
		if bufferAmt > 0 {
			tmp := make([]byte, bufferAmt)
			msg = append(msg, tmp...)
		}
		mode := cipher.NewCBCEncrypter(block, IV)
		mode.CryptBlocks(msg, msg)
		msg = append(IV, msg...)
		msg = hmacHash.Sum(msg)
		writeCh <- msg
	}

}

func asyncRead(conn net.Conn, readCh chan []byte, errCh chan error, confkey, authkey []byte) {
	block, _ := aes.NewCipher(confkey)
	hmacHash   := hmac.New(sha256.New, authkey)
	blocksize := block.BlockSize()
	for {
		data := make([]byte, 512)
		_, err := conn.Read(data)
		if err != nil {
			errCh <- err
			return
		}
		trimIndex := len(data)-1
		for i := len(data)-1; i > 0; i-- {
			if data[i] != 0 {
				diff := (i+1)%blocksize
				if diff < 1 {
					i += 1-diff 
				} else if diff > 1 {
					i += (blocksize-diff)+4
				}
				trimIndex = i
				break
			}
		}
		data = data[:trimIndex]
		macIndex := len(data)-32
		msg := data[:macIndex]
		if !hmac.Equal(hmacHash.Sum(msg), data) {
			log.Print("Invalid HMAC Encountered - someone might be trying to do bad things")
			continue
		}
		IV  := msg[:blocksize]
		msg = msg[blocksize:]
		mode := cipher.NewCBCDecrypter(block, IV)
		mode.CryptBlocks(msg, msg)
		readCh <- msg
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
