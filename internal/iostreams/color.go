package iostreams

import "github.com/muesli/termenv"

// Colorizer wraps a termenv.Output for producing styled strings.
type Colorizer struct {
	out *termenv.Output
}

func NewColorizer(out *termenv.Output) *Colorizer {
	return &Colorizer{out: out}
}

func (c *Colorizer) Bold(s string) string {
	return c.out.String(s).Bold().String()
}

func (c *Colorizer) Green(s string) string {
	return c.out.String(s).Foreground(c.out.Color("2")).String()
}

func (c *Colorizer) Red(s string) string {
	return c.out.String(s).Foreground(c.out.Color("1")).String()
}

func (c *Colorizer) Yellow(s string) string {
	return c.out.String(s).Foreground(c.out.Color("3")).String()
}

func (c *Colorizer) Gray(s string) string {
	return c.out.String(s).Foreground(c.out.Color("8")).String()
}

// Convenience package-level functions using termenv.DefaultOutput().
// These are used by code that doesn't have access to IOStreams (e.g. auth status).
// For accurate TTY-based color, prefer IOStreams.NewColorizer().

func Bold(s string) string {
	return defaultColorizer().Bold(s)
}

func Green(s string) string {
	return defaultColorizer().Green(s)
}

func Red(s string) string {
	return defaultColorizer().Red(s)
}

func Yellow(s string) string {
	return defaultColorizer().Yellow(s)
}

func Gray(s string) string {
	return defaultColorizer().Gray(s)
}

func defaultColorizer() *Colorizer {
	return &Colorizer{out: termenv.DefaultOutput()}
}
