package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	var addr string
	var infile string
	var crtfile, keyfile string
	flag.StringVar(&addr, "a", "127.0.0.1:8000", "address of tls host")
	flag.StringVar(&infile, "i", "", "input file")
	flag.StringVar(&crtfile, "c", "../../certs/tls.crt", "certificate file")
	flag.StringVar(&keyfile, "k", "../../certs/tls.key", "key file")
	flag.Parse()

	cert, err := tls.LoadX509KeyPair(crtfile, keyfile)
	if err != nil {
		log.Fatalf("error loading key pair: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr, &config)
	if err != nil {
		log.Fatalf("error connecting to host: %s", err)
	}
	defer conn.Close()
	log.Printf("connected to %s", conn.RemoteAddr())

	var reader *bufio.Reader
	if infile != "" {
		f, err := os.Open(infile)
		if err != nil {
			log.Fatalf("error reading input file: %s", err)
			return
		}
		defer f.Close()

		reader = bufio.NewReader(f)
	} else {
		reader = bufio.NewReader(os.Stdin)
	}

	quit := make(chan interface{})

	// Goroutine to read from host
	go func() {
		for {
			select {
			case <-quit:
				log.Println("reading done")
				return
			default:
				reply := make([]byte, 256)
				n, err := conn.Read(reply)
				if err == io.EOF {
					log.Print("eof received from host")
					quit <- true
				} else if err != nil {
					log.Printf("error reading from host: %s", err)
					quit <- true
				} else {
					log.Printf("received %d bytes - %q", n, string(reply[:n]))
				}
			}
		}
	}()

	// Infinite loop to read from stdin / input file and write to host
	for {
		select {
		case <-quit:
			log.Println("writing done")
			break
		default:
			var message string
			var err error
			if infile != "" {
				var buf []byte
				buf, err = reader.ReadBytes(16) // Reads up to the next newline byte (\n = 0x10 = 16)
				if len(buf) > 0 {
					message = string(buf[:len(buf)-1])
				}
			} else {
				message, err = reader.ReadString('\n')
			}
			if err != nil {
				log.Fatalf("error reading from source: %s", err)
				return
			}

			n, err := io.WriteString(conn, message)
			if err != nil {
				log.Fatalf("error writing to host: %s", err)
			}
			log.Printf("write %d bytes - %q", n, message)

			// This delay is of course not necessary but can be usfeul to approximate the kind of delays that can be expected from a real client device.
			// Also, because the tx/rx is asynchronous, this gives the host time to respond before the next message is transmitted which is especially useful
			// if a binary file containing newline seperated messages is used as input. It roughly simulates synchronous behaviour.
			time.Sleep(1000 * time.Millisecond)
		}
	}
}
