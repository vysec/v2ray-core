// Package alloc provides a light-weight memory allocation mechanism.
package alloc

import (
	"io"
)

const (
	defaultOffset = 16
)

type BytesWriter func([]byte) int

// Buffer is a recyclable allocation of a byte array. Buffer.Release() recycles
// the buffer into an internal buffer pool, in order to recreate a buffer more
// quickly.
type Buffer struct {
	head []byte
	pool Pool

	start int
	end   int
}

func CreateBuffer(container []byte, parent Pool) *Buffer {
	b := new(Buffer)
	b.head = container
	b.pool = parent
	b.start = defaultOffset
	b.end = defaultOffset
	return b
}

// Release recycles the buffer into an internal buffer pool.
func (b *Buffer) Release() {
	if b == nil || b.head == nil {
		return
	}
	if b.pool != nil {
		b.pool.Free(b)
	}
	b.head = nil
	b.pool = nil
}

// Clear clears the content of the buffer, results an empty buffer with
// Len() = 0.
func (b *Buffer) Clear() {
	b.start = defaultOffset
	b.end = defaultOffset
}

// Reset resets this Buffer into its original state.
func (b *Buffer) Reset() {
	b.start = defaultOffset
	b.end = len(b.head)
}

// AppendBytes appends one or more bytes to the end of the buffer.
func (b *Buffer) AppendBytes(bytes ...byte) {
	b.Append(bytes)
}

// Append appends a byte array to the end of the buffer.
func (b *Buffer) Append(data []byte) {
	nBytes := copy(b.head[b.end:], data)
	b.end += nBytes
}

func (b *Buffer) AppendFunc(writer BytesWriter) {
	nBytes := writer(b.head[b.end:])
	b.end += nBytes
}

// Prepend prepends bytes in front of the buffer. Caller must ensure total bytes prepended is
// no more than 16 bytes.
func (b *Buffer) Prepend(data []byte) {
	b.SliceBack(len(data))
	copy(b.head[b.start:], data)
}

func (b *Buffer) PrependBytes(data ...byte) {
	b.Prepend(data)
}

func (b *Buffer) PrependFunc(offset int, writer BytesWriter) {
	b.SliceBack(offset)
	writer(b.head[b.start:])
}

func (b *Buffer) Byte(index int) byte {
	return b.head[b.start+index]
}

func (b *Buffer) SetByte(index int, value byte) {
	b.head[b.start+index] = value
}

// Bytes returns the content bytes of this Buffer.
func (b *Buffer) Bytes() []byte {
	return b.head[b.start:b.end]
}

func (b *Buffer) BytesRange(from, to int) []byte {
	if from < 0 {
		from += b.Len()
	}
	if to < 0 {
		to += b.Len()
	}
	return b.head[b.start+from : b.start+to]
}

func (b *Buffer) BytesFrom(from int) []byte {
	if from < 0 {
		from += b.Len()
	}
	return b.head[b.start+from : b.end]
}

func (b *Buffer) BytesTo(to int) []byte {
	if to < 0 {
		to += b.Len()
	}
	return b.head[b.start : b.start+to]
}

// Slice cuts the buffer at the given position.
func (b *Buffer) Slice(from, to int) {
	if from < 0 {
		from += b.Len()
	}
	if to < 0 {
		to += b.Len()
	}
	if to < from {
		panic("Invalid slice")
	}
	b.end = b.start + to
	b.start += from
}

// SliceFrom cuts the buffer at the given position.
func (b *Buffer) SliceFrom(from int) {
	if from < 0 {
		from += b.Len()
	}
	b.start += from
}

// SliceBack extends the Buffer to its front by offset bytes.
// Caller must ensure cumulated offset is no more than 16.
func (b *Buffer) SliceBack(offset int) {
	b.start -= offset
	if b.start < 0 {
		panic("Negative buffer offset.")
	}
}

// Len returns the length of the buffer content.
func (b *Buffer) Len() int {
	if b == nil {
		return 0
	}
	return b.end - b.start
}

func (b *Buffer) IsEmpty() bool {
	return b.Len() == 0
}

// IsFull returns true if the buffer has no more room to grow.
func (b *Buffer) IsFull() bool {
	return b.end == len(b.head)
}

// Write implements Write method in io.Writer.
func (b *Buffer) Write(data []byte) (int, error) {
	nBytes := copy(b.head[b.end:], data)
	b.end += nBytes
	return nBytes, nil
}

// Read implements io.Reader.Read().
func (b *Buffer) Read(data []byte) (int, error) {
	if b.Len() == 0 {
		return 0, io.EOF
	}
	nBytes := copy(data, b.head[b.start:b.end])
	if nBytes == b.Len() {
		b.Clear()
	} else {
		b.start += nBytes
	}
	return nBytes, nil
}

func (b *Buffer) FillFrom(reader io.Reader) (int, error) {
	nBytes, err := reader.Read(b.head[b.end:])
	b.end += nBytes
	return nBytes, err
}

func (b *Buffer) FillFullFrom(reader io.Reader, amount int) (int, error) {
	nBytes, err := io.ReadFull(reader, b.head[b.end:b.end+amount])
	b.end += nBytes
	return nBytes, err
}

func (b *Buffer) String() string {
	return string(b.head[b.start:b.end])
}

// NewBuffer creates a Buffer with 8K bytes of arbitrary content.
func NewBuffer() *Buffer {
	return mediumPool.Allocate()
}

func NewSmallBuffer() *Buffer {
	return smallPool.Allocate()
}

func NewLocalBuffer(size int) *Buffer {
	return CreateBuffer(make([]byte, size), nil)
}
