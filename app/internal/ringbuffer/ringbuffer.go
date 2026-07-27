package ringbuffer

import "sync"

type RingBuffer struct {
	mu sync.RWMutex

	lines []string

	head int
	size int
}

func New(capacity int) *RingBuffer {
	return &RingBuffer{
		lines: make([]string, capacity),
	}
}

func (r *RingBuffer) Add(line string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lines[r.head] = line

	r.head = (r.head + 1) % len(r.lines)

	if r.size < len(r.lines) {
		r.size++
	}
}

func (r *RingBuffer) Snapshot() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0, r.size)

	if r.size < len(r.lines) {
		return append(result, r.lines[:r.size]...)
	}

	result = append(result, r.lines[r.head:]...)
	result = append(result, r.lines[:r.head]...)

	return result
}
