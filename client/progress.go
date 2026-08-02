package client

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// UploadProgress receives progress updates for NAR package uploads. It is
// optional; clients that leave Progress nil retain the normal log output.
type UploadProgress interface {
	SetTotal(packages int)
	Start(packageName string, totalBytes uint64) UploadProgressTask
	Wait()
}

// UploadProgressTask tracks one package in a parallel upload.
type UploadProgressTask interface {
	Add(bytes uint64)
	Complete()
}

type noopProgressTask struct{}

func (noopProgressTask) Add(uint64) {}
func (noopProgressTask) Complete()  {}

// packageName removes the Nix store hash while preserving the package name.
func packageName(storePath string) string {
	base := filepath.Base(storePath)
	if len(base) > 33 && base[32] == '-' {
		return strings.TrimPrefix(base[32:], "-")
	}

	return base
}

// progressWriter counts uncompressed NAR bytes as they are serialized. For
// multipart uploads the compressor is back-pressured by the network, so this
// closely follows bytes uploaded while retaining the exact NAR size as total.
type progressWriter struct {
	contextErr func() error
	writer     io.Writer
	task       UploadProgressTask
	pending    uint64
}

const progressUpdateSize = 256 * 1024

func (w *progressWriter) Write(p []byte) (int, error) {
	if err := w.contextErr(); err != nil {
		return 0, fmt.Errorf("serializing canceled: %w", err)
	}

	n, err := w.writer.Write(p)
	if n > 0 {
		w.pending += uint64(n)
	}

	if w.pending >= progressUpdateSize {
		w.task.Add(w.pending)
		w.pending = 0
	}

	return n, err //nolint:wrapcheck // io.Writer implementations pass through underlying errors
}

func (w *progressWriter) Flush() {
	if w.pending > 0 {
		w.task.Add(w.pending)
		w.pending = 0
	}
}
