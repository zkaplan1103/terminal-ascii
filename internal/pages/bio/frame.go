package bio

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderBrailleColumn renders the entire right-side text column as a series
// of braille-filled rows. Each text line has its content on the left and a
// braille gradient filling the space to the right up to totalWidth.
//
// Row numbering (0-based, top to bottom) is used to determine which rows are
// lit by the glow wavefront. The caller passes:
//
//   - titleRowCount: how many rows the title occupies (they are row 0..N-1)
//   - glowRow: current wavefront row index from titleAnim.titleGlowRow()
//   - flicker: current flicker colour override (empty = not flickering)
//   - totalWidth: hard right edge in terminal columns
//
// Structure and row indices:
//
//	[0 .. titleRows-1]      title lines         (rendered by caller, not here)
//	[titleRows]             divider row
//	[titleRows+1 .. T-2]   bio content rows
//	[T-1]                   divider row
//	[T]                     nav row             (rendered by caller, not here)
//
// This function returns the braille suffix for each row — the caller
// prepends the actual text content.

// textRow returns the braille fill suffix for a content row.
// rowIdx is the global row index in the braille field.
func textRowFill(text string, totalWidth, pad, rowIdx, glowRow int, flicker lipgloss.Color) string {
	lit := glowRow >= 0 && rowIdx <= glowRow
	fill := brailleContentRow(len(text), totalWidth, pad, rowSeed(rowIdx), lit, flicker)
	return text + fill
}

// divRowFill returns a full-width divider row (no text prefix).
func divRowFill(totalWidth, rowIdx, glowRow int, flicker lipgloss.Color) string {
	lit := glowRow >= 0 && rowIdx <= glowRow
	return brailleDividerRow(totalWidth, rowSeed(rowIdx), lit, flicker)
}

// renderBrailleTextColumn builds the full right column:
//
//	titleStr   (pre-rendered, multi-line)
//	divider row
//	bio rows
//	[padding blank rows if needed]
//	divider row
//	nav row
//
// Returns the composed string and the total braille row count
// (used to set titleAnim.brailRows so the drop animation knows when to stop).
func renderBrailleTextColumn(
	titleStr string,
	paragraphs [][]string,
	navStr string,
	totalWidth int, // hard right edge
	pad int, // spaces between text and braille
	glowRow int,
	flicker lipgloss.Color,
	padToHeight int, // portrait height — text column must match this
) (out string, brailRowCount int) {

	// ── Flatten bio lines ────────────────────────────────────────────────────
	var bioLines []string
	for pi, para := range paragraphs {
		for _, line := range para {
			bioLines = append(bioLines, line)
		}
		if pi < len(paragraphs)-1 {
			bioLines = append(bioLines, "") // blank between paragraphs
		}
	}

	// ── Count title rows ─────────────────────────────────────────────────────
	titleRows := strings.Count(titleStr, "\n") + 1

	// ── Assign global row indices ─────────────────────────────────────────────
	// row 0 .. titleRows-1      : title (braille fill only, text pre-rendered)
	// row titleRows             : divider after title
	// row titleRows+1 .. T-2   : bio lines
	// row T-1                   : divider after bio / before nav
	// row T                     : nav
	divider1Idx := titleRows
	bioStartIdx := titleRows + 1
	bioEndIdx := bioStartIdx + len(bioLines) // exclusive

	// We may need padding rows between bio and nav to match padToHeight.
	// Natural height = titleRows + 1(div) + len(bioLines) + 1(div) + 1(nav)
	naturalH := titleRows + 1 + len(bioLines) + 1 + 1
	extraPad := 0
	if padToHeight > naturalH {
		extraPad = padToHeight - naturalH
	}

	divider2Idx := bioEndIdx + extraPad
	navIdx := divider2Idx + 1
	totalRows := navIdx + 1

	brailRowCount = totalRows

	// ── Build rows ────────────────────────────────────────────────────────────
	var rows []string

	// Title rows — text is already colourised by title.render(); we only add
	// the braille fill to the right. Each title line is a separate row.
	titleLines := strings.Split(titleStr, "\n")
	for i, tl := range titleLines {
		// Strip ANSI for length measurement — we need the visible width.
		visible := lipgloss.Width(tl)
		fill := func() string {
			lit := glowRow >= 0 && i <= glowRow
			return brailleContentRow(visible, totalWidth, pad, rowSeed(i), lit, flicker)
		}()
		rows = append(rows, tl+fill)
	}

	// Divider after title.
	rows = append(rows, divRowFill(totalWidth, divider1Idx, glowRow, flicker))

	// Bio rows.
	for i, bl := range bioLines {
		rowIdx := bioStartIdx + i
		if bl == "" {
			// Blank paragraph separator — treat as a faint divider.
			rows = append(rows, divRowFill(totalWidth, rowIdx, glowRow, flicker))
		} else {
			indent := "  "
			full := indent + bl
			lit := glowRow >= 0 && rowIdx <= glowRow
			fill := brailleContentRow(len(full), totalWidth, pad, rowSeed(rowIdx), lit, flicker)
			rows = append(rows, full+fill)
		}
	}

	// Extra padding rows between bio and nav.
	// Rather than uniform divider rows, each one gets a unique density profile
	// that varies with its position in the padding zone — some are wisp-only,
	// some have a mid patch on the right, none look identical.
	for i := 0; i < extraPad; i++ {
		rowIdx := bioEndIdx + i
		lit := glowRow >= 0 && rowIdx <= glowRow
		rows = append(rows, braillePadRow(totalWidth, rowSeed(rowIdx), i, extraPad, lit, flicker))
	}

	// Divider before nav — sparsest row of the block.
	rows = append(rows, divRowFill(totalWidth, divider2Idx, glowRow, flicker))

	// Nav row — braille fills to the right of the nav text.
	navFull := "  " + navStr
	navLit := glowRow >= 0 && navIdx <= glowRow
	navFill := brailleContentRow(len(navFull), totalWidth, pad, rowSeed(navIdx), navLit, flicker)
	rows = append(rows, navFull+navFill)

	return strings.Join(rows, "\n"), brailRowCount
}
