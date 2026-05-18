package meta_test

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"shareDir/procotal"
)

func TestMetaSendReceive(t *testing.T) {
	// setup a temp dir with a test file
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "hello.txt")
	if err := os.WriteFile(srcPath, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	dstDir := t.TempDir()

	// start a listener that acts as the receiver
	listener, err := net.Listen("tcp", "127.0.0.1:34567")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	recvErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			recvErr <- err
			return
		}
		defer conn.Close()

		proto, err := procotal.New(conn, "meta")
		if err != nil {
			recvErr <- err
			return
		}

		if _, err := proto.ReceiveFile(dstDir); err != nil {
			recvErr <- err
			return
		}
		recvErr <- nil
	}()

	// sender dials the listener
	addr := listener.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	sproto, err := procotal.New(conn, "meta")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := sproto.SendFile(srcPath); err != nil {
		t.Fatal(err)
	}

	// wait for receiver to finish
	if err := <-recvErr; err != nil {
		t.Fatal("receive failed:", err)
	}

	// verify received file matches
	dstPath := filepath.Join(dstDir, "hello.txt")
	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello world" {
		t.Fatalf("got %q, want %q", string(got), "hello world")
	}
}
