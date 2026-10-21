package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"udirstats/internal/scan"
)

type sortMode int

const (
	bySize sortMode = iota
	byName
	byFiles
)

func (s sortMode) String() string {
	return [...]string{"tamanho", "nome", "arquivos"}[s]
}

const barWidth = 20

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	dirStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	selectedStyle = lipgloss.NewStyle().Reverse(true)
	spinner       = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

type tickMsg time.Time

type scanDoneMsg struct {
	target *scan.Node // nil no scan inicial; nó a substituir no rescan
	node   *scan.Node
	err    error
}

// Model é o estado da TUI.
type Model struct {
	rootPath string
	opts     scan.Options

	root *scan.Node
	cur  *scan.Node

	items  []*scan.Node // filhos ordenados de cur; nil representa ".."
	cursor int
	offset int
	stack  []int // cursores dos diretórios anteriores

	sort     sortMode
	apparent bool

	scanning  bool
	scanStart time.Time
	prog      *scan.Progress
	frame     int
	err       error
	status    string

	width, height int
}

// New cria o modelo que escaneia path ao iniciar.
func New(path string, opts scan.Options, apparent bool) Model {
	return Model{
		rootPath: path, opts: opts, apparent: apparent, width: 80, height: 24,
		scanning: true, scanStart: time.Now(), prog: &scan.Progress{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(scanCmd(m.rootPath, nil, m.opts, m.prog), tick())
}

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// startScan reescaneia target em segundo plano.
func (m *Model) startScan(target *scan.Node) tea.Cmd {
	m.scanning = true
	m.scanStart = time.Now()
	m.prog = &scan.Progress{}
	return scanCmd(target.Path(), target, m.opts, m.prog)
}

func scanCmd(path string, target *scan.Node, opts scan.Options, prog *scan.Progress) tea.Cmd {
	return func() tea.Msg {
		n, err := scan.Scan(path, opts, prog)
		return scanDoneMsg{target: target, node: n, err: err}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.clamp()
		return m, nil

	case tickMsg:
		if !m.scanning {
			return m, nil
		}
		m.frame++
		return m, tick()

	case scanDoneMsg:
		m.scanning = false
		elapsed := time.Since(m.scanStart).Round(time.Millisecond)
		if msg.err != nil {
			if msg.target == nil {
				m.err = msg.err
				return m, tea.Quit
			}
			m.status = "erro ao reescanear: " + msg.err.Error()
			return m, nil
		}
		if msg.target == nil {
			m.root, m.cur = msg.node, msg.node
		} else {
			msg.target.ReplaceWith(msg.node)
		}
		m.status = fmt.Sprintf("scan concluído em %s", elapsed)
		m.refresh()
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.scanning {
			if msg.String() == "q" {
				return m, tea.Quit
			}
			return m, nil
		}
		return m.handleKey(msg.String())
	}
	return m, nil
}

func (m Model) handleKey(key string) (tea.Model, tea.Cmd) {
	page := m.listHeight()
	switch key {
	case "q":
		return m, tea.Quit
	case "up", "k":
		m.cursor--
	case "down", "j":
		m.cursor++
	case "pgup", "ctrl+u":
		m.cursor -= page
	case "pgdown", "ctrl+d", " ":
		m.cursor += page
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		m.cursor = len(m.items) - 1
	case "enter", "right", "l":
		if m.cursor < len(m.items) {
			if sel := m.items[m.cursor]; sel == nil {
				m.goUp()
			} else if sel.IsDir && !sel.OtherFS {
				m.stack = append(m.stack, m.cursor)
				m.cur = sel
				m.cursor, m.offset = 0, 0
				m.refresh()
				if len(m.items) > 1 {
					m.cursor = 1 // pula o ".."
				}
			}
		}
	case "backspace", "left", "h", "esc":
		m.goUp()
	case "s":
		m.sort = (m.sort + 1) % 3
		m.refresh()
		m.status = "ordenando por " + m.sort.String()
	case "a":
		m.apparent = !m.apparent
		m.refresh()
		m.status = "exibindo " + m.sizeLabel()
	case "r":
		m.status = ""
		return m, tea.Batch(m.startScan(m.cur), tick())
	}
	m.clamp()
	return m, nil
}

func (m *Model) goUp() {
	if m.cur.Parent == nil {
		return
	}
	child := m.cur
	m.cur = m.cur.Parent
	m.refresh()
	m.cursor = 0
	if n := len(m.stack); n > 0 {
		m.cursor = m.stack[n-1]
		m.stack = m.stack[:n-1]
	}
	// a ordem pode ter mudado (rescan, troca de ordenação): reposiciona no filho de onde viemos
	if m.cursor >= len(m.items) || m.items[m.cursor] != child {
		for i, it := range m.items {
			if it == child {
				m.cursor = i
			}
		}
	}
	m.offset = m.cursor - m.listHeight()/2
	m.clamp()
}

func (m *Model) size(n *scan.Node) int64 {
	if m.apparent {
		return n.Apparent
	}
	return n.Size
}

func (m *Model) sizeLabel() string {
	if m.apparent {
		return "tamanho aparente"
	}
	return "uso em disco"
}

// refresh recalcula a lista exibida a partir de cur.
func (m *Model) refresh() {
	if m.cur == nil {
		return
	}
	items := make([]*scan.Node, 0, len(m.cur.Children)+1)
	if m.cur.Parent != nil {
		items = append(items, nil)
	}
	children := append([]*scan.Node(nil), m.cur.Children...)
	sort.SliceStable(children, func(i, j int) bool {
		a, b := children[i], children[j]
		switch m.sort {
		case byName:
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		case byFiles:
			if a.Files != b.Files {
				return a.Files > b.Files
			}
		}
		return m.size(a) > m.size(b)
	})
	m.items = append(items, children...)
	m.clamp()
}

func (m *Model) listHeight() int {
	// título, caminho, cabeçalho, linha de status, ajuda
	return max(1, m.height-5)
}

func (m *Model) clamp() {
	m.cursor = max(0, min(m.cursor, len(m.items)-1))
	h := m.listHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+h {
		m.offset = m.cursor - h + 1
	}
	m.offset = max(0, min(m.offset, len(m.items)-h))
}

func (m Model) View() string {
	if m.err != nil {
		return ""
	}
	if m.root == nil || (m.scanning && m.cur == nil) {
		return m.scanView()
	}

	var b strings.Builder
	w := m.width
	total := m.size(m.cur)

	title := fmt.Sprintf("udirstats  %s  %d arquivos  [%s · ordem: %s]",
		HumanSize(total), m.cur.Files, m.sizeLabel(), m.sort)
	b.WriteString(titleStyle.Render(truncate(title, w)) + "\n")
	b.WriteString(dirStyle.Render(truncate(m.cur.Path(), w)) + "\n")
	b.WriteString(dimStyle.Render(truncate(fmt.Sprintf("%10s %6s  %-*s  %8s  nome",
		"tamanho", "%", barWidth+2, "", "arquivos"), w)) + "\n")

	h := m.listHeight()
	for i := m.offset; i < m.offset+h; i++ {
		if i < len(m.items) {
			b.WriteString(m.row(i, total, w))
		}
		b.WriteString("\n")
	}

	status := m.status
	if m.scanning {
		status = fmt.Sprintf("%s reescaneando… %d arquivos, %s",
			spinner[m.frame%len(spinner)], m.prog.Files.Load(), HumanSize(m.prog.Bytes.Load()))
	} else if m.cursor < len(m.items) && m.items[m.cursor] != nil {
		if sel := m.items[m.cursor]; sel.Err != nil {
			status = "erro: " + sel.Err.Error()
		}
	}
	b.WriteString(truncate(status, w) + "\n")
	b.WriteString(dimStyle.Render(truncate(
		"↑↓ mover · enter entrar · ← voltar · s ordenar · a disco/aparente · r reescanear · q sair", w)))
	return b.String()
}

func (m *Model) row(i int, total int64, w int) string {
	n := m.items[i]
	var line string
	if n == nil {
		line = fmt.Sprintf("%10s %6s  %-*s  %8s  ..", "", "", barWidth+2, "", "")
	} else {
		sz := m.size(n)
		frac := 0.0
		if total > 0 {
			frac = float64(sz) / float64(total)
		}
		files := ""
		if n.IsDir {
			files = fmt.Sprint(n.Files)
		}
		name := n.Name
		switch {
		case n.OtherFS:
			name += "/ (outro filesystem)"
		case n.IsDir:
			name += "/"
		}
		if n.Err != nil {
			name = "! " + name
		}
		line = fmt.Sprintf("%10s %5.1f%%  [%s]  %8s  %s", HumanSize(sz), frac*100, Bar(frac, barWidth), files, name)
	}
	line = truncate(line, w)
	switch {
	case i == m.cursor:
		return selectedStyle.Render(runewidth.FillRight(line, w))
	case n != nil && n.Err != nil:
		return errStyle.Render(line)
	case n == nil || n.IsDir:
		return dirStyle.Render(line)
	}
	return line
}

func (m Model) scanView() string {
	elapsed := time.Since(m.scanStart).Round(time.Second)
	return fmt.Sprintf("\n  %s Escaneando %s\n\n  %d diretórios · %d arquivos · %s · %s\n\n  %s\n\n  %s\n",
		spinner[m.frame%len(spinner)], titleStyle.Render(m.rootPath),
		m.prog.Dirs.Load(), m.prog.Files.Load(), HumanSize(m.prog.Bytes.Load()), elapsed,
		dimStyle.Render(truncate(m.prog.Current(), m.width-4)),
		dimStyle.Render("q para sair"))
}

// Err devolve o erro fatal do scan inicial, se houve.
func (m Model) Err() error { return m.err }

func truncate(s string, w int) string {
	return runewidth.Truncate(s, max(0, w), "…")
}
