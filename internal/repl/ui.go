// internal/repl/ui.go
package repl

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	gray   = color.New(color.FgHiBlack).SprintFunc()
)

// PrintSuccess 打印成功信息
func PrintSuccess(format string, args ...interface{}) {
	printStyled(green("✓"), format, args...)
}

// PrintError 打印错误信息
func PrintError(format string, args ...interface{}) {
	printStyled(red("✗"), format, args...)
}

// PrintWarning 打印警告信息
func PrintWarning(format string, args ...interface{}) {
	printStyled(yellow("⚠"), format, args...)
}

// PrintInfo 打印提示信息
func PrintInfo(format string, args ...interface{}) {
	printStyled(gray("●"), format, args...)
}

// PrintCurrent 打印当前状态
func PrintCurrent(format string, args ...interface{}) {
	printStyled(blue("→"), format, args...)
}

func printStyled(prefix, format string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("%s %s\n", prefix, fmt.Sprintf(format, args...))
	} else {
		fmt.Printf("%s %s\n", prefix, format)
	}
}

// NewTable 创建新表格
func NewTable() *tablewriter.Table {
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithRowAlignment(tw.AlignLeft),
		tablewriter.WithBorders(tw.Border{
			Left:   tw.On,
			Right:  tw.On,
			Top:    tw.On,
			Bottom: tw.On,
		}),
		tablewriter.WithTrimSpace(tw.On),
	)
	return table
}
