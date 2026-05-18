package procotal

import (
	"fmt"
	"net"
	"shareDir/procotal/meta"
)

// file transport procotal
type Protocal interface {
	SendFile(path string) (int, error)
	ReceiveFile(dir string) (int, error)
}

func New(conn net.Conn, way string) (Protocal, error) {
	switch way {
	case "meta":
		return meta.New(conn)
	default:
		return nil, fmt.Errorf("Unexpected protocal")
	}
}
