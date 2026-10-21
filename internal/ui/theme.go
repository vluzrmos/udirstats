package ui

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/mattn/go-runewidth"
)

// As cores fixas têm valor explícito para true color, 256 e 16 cores: a conversão
// automática do termenv distorce cinzas e tons escuros (cinza claro vira magenta
// em 16 cores). Os degradês são calculados em true color e convertidos pelo
// termenv. NO_COLOR é respeitado.

var (
	heatStops  = []string{"#4ade80", "#facc15", "#fb923c", "#f43f5e"} // verde → amarelo → laranja → vermelho
	titleStops = []string{"#a78bfa", "#f472b6", "#fb923c"}            // violeta → rosa → laranja

	fgColor    = adaptive(cc("#1f2937", "235", "0"), cc("#e5e7eb", "254", "7"))
	dimColor   = adaptive(cc("#9ca3af", "247", "8"), cc("#6b7280", "243", "8"))
	trackColor = adaptive(cc("#e5e7eb", "254", "7"), cc("#374151", "237", "8"))
	selColor   = adaptive(cc("#dbeafe", "189", "7"), cc("#2e3350", "238", "8"))
	badgeColor = adaptive(cc("#e0e7ff", "253", "7"), cc("#3b3f63", "239", "8"))
	accent     = cc("#c084fc", "177", "13")
	dirColor   = cc("#60a5fa", "75", "12")
	errColor   = cc("#f43f5e", "204", "9")

	plain     = lipgloss.NewStyle().Foreground(fgColor)
	dimStyle  = lipgloss.NewStyle().Foreground(dimColor)
	dirStyle  = lipgloss.NewStyle().Foreground(dirColor).Bold(true)
	linkStyle = lipgloss.NewStyle().Foreground(cc("#2dd4bf", "43", "14")).Italic(true)
	errStyle  = lipgloss.NewStyle().Foreground(errColor)
	keyStyle  = lipgloss.NewStyle().Foreground(accent).Bold(true)
	pathStyle = lipgloss.NewStyle().Foreground(fgColor).Bold(true)
	badge     = lipgloss.NewStyle().Foreground(fgColor).Background(badgeColor).Padding(0, 1)
	spinner   = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	heatTable = gradientTable(heatStops, 101)
)

// Cores por tipo de arquivo, como o WinDirStat faz por extensão.
var kindStyles = map[string]lipgloss.Style{}

func init() {
	kinds := []struct {
		color lipgloss.CompleteColor
		exts  string
	}{
		{cc("#f472b6", "211", "13"), "jpg jpeg png gif bmp webp svg ico tif tiff heic psd raw"},        // imagem
		{cc("#c084fc", "177", "5"), "mp4 mkv avi mov wmv webm flv m4v mpg mpeg"},                       // vídeo
		{cc("#22d3ee", "44", "14"), "mp3 flac wav ogg m4a aac opus wma"},                               // áudio
		{cc("#fb7185", "210", "9"), "zip tar gz tgz bz2 xz zst 7z rar deb rpm iso img apk jar whl"},    // compactado
		{cc("#4ade80", "78", "10"), "go rs c h cpp hpp py js ts tsx jsx php rb java kt cs sh lua sql"}, // código
		{cc("#facc15", "220", "11"), "pdf doc docx odt xls xlsx ods ppt pptx txt md csv rtf epub"},     // documento
		{cc("#fb923c", "209", "3"), "so a o dll exe bin ko pyc class wasm"},                            // binário
		{cc("#a3a3a3", "248", "8"), "log tmp bak swp cache lock"},                                      // descartável
		{cc("#93c5fd", "111", "6"), "json yaml yml toml xml ini conf cfg env html css"},                // config/web
	}
	for _, k := range kinds {
		st := lipgloss.NewStyle().Foreground(k.color)
		for _, e := range strings.Fields(k.exts) {
			kindStyles[e] = st
		}
	}
}

func cc(trueColor, ansi256, ansi string) lipgloss.CompleteColor {
	return lipgloss.CompleteColor{TrueColor: trueColor, ANSI256: ansi256, ANSI: ansi}
}

func adaptive(light, dark lipgloss.CompleteColor) lipgloss.CompleteAdaptiveColor {
	return lipgloss.CompleteAdaptiveColor{Light: light, Dark: dark}
}

func fileStyle(name string) lipgloss.Style {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	if st, ok := kindStyles[ext]; ok {
		return st
	}
	return plain
}

// gradientTable interpola n cores entre as paradas, no espaço Luv (perceptual).
func gradientTable(stops []string, n int) []lipgloss.Color {
	cs := make([]colorful.Color, len(stops))
	for i, s := range stops {
		cs[i], _ = colorful.Hex(s)
	}
	out := make([]lipgloss.Color, n)
	for i := range out {
		t := float64(i) / float64(max(1, n-1)) * float64(len(cs)-1)
		j := min(int(t), len(cs)-2)
		out[i] = lipgloss.Color(cs[j].BlendLuv(cs[j+1], t-float64(j)).Clamped().Hex())
	}
	return out
}

func heat(frac float64) lipgloss.Color {
	return heatTable[int(max(0, min(1, frac))*100)]
}

// seg é um trecho de texto com estilo próprio dentro de uma linha.
type seg struct {
	text  string
	style lipgloss.Style
}

// line junta os trechos, truncando em width colunas; com selected,
// aplica o fundo de seleção em todos os trechos e preenche até o fim.
func line(segs []seg, width int, selected bool) string {
	var b strings.Builder
	left := width
	for _, s := range segs {
		if left <= 0 {
			break
		}
		t := Sanitize(s.text) // todo texto exibido passa por aqui, inclusive nomes e erros
		if w := runewidth.StringWidth(t); w > left {
			t = runewidth.Truncate(t, left, "…")
		}
		left -= runewidth.StringWidth(t)
		st := s.style
		if selected {
			// texto apagado sumiria no fundo de seleção
			if st.GetForeground() == lipgloss.TerminalColor(dimColor) {
				st = st.Foreground(fgColor)
			}
			st = st.Background(selColor)
		}
		b.WriteString(st.Render(t))
	}
	if selected && left > 0 {
		b.WriteString(lipgloss.NewStyle().Background(selColor).Render(strings.Repeat(" ", left)))
	}
	return b.String()
}

// barSegs colore cada coluna da barra pela posição, sobre um trilho de fundo:
// barras longas chegam ao vermelho.
func barSegs(frac float64, width int) []seg {
	segs := make([]seg, 0, width)
	track := lipgloss.NewStyle().Background(trackColor)
	for i, r := range []rune(Bar(frac, width)) {
		segs = append(segs, seg{string(r), track.Foreground(heat(float64(i) / float64(width-1)))})
	}
	return segs
}

// gradientText pinta cada caractere de s ao longo das paradas.
func gradientText(s string, stops []string) []seg {
	rs := []rune(s)
	cols := gradientTable(stops, len(rs))
	segs := make([]seg, len(rs))
	for i, r := range rs {
		segs[i] = seg{string(r), lipgloss.NewStyle().Foreground(cols[i]).Bold(true)}
	}
	return segs
}
