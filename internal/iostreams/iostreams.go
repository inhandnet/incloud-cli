package iostreams

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

type IOStreams struct {
	In       io.Reader
	Out      io.Writer
	ErrOut   io.Writer
	outIsTTY bool
}

func (s *IOStreams) IsStdoutTTY() bool {
	return s.outIsTTY
}

func System() *IOStreams {
	return &IOStreams{
		In:       os.Stdin,
		Out:      os.Stdout,
		ErrOut:   os.Stderr,
		outIsTTY: isTerminal(os.Stdout),
	}
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
