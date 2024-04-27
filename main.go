package main

/*
TCP

   1 byte   1 byte
| version |   cmd   | ... |

cmds:

0x1: create room
    no further body
    resp - uuid

0x2: connect to room
    data - uuid
    resp - host IP

*/

import (
	"fmt"
	"net"
	"net/netip"
	"sync"

	"github.com/google/uuid"
)

const (
	VERSOIN = 1
)

// CMDs
const (
	CREATE_ROOM = 0x1
	JOIN_ROOM   = 0x2
)

type Rooms struct {
	rooms map[uuid.UUID]net.IP
	mu    sync.RWMutex
}

func (r *Rooms) Create(fromIP net.IP) (*uuid.UUID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	r.rooms[id] = fromIP
	return &id, nil
}

func (r *Rooms) Join(id uuid.UUID) (*net.IP, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	el, ok := r.rooms[id]
	if !ok {
		return nil, fmt.Errorf("Room with ID %s is not found", id.String())
	}
	return &el, nil
}

func main() {
	hostAddr := net.TCPAddrFromAddrPort(netip.MustParseAddrPort("0.0.0.0:4589"))

	listner, err := net.ListenTCP("tpc", hostAddr)
	if err != nil {
		panic("Could not start listner: " + err.Error())
	}
	defer listner.Close()

	for {
		conn, err := listner.AcceptTCP()
		if err != nil {
			fmt.Printf("Could not accept connection: %s", err.Error())
			continue
		}
		go handleConn(conn)
	}

}

func handleConn(conn *net.TCPConn) {
	// TODO
}
