package ptt

// NewTerminal new terminal
func NewTerminal(in chan rune) *Terminal {
	t := &Terminal{
		in: in,
	}
	return t
}

// Terminal vt100
type Terminal struct {
	in chan rune
}
