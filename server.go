package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

type eventType int

const (
	Follow eventType = 1 << iota
	Unfollow
	Broadcast
	PrivateMessage
	StatusUpdate
)

type event struct {
	SequenceID int
	Type       eventType
	FromID     int
	ToID       int
	RawMessage string
}

type Server struct {
	eventSrcSrv net.Listener
	clientSrv   net.Listener

	done chan struct{}
}

func StartServer(eventSrcAddr, clientAddr string) (*Server, error) {

	done := make(chan struct{}) // used to notify Server goroutines to stop when we're done.

	eventSrcSrv, err := net.Listen("tcp", eventSrcAddr)
	if err != nil {
		return nil, err
	}
	clientSrv, err := net.Listen("tcp", clientAddr)
	if err != nil {
		return nil, err
	}

	// TODO: start the notifier.

	// TODO: We will get these from the notifier.
	eventch := make(chan event)
	clientConnch := make(chan clientConn)

	// start listeners for events and clients in separate goroutines.
	go listen(eventSrcSrv, eventSrcHandler(eventch), done)
	go listen(clientSrv, clientHandler(clientConnch), done)

	return &Server{}, nil
}

type handler func(c net.Conn)

// eventSrcHandler returns a handler func which:
// parses an event.
// sends it to events channel.
func eventSrcHandler(eventch chan<- event) handler {
	return func(conn net.Conn) {
		scanner := bufio.NewScanner(conn)

		for scanner.Scan() {
			eventStr := scanner.Text()
			e, err := parseEvent(eventStr)
			if err != nil || scanner.Err() != nil {
				log.Print(err)
			}
			eventch <- e
		}
	}
}

type clientConn struct {
	conn net.Conn
	ID   int
}

func listen(l net.Listener, h handler, done chan struct{}) {
	for {
		// Accept a connection in a goroutine and send it to connCh.
		connCh := make(chan net.Conn)
		go func() {
			conn, err := l.Accept()
			if err != nil {
				log.Printf("connection error", err)
			}
			connCh <- conn
		}()

		select {
		case c := <-connCh:
			go h(c)

		case <-done:
			l.Close()
			return
		}
	}
}

func clientHandler(clientConnch chan<- clientConn) handler {
	return func(conn net.Conn) {
		scanner := bufio.NewScanner(conn)
		scanner.Scan()
		clientIDstr := scanner.Text()
		clientID, err := strconv.Atoi(clientIDstr)
		if err != nil || scanner.Err() != nil {
			log.Print(err)
		}
		clientConnch <- clientConn{ID: clientID, conn: conn}
	}
}

// Parses event string into an event struct.
func parseEvent(s string) (event, error) {
	var e event
	var err error
	eventParts := strings.Split(s, "|")
	if len(eventParts) < 2 {
		errMsg := fmt.Sprintf("error parsing event - invalid event %s\n", s)
		return e, errors.New(errMsg)
	}

	e.RawMessage = s
	e.SequenceID, err = strconv.Atoi(eventParts[0])
	if err != nil {
		return e, err
	}

	eventType := eventParts[1]
	switch eventType {
	case "F":
		e.Type = Follow
		goto parseFromIDToID
	case "U":
		e.Type = Unfollow
		goto parseFromIDToID
	case "B":
		e.Type = Broadcast
		return e, nil
	case "P":
		e.Type = PrivateMessage
		goto parseFromIDToID
	case "S":
		e.Type = StatusUpdate
		if len(eventParts) != 3 {
			errMsg := fmt.Sprintf("error parsing event: %s", s)
			return e, errors.New(errMsg)
		}
		e.FromID, err = strconv.Atoi(eventParts[2])
		return e, nil
	default:
		errMsg := fmt.Sprintf("unknown event type: %s", eventType)
		return e, errors.New(errMsg)
	}

parseFromIDToID:
	if len(eventParts) != 4 {
		errMsg := fmt.Sprintf("error parsing event: %s", s)
		return e, errors.New(errMsg)
	}
	e.FromID, err = strconv.Atoi(eventParts[2])
	e.ToID, err = strconv.Atoi(eventParts[3])

	return e, err
}

func main() {
}
