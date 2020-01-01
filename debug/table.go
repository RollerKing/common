package debug

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

type Table interface {
	SetHeader(v ...string) Table
	AddRow(v ...interface{}) Table
	Render()
}

type table struct {
	tw *tablewriter.Table
}

func NewTable() Table {
	return &table{
		tw: tablewriter.NewWriter(os.Stdout),
	}
}

func (t *table) SetHeader(v ...string) Table {
	t.tw.SetHeader(v)
	return t
}

func (t *table) AddRow(cells ...interface{}) Table {
	row := make([]string, len(cells))
	for i, val := range cells {
		row[i] = fmt.Sprint(val)
	}
	t.tw.Append(row)
	return t
}

func (t *table) Render() {
	t.tw.Render()
}
