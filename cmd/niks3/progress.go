package main

import (
	"fmt"
	"io"
	"math"
	"sync"
	"sync/atomic"

	"github.com/Mic92/niks3/client"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

type progressDisplay struct {
	progress  *mpb.Progress
	summary   *mpb.Bar
	remaining atomic.Int64
}

func newProgressDisplay(output io.Writer) *progressDisplay {
	return &progressDisplay{
		progress: mpb.New(
			mpb.WithOutput(output),
			// --output=progress is also useful when the caller explicitly
			// requests it through a terminal wrapper not detected by mpb.
			mpb.WithAutoRefresh(),
		),
	}
}

func (d *progressDisplay) SetTotal(packages int) {
	d.remaining.Store(int64(packages))

	if packages == 0 {
		return
	}

	d.summary = d.progress.AddBar(
		int64(packages),
		mpb.PrependDecorators(decor.Name("Packages ")),
		mpb.AppendDecorators(
			decor.CountersNoUnit("%d/%d"),
			decor.Any(func(decor.Statistics) string {
				return fmt.Sprintf(" (%d left)", d.remaining.Load())
			}),
		),
	)
}

func progressInt64(bytes uint64) int64 {
	if bytes > math.MaxInt64 {
		return math.MaxInt64
	}

	return int64(bytes)
}

//nolint:ireturn // client.UploadProgress requires its task interface here
func (d *progressDisplay) Start(packageName string, totalBytes uint64) client.UploadProgressTask {
	total := progressInt64(totalBytes)
	bar := d.progress.AddBar(
		total,
		mpb.PrependDecorators(decor.Name(packageName+" ", decor.WCSyncSpaceR)),
		mpb.AppendDecorators(
			decor.CountersKibiByte(" % .1f / % .1f "),
			decor.Percentage(),
		),
		mpb.BarRemoveOnComplete(),
	)

	return &progressTask{display: d, bar: bar, total: total}
}

func (d *progressDisplay) Wait() {
	// An upload can fail while pending objects are still being prepared. In
	// that case some expected package bars never start, so remove the summary
	// rather than waiting forever for increments that cannot arrive.
	if d.summary != nil && d.remaining.Load() > 0 {
		d.summary.Abort(true)
	}

	d.progress.Wait()
}

type progressTask struct {
	display *progressDisplay
	bar     *mpb.Bar
	total   int64
	once    sync.Once
}

func (t *progressTask) Add(bytes uint64) {
	t.bar.IncrInt64(progressInt64(bytes))
}

func (t *progressTask) Complete() {
	t.once.Do(func() {
		t.bar.SetTotal(t.total, true)
		t.display.remaining.Add(-1)

		if t.display.summary != nil {
			t.display.summary.Increment()
		}
	})
}
