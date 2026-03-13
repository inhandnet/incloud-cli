package iostreams

import "github.com/muesli/termenv"

var profile = termenv.ColorProfile()

func Bold(s string) string {
	return termenv.String(s).Bold().String()
}

func Green(s string) string {
	return termenv.String(s).Foreground(profile.Color("2")).String()
}

func Red(s string) string {
	return termenv.String(s).Foreground(profile.Color("1")).String()
}

func Yellow(s string) string {
	return termenv.String(s).Foreground(profile.Color("3")).String()
}

func Gray(s string) string {
	return termenv.String(s).Foreground(profile.Color("8")).String()
}
