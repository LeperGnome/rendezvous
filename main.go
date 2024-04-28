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
	CREATE_ROOM = 0x10
	JOIN_ROOM   = 0x20
)

type Rooms struct {
	rooms map[uuid.UUID]net.Addr
	mu    sync.RWMutex
}

func NewRooms() Rooms {
	return Rooms{
		rooms: make(map[uuid.UUID]net.Addr),
		mu:    sync.RWMutex{},
	}
}

func (r *Rooms) Create(from net.Addr) (uuid.UUID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.UUID{}, err
	}

	r.rooms[id] = from
	return id, nil
}

func (r *Rooms) Remove(id uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.rooms, id)
}

func (r *Rooms) Get(id uuid.UUID) (net.Addr, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	addr, ok := r.rooms[id]
	if !ok {
		return nil, fmt.Errorf("Room with ID %s is not found", id.String())
	}
	return addr, nil
}

func main() {
	laddr := "0.0.0.0:4589"
	hostAddr := net.TCPAddrFromAddrPort(netip.MustParseAddrPort(laddr))

	listner, err := net.ListenTCP("tcp", hostAddr)
	if err != nil {
		panic("Could not start listner: " + err.Error())
	}
	defer listner.Close()

	rooms := NewRooms()

	fmt.Printf("Accepting connections on %s\n", laddr)
	for {
		conn, err := listner.AcceptTCP()
		if err != nil {
			fmt.Printf("Could not accept connection: %s\n", err.Error())
			continue
		}
		go handleConn(conn, &rooms)
	}

}

func handleConn(conn *net.TCPConn, rooms *Rooms) {
	fmt.Println("New connection")
	defer func() {
		fmt.Println("Closing connection")
		conn.Close()
	}()
	buff := make([]byte, 1024)
    // TODO: not reading version yet

	for {
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Printf("Could not read from connection: %s\n", err.Error())
			return
		}
		if buff[0] == CREATE_ROOM {
			// creating room
			id, err := rooms.Create(conn.RemoteAddr())
			if err != nil {
				fmt.Printf("Failed creating room: %s\n", err.Error())
				return
			}
			// remove on disconnect
			defer rooms.Remove(id)
			fmt.Printf("Created a room: %s\n", id.String())

			// response
			_, err = conn.Write([]byte(id.String()))
			if err != nil {
				fmt.Printf("Failed responding: %s\n", err.Error())
				return
			}

		} else if buff[0] == JOIN_ROOM {
			// getting room
			idRaw := string(buff[1 : n-1]) // TODO: should trim newline better
			id, err := uuid.Parse(idRaw)
			if err != nil {
				fmt.Printf("Error joining room: %s\n", err.Error())
				return
			}

			rAddr, err := rooms.Get(id)
			if err != nil {
				fmt.Printf("Error joining room: %s\n", err.Error())
				return
			}

			// response
			_, err = conn.Write([]byte(rAddr.String()))
			if err != nil {
				fmt.Printf("Failed responding: %s\n", err.Error())
				return
			}

		} else {
			fmt.Printf("Unknown command: %v\n", buff[0])
			return
		}
	}
}
