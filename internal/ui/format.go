package ui

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
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

// Sanitize torna visíveis caracteres de controle, controles bidirecionais e
// bytes UTF-8 inválidos (como o `ls -b`), para que nomes de arquivo não
// injetem sequências de escape no terminal nem falsifiquem o texto exibido.
func Sanitize(s string) string {
	if !needsSanitize(s) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == utf8.RuneError && size == 1:
			fmt.Fprintf(&b, `\x%02x`, s[i])
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == 0x1b:
			b.WriteString(`\e`)
		case r < 0x100 && unicode.IsControl(r):
			fmt.Fprintf(&b, `\x%02x`, r)
		case unicode.IsControl(r) || unicode.Is(unicode.Bidi_Control, r):
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			b.WriteRune(r)
		}
		i += size
	}
	return b.String()
}

func needsSanitize(s string) bool {
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if (r == utf8.RuneError && size == 1) || unicode.IsControl(r) || unicode.Is(unicode.Bidi_Control, r) {
			return true
		}
		i += size
	}
	return false
}
