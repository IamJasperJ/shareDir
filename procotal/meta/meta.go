package meta

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const chunkSize = 1024

// meta sends file metadata (name + size) over a TCP connection,
// then streams the file content in 1024-byte chunks.
type meta struct {
	conn net.Conn
}

func New(conn net.Conn) (*meta, error) {
	return &meta{conn: conn}, nil
}

// SendFile sends a single file to the remote peer.
// Wire format:  filename\r\n  filesize\r\n  <chunks>
func (m *meta) SendFile(path string) (int, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("stat %s: %w", path, err)
	}

	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	// header: filename
	n1, err := fmt.Fprintf(m.conn, "%s\r\n", filepath.Base(path))
	if err != nil {
		return n1, fmt.Errorf("send filename: %w", err)
	}

	// header: filesize
	n2, err := fmt.Fprintf(m.conn, "%d\r\n", fi.Size())
	if err != nil {
		return n1 + n2, fmt.Errorf("send filesize: %w", err)
	}

	// body: file content in 1024-byte chunks
	n3, err := sendChunks(m.conn, file)
	if err != nil {
		return n1 + n2 + n3, err
	}

	return n1 + n2 + n3, nil
}

// ReceiveFile reads a file from the remote peer and saves it under dir.
// Skips the transfer if the local file already has the same size.
func (m *meta) ReceiveFile(dir string) (int, error) {
	r := bufio.NewReader(m.conn)

	nameLine, err := r.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("read filename: %w", err)
	}
	name := strings.TrimRight(nameLine, "\r\n")

	sizeLine, err := r.ReadString('\n')
	if err != nil {
		return len(nameLine), fmt.Errorf("read filesize: %w", err)
	}
	sizeStr := strings.TrimRight(sizeLine, "\r\n")

	fileSize, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return len(nameLine) + len(sizeLine), fmt.Errorf("parse filesize %q: %w", sizeStr, err)
	}

	headerBytes := len(nameLine) + len(sizeLine)
	localPath := filepath.Join(dir, name)
	fi, err := os.Stat(localPath)

	// already have the file — drain and skip
	if err == nil && fi.Size() == fileSize {
		n, drainErr := io.Copy(io.Discard, io.LimitReader(r, fileSize))
		return headerBytes + int(n), drainErr
	}

	// missing or stale — overwrite
	file, err := os.Create(localPath)
	if err != nil {
		return headerBytes, fmt.Errorf("create %s: %w", localPath, err)
	}
	defer file.Close()

	n3, err := receiveChunks(file, r, fileSize)
	if err != nil {
		return headerBytes + n3, err
	}

	return headerBytes + n3, nil
}

// sendChunks reads from r in 1024-byte blocks and writes each block to w.
func sendChunks(w io.Writer, r io.Reader) (int, error) {
	buf := make([]byte, chunkSize)
	total := 0
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return total, fmt.Errorf("send chunk: %w", werr)
			}
			total += n
		}
		if err != nil {
			if err == io.EOF {
				return total, nil
			}
			return total, fmt.Errorf("read file: %w", err)
		}
	}
}

// receiveChunks reads exactly remain bytes from r in 1024-byte blocks and writes to w.
func receiveChunks(w io.Writer, r io.Reader, remain int64) (int, error) {
	buf := make([]byte, chunkSize)
	total := 0
	for remain > 0 {
		limit := int64(chunkSize)
		if limit > remain {
			limit = remain
		}
		n, err := io.ReadFull(r, buf[:limit])
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return total, fmt.Errorf("write chunk: %w", werr)
			}
			remain -= int64(n)
			total += n
		}
		if err != nil {
			if err == io.EOF {
				return total, nil
			}
			return total, fmt.Errorf("read chunk: %w", err)
		}
	}
	return total, nil
}
