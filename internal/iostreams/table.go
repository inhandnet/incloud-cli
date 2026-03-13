package iostreams

import (
	"fmt"
	"io"
	"strings"
)

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
	// calculate column widths
	widths := make([]int, len(t.rows[0]))
	for _, row := range t.rows {
		for i, col := range row {
			if i < len(widths) && len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	for _, row := range t.rows {
		parts := make([]string, len(row))
		for i, col := range row {
			if i < len(widths)-1 {
				parts[i] = fmt.Sprintf("%-*s", widths[i], col)
			} else {
				parts[i] = col
			}
		}
		_, err := fmt.Fprintln(t.out, strings.Join(parts, "  "))
		if err != nil {
			return err
		}
	}
	return nil
}
