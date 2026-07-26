package lib

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"
)

// RotatingFileWriter writes to a file and rotates when size exceeds maxBytes.
// Old file is truncated; production would archive/compress.
type RotatingFileWriter struct {
	path     string
	maxBytes int64
	file     *os.File
	writer   *bufio.Writer
	size     int64
	mu       sync.Mutex
}

func NewRotatingFileWriter(path string, maxBytes int64) (*RotatingFileWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		return &RotatingFileWriter{
			path:     path,
			maxBytes: maxBytes,
			file:     f,
			writer:   bufio.NewWriter(f),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &RotatingFileWriter{
		path:     path,
		maxBytes: maxBytes,
		file:     f,
		writer:   bufio.NewWriter(f),
		size:     fi.Size(),
	}, nil
}

func (w *RotatingFileWriter) WriteLine(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.size+int64(len(data))+1 > w.maxBytes {
		if err := w.rotate(); err != nil {
			return err
		}
	}
	n, err := w.writer.Write(data)
	if err != nil {
		return err
	}
	w.size += int64(n)
	if err := w.writer.WriteByte('\n'); err != nil {
		return err
	}
	w.size++
	return w.writer.Flush()
}

func (w *RotatingFileWriter) rotate() error {
	w.writer.Flush()
	w.file.Close()
	f, err := os.Create(w.path)
	if err != nil {
		return err
	}
	w.file = f
	w.writer = bufio.NewWriter(f)
	w.size = 0
	return nil
}

func (w *RotatingFileWriter) Tail(n int) ([][]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writer.Flush()

	f, err := os.Open(w.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines [][]byte
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, []byte(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if n > 0 && len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}

func (w *RotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writer.Flush()
	return w.file.Close()
}
