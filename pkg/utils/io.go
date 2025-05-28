package utils

import (
	"fmt"
	"strings"
)

var (
	DEFAULT_EMPTY_MSG                = "No rows to display"
	DEFAULT_SEPARATION_LENGTH        = 2
	DEFAULT_PRINT_HEADERS_WHEN_EMPTY = false
)

type Column string

type Row []Column

type Table struct {
	headers               []Column
	rows                  []Row
	quietColIndices       []int
	emptyMsg              string
	separationLength      int
	printHeadersWhenEmpty bool
	quiet                 bool
}

func NewTable(headers ...Column) *Table {
	return &Table{
		headers:               headers,
		rows:                  []Row{},
		emptyMsg:              DEFAULT_EMPTY_MSG,
		separationLength:      DEFAULT_SEPARATION_LENGTH,
		printHeadersWhenEmpty: DEFAULT_PRINT_HEADERS_WHEN_EMPTY,
		quiet:                 false,
		quietColIndices:       []int{},
	}
}

func (t *Table) AddRow(cols ...string) {
	if len(cols) != len(t.headers) {
		panic("invalid number of columns")
	}
	row := Row{}
	for _, col := range cols {
		row = append(row, Column(col))
	}
	t.rows = append(t.rows, row)
}

func (t *Table) GetColumnWidths() []int {
	widths := []int{}
	for _, header := range t.headers {
		widths = append(widths, len(header))
	}
	for _, row := range t.rows {
		for i, col := range row {
			if len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}
	return widths
}

func (t *Table) SetQuietCols(cols ...string) {
	t.quiet = true
	for _, col := range cols {
		for i, header := range t.headers {
			if string(header) == col {
				t.quietColIndices = append(t.quietColIndices, i)
			}
		}
	}
}

func containsInt(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func (t *Table) Print() {
	widths := t.GetColumnWidths()
	format := "%-*s"
	hasRows := !t.IsEmpty()
	if !t.quiet && (hasRows || t.printHeadersWhenEmpty) {
		for i, header := range t.headers {
			value := strings.ToUpper(string(header))
			fmt.Printf(format, widths[i]+t.separationLength, value)
		}
		fmt.Println()
	}
	if hasRows {
		for _, row := range t.rows {
			for i, col := range row {
				if len(col) == 0 {
					col = "-"
				}
				if !t.quiet || containsInt(t.quietColIndices, i) {
					fmt.Printf(format, widths[i]+t.separationLength, col)
				}
			}
			fmt.Println()
		}
	} else if !t.quiet {
		fmt.Println(t.emptyMsg)
	}
}

func (t *Table) PrintSeparator() {
	if !t.quiet {
		fmt.Println()
	}
}

func (t *Table) SetEmptyMsg(msg string) {
	t.emptyMsg = msg
}

func (t *Table) SetSeparationLength(length int) {
	t.separationLength = length
}

func (t *Table) SetPrintHeadersWhenEmpty(choice bool) {
	t.printHeadersWhenEmpty = choice
}

func (t *Table) IsEmpty() bool {
	return len(t.rows) == 0
}
