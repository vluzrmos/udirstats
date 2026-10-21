package ui

import (
	"fmt"
	"strings"
)

// HumanSize formata bytes em unidades binárias (KiB, MiB, ...).
func HumanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit && exp < 5; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

var partials = []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉'}

// Bar desenha uma barra com width colunas preenchida na fração frac (0..1),
// usando blocos parciais para resolução de 1/8 de coluna.
func Bar(frac float64, width int) string {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	eighths := int(frac*float64(width*8) + 0.5)
	full, rest := eighths/8, eighths%8
	var sb strings.Builder
	sb.WriteString(strings.Repeat("█", full))
	if full < width {
		sb.WriteRune(partials[rest])
		sb.WriteString(strings.Repeat(" ", width-full-1))
	}
	return sb.String()
}
