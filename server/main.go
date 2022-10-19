package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

type client chan<- string

type clientPerson struct {
	nickname string
	who      string
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)
	who := conn.RemoteAddr().String()
	ch <- "You are " + who
	messages <- who + " has arrived"
	entering <- ch

	ch <- "What is your nickname?"

	var client = clientPerson{
		who: who,
	}

	client.processInput(conn, ch)

	leaving <- ch
	client.notifyAboutLeaving()
	err := conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func (c *clientPerson) processInput(conn net.Conn, ch chan string) {
	input := bufio.NewScanner(conn)

	for input.Scan() {
		c.writeNickname(input, ch)
		c.writeMessage(input)
	}
}

func (c *clientPerson) writeNickname(input *bufio.Scanner, ch chan string) {
	if len(c.nickname) == 0 {
		c.nickname = input.Text()
		ch <- "Wrote your nickname as: " + c.nickname
	}
}

func (c *clientPerson) writeMessage(input *bufio.Scanner) {
	if len(c.nickname) == 0 {
		messages <- c.who + ": " + input.Text()
		return
	}

	messages <- c.nickname + ": " + input.Text()
}

func (c *clientPerson) notifyAboutLeaving() {
	if len(c.nickname) == 0 {
		messages <- c.who + " has left"
		return
	}

	messages <- c.nickname + " has left"
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		_, err := fmt.Fprintln(conn, msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}
