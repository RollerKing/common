package debug

import (
	gotable "github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	"os"
)

type Table interface {
	SetHeader(v ...interface{}) Table
	AddRow(v ...interface{}) Table
	Render()
}

type table struct {
	tw gotable.Writer
}

func NewTable() Table {
	t := &table{
		tw: gotable.NewWriter(),
	}
	style := gotable.StyleColoredBright
	style.Format.Header = text.FormatDefault
	t.tw.SetStyle(style)
	t.tw.SetOutputMirror(os.Stdout)
	return t
}

func (t *table) SetHeader(v ...interface{}) Table {
	t.tw.AppendHeader(cellsToRow(v...))
	return t
}

func (t *table) AddRow(cells ...interface{}) Table {
	t.tw.AppendRows([]gotable.Row{cellsToRow(cells...)})
	return t
}

func cellsToRow(list ...interface{}) gotable.Row {
	return gotable.Row(list)
}

func (t *table) Render() {
	t.tw.Render()
}
