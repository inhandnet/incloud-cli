package iostreams

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
)

type IOStreams struct {
	In       io.Reader
	Out      io.Writer
	ErrOut   io.Writer
	outIsTTY bool

	// termOut is the termenv output tied to Out for color rendering.
	termOut *termenv.Output
}

func (s *IOStreams) IsStdoutTTY() bool {
	return s.outIsTTY
}

// TermOutput returns the termenv.Output for color rendering.
func (s *IOStreams) TermOutput() *termenv.Output {
	return s.termOut
}

func System() *IOStreams {
	out := os.Stdout
	isTTY := isTerminal(out)

	// Create termenv output explicitly from stdout.
	// If not a TTY, force Ascii profile (no color).
	tOut := termenv.NewOutput(out)
	if !isTTY {
		tOut = termenv.NewOutput(out, termenv.WithProfile(termenv.Ascii))
	}

	return &IOStreams{
		In:       os.Stdin,
		Out:      out,
		ErrOut:   os.Stderr,
		outIsTTY: isTTY,
		termOut:  tOut,
	}
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
