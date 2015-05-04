package bitfield

type Bitfield struct {
	bytes  []byte
	length int // in bits
}

func New(length int) *Bitfield {
	b := &Bitfield{}
	b.length = length
	b.bytes = make([]byte, (b.length+7)/8)
	return b
}

func (b *Bitfield) Length() int {
	return b.length
}

func (b *Bitfield) Get(index int) int {
	if index > b.length-1 {
		panic("out of bounds")
	}

	i := index / 8
	shiftAmt := uint(7 - index%8)

	return int((b.bytes[i] >> shiftAmt) & 0x1)
}

func (b *Bitfield) Set(index, value int) {
	if index > b.length-1 {
		panic("out of bounds")
	}

	i := index / 8
	pos := 8 - uint(index%8)
	if value == 0 {
		b.bytes[i] &= (1 << (pos - 1)) - 1
	} else {
		b.bytes[i] |= 1 << (pos - 1)
	}
}

func (b *Bitfield) Bytes() []byte {
	bytes := make([]byte, len(b.bytes))
	copy(bytes, b.bytes)
	return bytes
}

func (b *Bitfield) SetBytes(bytes []byte) {
	bytesToCopy := b.length / 8

	copy(b.bytes, bytes[:bytesToCopy])
}
