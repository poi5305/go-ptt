package main

func NewB2U() *B2U {
	return &B2U{
		buf:   make([]byte, 1024, 1024),
		p:     0,
		state: 0,
	}
}

type B2U struct {
	buf   []byte
	p     int
	state int
}

func (b *B2U) Write(bs []byte) (int, error) {
	b.buf = append(b.buf, bs...)
	return len(bs), nil
}

func (b *B2U) Read(bs []byte) (int, error) {
	b.p = 0
	for ; b.p < len(b.buf); b.p++ {
		c := b.buf[b.p]
		switch c {
		case IAC:

		}
	}
}
