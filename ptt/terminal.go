package ptt

// NewTerminal new terminal
func NewTerminal(in chan byte) *Terminal {
	t := &Terminal{
		in: in,
	}
	return t
}

// Terminal vt100
type Terminal struct {
	in chan byte
}
