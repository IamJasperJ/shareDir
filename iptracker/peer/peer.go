package peer

import (
	"context"
	"fmt"
	"net"
	"time"
)

type Peer struct {
	id         string
	serverAddr string
	ctx        context.Context
	cancel     context.CancelCauseFunc
}

func New(id, serverAddr string) *Peer {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &Peer{
		id:         id,
		serverAddr: serverAddr,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (p *Peer) Start() error {
	fmt.Println("Peer IP tracker is running")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	p.heartbeat()

	for {
		select {
		case <-p.ctx.Done():
			fmt.Println("Peer stopped:", context.Cause(p.ctx))
			return nil
		case <-ticker.C:
			p.heartbeat()
		}
	}
}

func (p *Peer) heartbeat() {
	var dialer net.Dialer
	conn, err := dialer.DialContext(p.ctx, "tcp", p.serverAddr)
	if err != nil {
		fmt.Println("Heartbeat dial fail:", err)
		return
	}
	defer conn.Close()

	if _, err := fmt.Fprintf(conn, "%s is alive\n", p.id); err != nil {
		fmt.Println("Heartbeat send fail:", err)
		return
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Heartbeat read fail:", err)
		return
	}

	fmt.Println("Response from tracker:", string(buf[:n]))
}

func (p *Peer) Stop(reason error) {
	p.cancel(reason)
}
