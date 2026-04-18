package progress

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ProgressBar displays download/extraction progress with speed tracking.
type ProgressBar struct {
	mu         sync.Mutex
	current    int64
	total      int64
	startTime  time.Time
	lastUpdate time.Time
	lastBytes  int64
	width      int
	label      string
	done       bool
	unit       string // "bytes" or "files"
}

// New creates a new ProgressBar with the given label and total size.
// If total is <= 0, the bar shows bytes without percentage.
func New(label string, total int64) *ProgressBar {
	return &ProgressBar{
		label:      label,
		total:      total,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
		width:      30,
		unit:       "bytes",
	}
}

// NewFiles creates a ProgressBar that tracks file count instead of bytes.
func NewFiles(label string, total int64) *ProgressBar {
	p := New(label, total)
	p.unit = "files"
	return p
}

// Update updates the current progress and renders the bar.
func (p *ProgressBar) Update(current int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.done {
		return
	}

	p.current = current
	p.Render()
}

// Render draws the progress bar to stdout.
func (p *ProgressBar) Render() {
	if p.total > 0 {
		p.renderWithPercentage()
	} else {
		p.renderBytesOnly()
	}
}

func (p *ProgressBar) renderWithPercentage() {
	percent := float64(p.current) / float64(p.total) * 100
	if percent > 100 {
		percent = 100
	}

	filled := int(percent / 100 * float64(p.width))
	if filled > p.width {
		filled = p.width
	}
	empty := p.width - filled

	bar := strings.Repeat("=", filled) + strings.Repeat(" ", empty)

	speed := p.calcSpeed()
	elapsed := time.Since(p.startTime).Truncate(time.Second)

	if p.unit == "files" {
		line := fmt.Sprintf("\r%s: [%s] %5.1f%% %d/%d files [%s]",
			p.label,
			bar,
			percent,
			p.current,
			p.total,
			elapsed,
		)
		fmt.Fprint(os.Stdout, line)
	} else {
		line := fmt.Sprintf("\r%s: [%s] %5.1f%% %s/%s (%s/s) [%s]",
			p.label,
			bar,
			percent,
			FormatSize(p.current),
			FormatSize(p.total),
			FormatSpeed(speed),
			elapsed,
		)
		fmt.Fprint(os.Stdout, line)
	}
}

func (p *ProgressBar) renderBytesOnly() {
	speed := p.calcSpeed()
	elapsed := time.Since(p.startTime).Truncate(time.Second)

	if p.unit == "files" {
		line := fmt.Sprintf("\r%s: %d files [%s]",
			p.label,
			p.current,
			elapsed,
		)
		fmt.Fprint(os.Stdout, line)
	} else {
		line := fmt.Sprintf("\r%s: %s (%s/s) [%s]",
			p.label,
			FormatSize(p.current),
			FormatSpeed(speed),
			elapsed,
		)
		fmt.Fprint(os.Stdout, line)
	}
}

func (p *ProgressBar) calcSpeed() float64 {
	now := time.Now()
	elapsed := now.Sub(p.lastUpdate).Seconds()
	if elapsed <= 0 {
		return 0
	}
	bytesDelta := float64(p.current - p.lastBytes)
	p.lastUpdate = now
	p.lastBytes = p.current
	return bytesDelta / elapsed
}

// Done finalizes the progress bar with a newline.
func (p *ProgressBar) Done() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.done = true

	if p.total > 0 {
		p.current = p.total
	}

	if p.total > 0 {
		p.renderWithPercentage()
	} else {
		p.renderBytesOnly()
	}

	fmt.Fprintln(os.Stdout)
}

// FormatSize returns a human-readable string for bytes.
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatSpeed returns a human-readable speed string.
func FormatSpeed(bytesPerSec float64) string {
	const unit = 1024
	if bytesPerSec < unit {
		return fmt.Sprintf("%.0f B", bytesPerSec)
	}
	div, exp := int64(unit), 0
	for n := bytesPerSec / unit; n >= unit; n /= unit {
		div *= int64(unit)
		exp++
	}
	return fmt.Sprintf("%.1f %cB", bytesPerSec/float64(div), "KMGTPE"[exp])
}
