package tracker

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"shareDir/iptracker/peerinfo"
	"strconv"
	"strings"
)

type Tracker struct {
	port     int
	peer     peerinfo.PeerManager
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelCauseFunc
}

func New(port int) *Tracker {
	ctx, cancel := context.WithCancelCause(context.Background())
	peer := peerinfo.GetPeerManager()
	return &Tracker{
		port:   port,
		peer:   peer,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (t *Tracker) Start() error {
	address := "0.0.0.0:" + strconv.Itoa(t.port)
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("fail to create a listener: %w", err)
	}
	t.listener = listen

	fmt.Println("Tracker IP tracker is running, listening on port:", t.port)

	for {
		conn, err := listen.Accept()
		if err != nil {
			select {
			case <-t.ctx.Done():
				return nil
			default:
				fmt.Println("Fail to accept a connection:", err)
				continue
			}
		}

		go t.handleHeartBeat(conn)
	}
}

func (t *Tracker) Stop(reason error) {
	t.cancel(reason)
	if t.listener != nil {
		t.listener.Close()
	}
	fmt.Println("Tracker connection closed gracefully.")
}

func (t *Tracker) handleHeartBeat(conn net.Conn) {
	defer conn.Close()

	addr := conn.RemoteAddr().String()
	fmt.Println("Receive heartbeat from:", addr)

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		message := scanner.Text()
		fields := strings.Fields(message)
		if len(fields) > 0 {
			peerID := fields[0]

			t.peer.Update(peerID, addr)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	_, _ = conn.Write([]byte("Copy\n"))
}
