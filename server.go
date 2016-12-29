package main

import (
	"errors"
	"fmt"
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
