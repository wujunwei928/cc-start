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
	if len(b.lines) == 0 {
		return ""
	}

	var sb strings.Builder

	for i, line := range b.lines {
		var styled string
		switch line.Type {
		case OutputSuccess:
			// 成功消息：绿色 ✓ 前缀
			styled = styles.Success.Render("✓") + " " + line.Content
		case OutputError:
			// 错误消息：红色 ✗ 前缀
			styled = styles.Error.Render("✗") + " " + line.Content
		case OutputWarning:
			// 警告消息：黄色 ⚠ 前缀
			styled = styles.Warning.Render("⚠") + " " + line.Content
		case OutputInfo:
			// 信息消息：蓝色 ● 前缀或原始内容
			if strings.HasPrefix(line.Content, "$ ") {
				// 命令行：原始内容
				styled = styles.Command.Render(line.Content)
			} else if strings.HasPrefix(line.Content, "───") {
				// 分隔线：灰色
				styled = styles.Separator.Render(line.Content)
			} else {
				// 其他信息：蓝色 ● 前缀
				styled = styles.Info.Render("●") + " " + line.Content
			}
		default:
			styled = line.Content
		}

		// 在命令行前添加空行
		if i > 0 && strings.HasPrefix(line.Content, "$ ") {
			sb.WriteString("\n")
		}

		sb.WriteString(styled + "\n")
	}
	return sb.String()
}
