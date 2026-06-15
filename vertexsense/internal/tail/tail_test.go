package tail

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func collect(t *testing.T, ch <-chan string, n int, d time.Duration) []string {
	t.Helper()
	var got []string
	deadline := time.After(d)
	for len(got) < n {
		select {
		case l := <-ch:
			got = append(got, l)
		case <-deadline:
			return got
		}
	}
	return got
}

func TestFollowAppendAndTruncate(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "log")
	if err := os.WriteFile(p, []byte("pre-existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan string, 16)
	go Follow(ctx, p, false, 30*time.Millisecond, ch) // fromStart=false: skip pre-existing

	time.Sleep(80 * time.Millisecond)
	f, _ := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString("line-a\nline-b\n")
	f.Close()

	got := collect(t, ch, 2, time.Second)
	if len(got) != 2 || got[0] != "line-a" || got[1] != "line-b" {
		t.Fatalf("append: got %v", got)
	}

	// Truncate in place, write fresh content -> tailer must reset and read it.
	os.WriteFile(p, []byte("after-trunc\n"), 0o644)
	got = collect(t, ch, 1, time.Second)
	if len(got) != 1 || got[0] != "after-trunc" {
		t.Fatalf("truncate: got %v", got)
	}
}

func TestFollowWaitsForCreation(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "later.log")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan string, 8)
	go Follow(ctx, p, true, 30*time.Millisecond, ch)

	time.Sleep(80 * time.Millisecond) // file does not exist yet
	os.WriteFile(p, []byte("born\n"), 0o644)

	got := collect(t, ch, 1, time.Second)
	if len(got) != 1 || got[0] != "born" {
		t.Fatalf("creation: got %v", got)
	}
}
