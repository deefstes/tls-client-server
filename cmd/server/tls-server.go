package main

import (
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	var port int
	var crtfile, keyfile string
	flag.IntVar(&port, "p", 8000, "port to listen on")
	flag.StringVar(&crtfile, "c", "../../certs/tls.crt", "certificate file")
	flag.StringVar(&keyfile, "k", "../../certs/tls.key", "key file")
	flag.Parse()

	cert, err := tls.LoadX509KeyPair(crtfile, keyfile)
	if err != nil {
		log.Fatalf("error loading key pair: %s", err)
	}
	cfg := tls.Config{
		Certificates: []tls.Certificate{cert},
		Rand:         rand.Reader,
	}
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := tls.Listen("tcp", addr, &cfg)
	if err != nil {
		log.Fatalf("error creating listener: %s", err)
	}
	log.Printf("listening on %s", addr)

	// Infinite loop to accept connections and spin up a goroutine to handle each connection
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accept connection: %s", err)
			break
		}
		defer conn.Close()

		log.Printf("connection accepted from %s", conn.RemoteAddr())
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing connection: %s", err)
		} else {
			log.Printf("connection closed")
		}
	}()

	buf := make([]byte, 32<<10) // Create 32KiB buffer
	for {
		log.Print("waiting")
		n, err := conn.Read(buf)
		if err == io.EOF {
			log.Print("eof received from client")
			break
		} else if err != nil {
			log.Printf("error reading from client %s", err)
			break
		}
		log.Printf("echoing %d bytes - %q", n, string(buf[:n]))
		n, err = conn.Write(buf[:n])
		if err != nil {
			log.Printf("error writing to client: %s", err)
			break
		}
	}
}
