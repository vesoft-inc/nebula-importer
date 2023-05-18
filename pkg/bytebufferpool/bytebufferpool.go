package bytebufferpool

import (
	bbp "github.com/valyala/bytebufferpool"
)

type (
	ByteBuffer bbp.ByteBuffer
)

func Get() *ByteBuffer {
	return (*ByteBuffer)(bbp.Get())
}

func Put(b *ByteBuffer) {
	bbp.Put((*bbp.ByteBuffer)(b))
}

func (b *ByteBuffer) Len() int {
	return (*bbp.ByteBuffer)(b).Len()
}

func (b *ByteBuffer) Bytes() []byte {
	return (*bbp.ByteBuffer)(b).Bytes()
}
func (b *ByteBuffer) Write(p []byte) (int, error) {
	return (*bbp.ByteBuffer)(b).Write(p)
}

func (b *ByteBuffer) WriteString(s string) (int, error) {
	return (*bbp.ByteBuffer)(b).WriteString(s)
}

func (b *ByteBuffer) Set(p []byte) {
	(*bbp.ByteBuffer)(b).Set(p)
}

func (b *ByteBuffer) SetString(s string) {
	(*bbp.ByteBuffer)(b).SetString(s)
}

func (b *ByteBuffer) String() string {
	return (*bbp.ByteBuffer)(b).String()
}

func (b *ByteBuffer) Reset() {
	(*bbp.ByteBuffer)(b).Reset()
}

func (b *ByteBuffer) WriteStringSlice(elems []string, sep string) (n int, err error) {
	switch len(elems) {
	case 0:
		return 0, nil
	case 1:
		return b.WriteString(elems[0])
	}
	return b.writeStringSliceSlow(elems, sep)
}

func (b *ByteBuffer) writeStringSliceSlow(elems []string, sep string) (int, error) {
	n, _ := b.WriteString(elems[0])
	for _, s := range elems[1:] {
		n1, _ := b.WriteString(sep)
		n += n1
		n1, _ = b.WriteString(s)
		n += n1
	}
	return n, nil
}
