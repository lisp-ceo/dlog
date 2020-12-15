package log

import (
	"bufio"
	"os"
	"sync"
)

// store persists log segments
type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

// newStore creates a new log.Store
func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		buf:  bufio.NewWriter(f),
		size: size,
	}, nil
}
