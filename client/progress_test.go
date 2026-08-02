package client //nolint:testpackage // exercises unexported progress byte accounting

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

type recordingProgressTask struct {
	bytes uint64
}

func (t *recordingProgressTask) Add(n uint64) { t.bytes += n }
func (t *recordingProgressTask) Complete()    {}

func TestPackageName(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"/nix/store/0123456789abcdfghijklmnpqrsvwxyz-niks3-1.0": "niks3-1.0",
		"relative-name": "relative-name",
	}

	for input, expected := range tests {
		if actual := packageName(input); actual != expected {
			t.Errorf("packageName(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestProgressWriterStopsOnCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var destination bytes.Buffer

	writer := &progressWriter{contextErr: ctx.Err, writer: &destination, task: &recordingProgressTask{}}
	if _, err := writer.Write([]byte("data")); !errors.Is(err, context.Canceled) {
		t.Fatalf("Write() error = %v, want context.Canceled", err)
	}

	if destination.Len() != 0 {
		t.Fatal("canceled writer wrote data")
	}
}

func TestProgressWriterReportsAllBytes(t *testing.T) {
	t.Parallel()

	var destination bytes.Buffer

	task := &recordingProgressTask{}
	writer := &progressWriter{contextErr: context.Background().Err, writer: &destination, task: task}
	data := bytes.Repeat([]byte("x"), progressUpdateSize+17)

	if _, err := writer.Write(data); err != nil {
		t.Fatal(err)
	}

	writer.Flush()

	if task.bytes != uint64(len(data)) {
		t.Fatalf("reported %d bytes, want %d", task.bytes, len(data))
	}

	if !bytes.Equal(destination.Bytes(), data) {
		t.Fatal("progress writer changed the data")
	}
}
