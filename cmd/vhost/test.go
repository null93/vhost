package main

import (
	"github.com/jetrails/proposal-nginx/pkg/utils"
)

func main() {
	tbl := utils.NewTable("ID", "Name", "Age")
	tbl.AddRow("13244", "John", "20")
	tbl.PrintSeparator()
	tbl.Print()
	tbl.PrintSeparator()
}
