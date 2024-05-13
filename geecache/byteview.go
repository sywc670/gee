package geecache

type ByteView struct {
	b []byte
}

func (bv ByteView) Len() int {
	return len(bv.b)
}

func (bv ByteView) String() string {
	return string(bv.b)
}

func (bv ByteView) ByteSlice() []byte {
	return clonebytes(bv.b)
}

func clonebytes(b []byte) []byte {
	nb := make([]byte, len(b))
	copy(nb, b)
	return nb
}
