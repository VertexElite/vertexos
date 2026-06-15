// Package tail follows a growing log file and emits complete new lines. It is
// poll-based (not inotify) on purpose: the sensor outputs live on a variety of
// filesystems (tmpfs, 9p under WSL, overlay in the live ISO) where inotify is
// unreliable, and polling a handful of log files costs nothing.
//
// It tolerates the realities of log files: the file may not exist yet, may be
// truncated in place, or rotated (replaced by a new inode). Partial trailing
// lines are buffered until their newline arrives.
package tail

import (
	"context"
	"errors"
	"io"
	"os"
	"time"
)

// Follow tails path, sending each complete line (without the trailing newline)
// to out until ctx is cancelled. If fromStart is false, pre-existing content is
// skipped and only lines appended after start are emitted (tail -f semantics).
func Follow(ctx context.Context, path string, fromStart bool, interval time.Duration, out chan<- string) {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	var (
		f      *os.File
		offset int64
		buf    []byte
	)
	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()

	open := func() {
		nf, err := os.Open(path)
		if err != nil {
			return
		}
		f = nf
		if fromStart {
			offset = 0
		} else if fi, err := f.Stat(); err == nil {
			offset = fi.Size() // start at current end
		}
		_, _ = f.Seek(offset, io.SeekStart)
	}

	read := func() {
		fi, err := f.Stat()
		if err != nil {
			_ = f.Close()
			f = nil
			return
		}
		// Truncated/rotated in place: size shrank below our offset -> restart.
		if fi.Size() < offset {
			offset = 0
			buf = nil
			_, _ = f.Seek(0, io.SeekStart)
		}
		chunk := make([]byte, 64*1024)
		for {
			n, err := f.Read(chunk)
			if n > 0 {
				offset += int64(n)
				buf = append(buf, chunk[:n]...)
				buf = emitLines(buf, out)
			}
			if errors.Is(err, io.EOF) || n == 0 {
				break
			}
			if err != nil {
				break
			}
		}
	}

	// Detect rotation: same path now points at a different (smaller/newer) file.
	rotated := func() bool {
		if f == nil {
			return false
		}
		cur, err := f.Stat()
		if err != nil {
			return true
		}
		onDisk, err := os.Stat(path)
		if err != nil {
			return false // path gone for now; keep draining current handle
		}
		return !os.SameFile(cur, onDisk)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if f == nil {
				open()
				continue
			}
			read()
			if rotated() {
				_ = f.Close()
				f = nil
				offset = 0
				buf = nil
				open()
				if f != nil {
					offset = 0
					_, _ = f.Seek(0, io.SeekStart) // read rotated file from start
					read()
				}
			}
		}
	}
}

// emitLines sends every complete line in buf to out and returns the remaining
// partial (unterminated) tail.
func emitLines(buf []byte, out chan<- string) []byte {
	for {
		i := indexByte(buf, '\n')
		if i < 0 {
			return buf
		}
		line := string(trimCR(buf[:i]))
		if line != "" {
			out <- line
		}
		buf = buf[i+1:]
	}
}

func indexByte(b []byte, c byte) int {
	for i := range b {
		if b[i] == c {
			return i
		}
	}
	return -1
}

func trimCR(b []byte) []byte {
	if len(b) > 0 && b[len(b)-1] == '\r' {
		return b[:len(b)-1]
	}
	return b
}
