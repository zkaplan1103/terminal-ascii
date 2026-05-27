package bio

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// renderBioBox wraps bio paragraphs in a spaced-circle frame.
//
// Shape:
//     ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○
//     ○  paragraph text line one                     ○
//     ○  paragraph text line two                     ○
//     ○                                              ○
//     ○  next paragraph line one                     ○
//     ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○  ○
//
// Top/bottom rails are "○  " repeated to fill the total width.
// Side borders are a single ○ on each end with 1 space padding.
// Frame chars are dim gray; text inside is normal brightness.
func renderBioBox(paragraphs [][]string, innerWidth int) string {
	// Total frame width = innerWidth + 4 (2 circle chars + 2 space padding).
	// Each circle is 3 UTF-8 bytes but 1 terminal cell wide.
	totalWidth := innerWidth + 4

	rail := buildRail(totalWidth)

	var rows []string
	rows = append(rows, frameStyle.Render(rail))

	for pi, para := range paragraphs {
		for _, line := range para {
			rows = append(rows, sideRow(line, innerWidth))
		}
		if pi < len(paragraphs)-1 {
			rows = append(rows, sideRow("", innerWidth))
		}
	}

	rows = append(rows, frameStyle.Render(rail))
	return strings.Join(rows, "\n")
}

// sideRow renders one content row: ○ + space + padded content + space + ○.
func sideRow(content string, innerWidth int) string {
	runeLen := utf8.RuneCountInString(content)
	if runeLen > innerWidth {
		runes := []rune(content)
		content = string(runes[:innerWidth])
	} else {
		content = content + strings.Repeat(" ", innerWidth-runeLen)
	}
	return frameStyle.Render("○") + " " + bodyStyle.Render(content) + " " + frameStyle.Render("○")
}

// buildRail makes a "○  ○  ○  ..." string that fills exactly width terminal cells.
// Pattern unit is "○  " (1 cell circle + 2 spaces = 3 cells). If width isn't a
// multiple of 3, the last unit is trimmed to fit.
func buildRail(width int) string {
	var b strings.Builder
	cells := 0
	for cells < width {
		remaining := width - cells
		if remaining >= 3 {
			b.WriteString("○  ")
			cells += 3
		} else if remaining == 2 {
			b.WriteString("○ ")
			cells += 2
		} else {
			b.WriteString("○")
			cells++
		}
	}
	return b.String()
}

var frameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
