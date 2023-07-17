package memory

type Size interface {
	uint8 | uint16 | uint32
}

type Memory interface {
	Get8(address uint32) (value uint8)
	Set8(address uint32, value uint8)
	Get16(address uint32) (value uint16)
	Set16(address uint32, value uint16)
	Get32(address uint32) (value uint32)
	Set32(address uint32, value uint32)
	GetSlice(address uint32, size uint32) (value []byte)
	SetSlice(address uint32, value []byte)
}

type Offset struct {
	offset uint32
	data   []byte
}

var _ Memory = new(Offset)

func NewOffset(offset uint32, size uint32) Offset {
	return Offset{
		offset: offset,
		data:   make([]byte, size),
	}
}

func (o Offset) Get8(address uint32) (value uint8) {
	address -= o.offset
	return o.data[address]
}

func (o Offset) Set8(address uint32, value uint8) {
	address -= o.offset
	o.data[address] = value
}

func (o Offset) Get16(address uint32) (value uint16) {
	address -= o.offset
	address &= ^uint32(1)
	value = uint16(o.data[address])
	value |= uint16(o.data[address+1]) << 8
	return
}

func (o Offset) Set16(address uint32, value uint16) {
	address -= o.offset
	address &= ^uint32(1)
	o.data[address] = uint8(value)
	o.data[address+1] = uint8(value >> 8)
}

func (o Offset) Get32(address uint32) (value uint32) {
	address -= o.offset
	address &= ^uint32(1)
	value = uint32(o.data[address])
	value |= uint32(o.data[address+1]) << 8
	value |= uint32(o.data[address+2]) << 16
	value |= uint32(o.data[address+3]) << 24
	return
}

func (o Offset) Set32(address uint32, value uint32) {
	address -= o.offset
	address &= ^uint32(1)
	o.data[address] = uint8(value)
	o.data[address+1] = uint8(value >> 8)
	o.data[address+2] = uint8(value >> 16)
	o.data[address+3] = uint8(value >> 24)
}

func (o Offset) GetSlice(address uint32, size uint32) (value []byte) {
	address -= o.offset
	return o.data[address : address+size]
}

func (o Offset) SetSlice(address uint32, value []byte) {
	address -= o.offset
	for i := uint32(0); i < uint32(len(value)); i++ {
		o.data[address+i] = value[i]
	}
}

type Map[T comparable] struct {
	keyer func(uint32) T
	data  map[T]Memory
}

var _ Memory = new(Map[int])

func NewMap[T comparable](keyer func(uint32) T) Map[T] {
	return Map[T]{
		keyer: keyer,
		data:  make(map[T]Memory),
	}
}

func (m Map[T]) Register(address uint32, sub Memory) {
	m.data[m.keyer(address)] = sub
}

func (m Map[T]) Get8(address uint32) (value uint8) {
	return m.data[m.keyer(address)].Get8(address)
}

func (m Map[T]) Set8(address uint32, value uint8) {
	m.data[m.keyer(address)].Set8(address, value)
}

func (m Map[T]) Get16(address uint32) (value uint16) {
	return m.data[m.keyer(address)].Get16(address)
}

func (m Map[T]) Set16(address uint32, value uint16) {
	m.data[m.keyer(address)].Set16(address, value)
}

func (m Map[T]) Get32(address uint32) (value uint32) {
	return m.data[m.keyer(address)].Get32(address)
}

func (m Map[T]) Set32(address uint32, value uint32) {
	k := m.keyer(address)
	v := m.data[k]
	v.Set32(address, value)
}

func (m Map[T]) GetSlice(address uint32, size uint32) (value []byte) {
	return m.data[m.keyer(address)].GetSlice(address, size)
}

func (m Map[T]) SetSlice(address uint32, value []byte) {
	m.data[m.keyer(address)].SetSlice(address, value)
}

type Mirror struct {
	Offset
}

var _ Memory = new(Mirror)

func (m Mirror) Get8(address uint32) (value uint8) {
	address -= m.offset
	address %= uint32(len(m.data))
	address += m.offset
	return m.Offset.Get8(address)
}

func (m Mirror) Set8(address uint32, value uint8) {
	address -= m.offset
	address %= uint32(len(m.data))
	address += m.offset
	m.Offset.Set8(address, value)
}

func (m Mirror) Get16(address uint32) (value uint16) {
	address -= m.offset
	address %= uint32(len(m.data))
	address += m.offset
	return m.Offset.Get16(address)
}

func (m Mirror) Set16(address uint32, value uint16) {
	address -= m.offset
	address %= uint32(len(m.data))
	address += m.offset
	m.Offset.Set16(address, value)
}

func (m Mirror) Get32(address uint32) (value uint32) {
	address -= m.offset
	address %= uint32(len(m.data))
	address += m.offset
	return m.Offset.Get32(address)
}

func (m Mirror) Set32(address uint32, value uint32) {
	address -= m.offset
	address %= uint32(len(m.data))
	address += m.offset
	m.Offset.Set32(address, value)
}

func (m Mirror) GetSlice(address uint32, size uint32) (value []byte) {
	address -= m.offset
	address %= uint32(len(m.data))
	address += m.offset
	return m.Offset.GetSlice(address, size)
}

func (m Mirror) SetSlice(address uint32, value []byte) {
	address -= m.offset
	address %= uint32(len(m.data))
	address += m.offset
	m.Offset.SetSlice(address, value)
}
