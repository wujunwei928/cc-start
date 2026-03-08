// internal/repl/output.go
package repl

import (
	"strings"
)

// OutputLine 输出行
type OutputLine struct {
	Content string
	Type    OutputType
}

// OutputType 输出类型
type OutputType int

const (
	OutputNormal OutputType = iota
	OutputSuccess
	OutputError
	OutputWarning
	OutputInfo
)

// OutputBuffer 输出缓冲区
type OutputBuffer struct {
	lines   []OutputLine
	maxSize int
}

// NewOutputBuffer 创建输出缓冲区
func NewOutputBuffer(maxSize int) *OutputBuffer {
	return &OutputBuffer{
		lines:   make([]OutputLine, 0),
		maxSize: maxSize,
	}
}

// Write 写入普通输出
func (b *OutputBuffer) Write(content string) {
	b.writeLine(content, OutputNormal)
}

// WriteSuccess 写入成功输出
func (b *OutputBuffer) WriteSuccess(content string) {
	b.writeLine(content, OutputSuccess)
}

// WriteError 写入错误输出
func (b *OutputBuffer) WriteError(content string) {
	b.writeLine(content, OutputError)
}

// WriteWarning 写入警告输出
func (b *OutputBuffer) WriteWarning(content string) {
	b.writeLine(content, OutputWarning)
}

// WriteInfo 写入信息输出
func (b *OutputBuffer) WriteInfo(content string) {
	b.writeLine(content, OutputInfo)
}

func (b *OutputBuffer) writeLine(content string, t OutputType) {
	// 按换行分割
	for _, line := range strings.Split(content, "\n") {
		b.lines = append(b.lines, OutputLine{Content: line, Type: t})
	}

	// 超过最大行数时裁剪
	if len(b.lines) > b.maxSize {
		b.lines = b.lines[len(b.lines)-b.maxSize:]
	}
}

// Clear 清空缓冲区
func (b *OutputBuffer) Clear() {
	b.lines = make([]OutputLine, 0)
}

// Lines 获取所有行
func (b *OutputBuffer) Lines() []OutputLine {
	return b.lines
}

// Render 渲染输出
func (b *OutputBuffer) Render(styles Styles, width int) string {
	var sb strings.Builder
	for _, line := range b.lines {
		var styled string
		switch line.Type {
		case OutputSuccess:
			styled = styles.Success.Render() + " " + line.Content
		case OutputError:
			styled = styles.Error.Render() + " " + line.Content
		case OutputWarning:
			styled = styles.Warning.Render() + " " + line.Content
		case OutputInfo:
			styled = styles.Info.Render() + " " + line.Content
		default:
			styled = line.Content
		}
		sb.WriteString(styled + "\n")
	}
	return sb.String()
}
