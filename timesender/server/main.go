package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

var (
	messages    = make(chan string)
	connections = map[net.Conn]bool{}
)

func main() {
	lc := net.ListenConfig{}
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	listener, err := lc.Listen(ctx, "tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
				return
			}
			wg.Add(1)
			connections[conn] = true
			go handleConn(ctx, conn, wg)
		}
	}()

	<-ctx.Done()

	err = listener.Close()
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()
}

func listenForMessage() {
	var msg string

	for {
		_, err := fmt.Fscan(os.Stdin, &msg)
		if err != nil {
			log.Fatal(err)
		}
		messages <- msg
	}

}

func handleConn(ctx context.Context, c net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func(c net.Conn) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)

	go listenForMessage()

	for {
		select {
		case <-ctx.Done():
			_, err := fmt.Fprintln(c, ", has left")
			if err != nil {
				log.Fatal(err)
			}
			delete(connections, c)
			return
		case msg := <-messages:
			for connect := range connections {
				_, err := fmt.Fprintln(connect, time.Now().Format("15:04:05\n\r"), msg)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
