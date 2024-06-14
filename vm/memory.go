package vm

import "errors"

var ErrOutOfMemory = errors.New("out of memory")

// Memory 是一个封装了内存操作的结构体，其中 Cell 表示内存的字节数组。
type Memory struct {
	Cell []byte
}

// NewMemory 创建并返回一个新的 Memory 实例。
func NewMemory(cell []byte) *Memory {
	mem := &Memory{
		Cell: make([]byte, 32),
	}
	copy(mem.Cell, cell)
	return mem
}

// WillIncrease 计算在给定偏移量和大小后，内存是否需要增长，返回新的偏移量、大小和增长量。
func (m *Memory) WillIncrease(offset uint64, size uint64) (o uint64, s uint64, i uint64, err error) {
	mLen := uint64(len(m.Cell)) // 当前内存大小
	bound := offset + size      // 从偏移量构造新的大整数边界

	if mLen < bound {
		i = bound - mLen // 计算需要增加的内存量
	}

	return offset, size, i, nil
}

// Malloc 在指定偏移量和大小的基础上分配内存，如果当前内存不足，则增加内存。
func (m *Memory) Malloc(offset uint64, size uint64) []byte {
	mLen := uint64(len(m.Cell))
	bound := offset + size
	if mLen < bound {
		newMem := make([]byte, bound-mLen) // 创建新的内存块以扩充内存
		m.Cell = append(m.Cell, newMem...)
	}

	return m.Cell[offset:bound] // 返回新分配的内存块
}

// Map 返回从指定偏移量开始的指定长度的内存片段，如果超出范围，返回错误。
func (m *Memory) Map(offset uint64, length uint64) ([]byte, error) {
	if offset+length > uint64(len(m.Cell)) {
		return nil, ErrOutOfMemory
	}

	return m.Cell[offset : offset+length], nil
}

// Store 将数据存储到指定偏移量的内存中，如果数据长度和偏移量的总和超出内存范围，则返回错误。
func (m *Memory) Store(offset uint64, data []byte) error {
	dLen := uint64(len(data))
	if dLen+offset > uint64(len(m.Cell)) {
		return ErrOutOfMemory
	}

	copy(m.Cell[offset:offset+dLen], data)
	return nil
}

// StoreNBytes 将指定数量的字节从数据中存储到指定偏移量的内存中，如果偏移量和数量超出内存范围，返回错误。
func (m *Memory) StoreNBytes(offset uint64, n uint64, data []byte) error {
	if offset+n > uint64(len(m.Cell)) {
		return ErrOutOfMemory
	}

	copy(m.Cell[offset:offset+n], data)
	return nil
}

// Set 将单个字节存储到指定索引的内存中，如果索引超出内存范围，返回错误。
func (m *Memory) Set(idx uint64, data byte) error {
	if idx > uint64(len(m.Cell))-1 {
		return ErrOutOfMemory
	}

	m.Cell[idx] = data
	return nil
}

// Copy 创建并返回从指定偏移量开始的指定长度的内存复制，如果范围超出内存大小，返回错误。
func (m *Memory) Copy(offset uint64, length uint64) ([]byte, error) {
	if offset+length > uint64(len(m.Cell)) {
		return nil, ErrOutOfMemory
	}

	ret := make([]byte, length, length)
	if length == 0 {
		return ret, nil
	}

	copy(ret, m.Cell[offset:offset+length])
	return ret, nil
}

// Size 返回当前内存的字节大小。
func (m *Memory) Size() int {
	return len(m.Cell)
}

// All 返回整个内存的副本。
func (m *Memory) All() []byte {
	return m.Cell
}
