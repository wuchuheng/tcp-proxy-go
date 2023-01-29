package main

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"github.com/fatih/color"
	"log"
	"net"
	"sync"
	"time"
)

type ClientInfo struct {
	clientId  int
	mutex     *sync.Mutex
	clientSet *map[int]bool
}

func main() {
	var clientId = 0
	clientList := make(map[int]bool)
	mutex := sync.Mutex{}
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
		clientId++
		mutex.Lock()
		clientList[clientId] = true
		mutex.Unlock()
		clientInfo := ClientInfo{clientId: clientId, mutex: &mutex, clientSet: &clientList}
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn, clientInfo)
	}
}

func handleConn(conn net.Conn, clientInfo ClientInfo) {
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
	if err != nil {
		log.Fatalf("client: dial: %s", err)
	}
	go handleImapConn(imapConn, conn, clientInfo)

	r := bufio.NewReader(conn)

	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			clientInfo.mutex.Lock()
			delete(*clientInfo.clientSet, clientInfo.clientId)
			clientInfo.mutex.Unlock()
			log.Println(err)
			return
		}
		t := time.Now()

		fmt.Println(
			convertColorStrByClientId(clientInfo.clientId,
				fmt.Sprintf(
					"C%d: %s  ",
					clientInfo.clientId, msg[:len(msg)-2],
				),
			),
			fmt.Sprintf("online: %s dateTime: %s", clientSetToStr(clientInfo), t.Format("2006-01-02 15:04:05")),
		)

		n, err := imapConn.Write([]byte(msg))
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}

func handleImapConn(imapConn net.Conn, parenConn net.Conn, info ClientInfo) {
	r := bufio.NewReader(imapConn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println(fmt.Printf("Imap: %s", err.Error()))
			return
		}
		t := time.Now()

		colorStr := convertColorStrByClientId(info.clientId, fmt.Sprintf("S%d: %s ",
			info.clientId,
			msg[:len(msg)-2],
		))
		logStr := fmt.Sprintf("%s online: %s dateTime: %s \n", colorStr, clientSetToStr(info), t.Format("2006-01-02 15:04:05"))
		fmt.Println(logStr)
		n, err := parenConn.Write([]byte(msg))
		if err != nil {
			log.Println(n, err)
			return
		}
	}
}

func clientSetToStr(info ClientInfo) string {
	result := ""
	info.mutex.Lock()
	for k, _ := range *info.clientSet {
		result = fmt.Sprintf("%s C%d", result, k)
	}
	info.mutex.Unlock()
	if len(result) > 0 {
		result = result[1:]
	}

	return fmt.Sprintf("[%s]", result)
}

type IdMapColor map[int]func(str string) string

func convertColorStrByClientId(clientId int, content string) string {
	idMapColor := IdMapColor{
		0: func(str string) string {
			return color.BlueString(str)
		},
		1: func(str string) string {
			return color.RedString(str)
		},
		2: func(str string) string {
			return color.GreenString(str)
		},
		3: func(str string) string {
			return color.YellowString(str)
		},
		4: func(str string) string {
			return color.MagentaString(str)
		},
	}
	colorId := clientId % len(idMapColor)

	return idMapColor[colorId](content)
}
