package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	cert, err := tls.LoadX509KeyPair(
		"certs/client.pem",
		"certs/private.key",
	)
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	config.Rand = rand.Reader
	service := "0.0.0.0:993"
	listener, err := tls.Listen("tcp", service, &config)
	if err != nil {
		log.Fatalf("server: listen: %s", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	cert, err := tls.LoadX509KeyPair(
		"certs/client.pem",
		"certs/private.key",
	)
	if err != nil {
		log.Fatalf("server: loadkeys: %s", err)
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	imapConn, err := tls.Dial("tcp", "imap.qq.com:993", &config)
	//imapConn, err := tls.Dial("tcp", "imap.163.com:993", &config)
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	go handleImapConn(imapConn, conn)

	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		t := time.Now()
		println("-> " + t.Format("2006-01-02 15:04:05") + " " + msg)
		n, err := imapConn.Write([]byte(msg))
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}

func handleImapConn(imapConn net.Conn, parenConn net.Conn) {
	r := bufio.NewReader(imapConn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println(fmt.Printf("Imap: %s", err.Error()))
			return
		}
		t := time.Now()
		println("<- " + t.Format("2006-01-02 15:04:05") + " " + msg)

		n, err := parenConn.Write([]byte(msg))
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}
