package iostreams

import (
	"fmt"
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// borderlessStyle is a clean style with no borders or separators,
// matching the previous hand-written TablePrinter output:
// two-space column padding, no lines between rows.
var borderlessStyle = table.Style{
	Name: "Borderless",
	Box: table.BoxStyle{
		PaddingLeft:  "",
		PaddingRight: "  ",
	},
	Format: table.FormatOptions{
		Header: text.FormatDefault, // callers already apply bold/uppercase
		Row:    text.FormatDefault,
	},
	Options: table.Options{
		DrawBorder:      false,
		SeparateColumns: false,
		SeparateHeader:  false,
		SeparateRows:    false,
		SeparateFooter:  false,
	},
}

type TablePrinter struct {
	out   io.Writer
	isTTY bool
	rows  [][]string
}

func NewTablePrinter(out io.Writer, isTTY bool) *TablePrinter {
	return &TablePrinter{out: out, isTTY: isTTY}
}

func (t *TablePrinter) AddRow(cols ...string) {
	t.rows = append(t.rows, cols)
}

func (t *TablePrinter) Render() error {
	if len(t.rows) == 0 {
		return nil
	}
	if !t.isTTY {
		return t.renderTSV()
	}
	return t.renderTable()
}

func (t *TablePrinter) renderTSV() error {
	for _, row := range t.rows {
		_, err := fmt.Fprintln(t.out, strings.Join(row, "\t"))
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TablePrinter) renderTable() error {
	tw := table.NewWriter()
	tw.SetStyle(borderlessStyle)

	for _, row := range t.rows {
		tableRow := make(table.Row, len(row))
		for i, col := range row {
			tableRow[i] = col
		}
		tw.AppendRow(tableRow)
	}

	// Trim trailing whitespace from each line (go-pretty applies PaddingRight to all columns).
	rendered := tw.Render()
	lines := strings.Split(rendered, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " ")
	}
	_, err := fmt.Fprintln(t.out, strings.Join(lines, "\n"))
	return err
}
