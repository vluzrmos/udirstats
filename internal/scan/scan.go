package scan

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
)

// Options controla o comportamento do scan.
type Options struct {
	OneFS bool // não cruzar para outros filesystems
}

// Progress expõe contadores do scan em andamento; seguro para leitura concorrente.
type Progress struct {
	Files   atomic.Int64
	Dirs    atomic.Int64
	Bytes   atomic.Int64
	current atomic.Pointer[string]
}

// Current devolve o diretório sendo lido no momento.
func (p *Progress) Current() string {
	if s := p.current.Load(); s != nil {
		return *s
	}
	return ""
}

type inode struct{ dev, ino uint64 }

type scanner struct {
	opts    Options
	prog    *Progress
	sem     chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	seen    map[inode]struct{}
	rootDev uint64
}

// Scan percorre root e devolve a árvore com totais calculados e filhos
// ordenados por tamanho. prog pode ser nil.
func Scan(root string, opts Options, prog *Progress) (*Node, error) {
	if prog == nil {
		prog = &Progress{}
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	fi, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("%s não é um diretório", abs)
	}
	st := fi.Sys().(*syscall.Stat_t)

	s := &scanner{
		opts:    opts,
		prog:    prog,
		sem:     make(chan struct{}, runtime.NumCPU()*4),
		seen:    make(map[inode]struct{}),
		rootDev: uint64(st.Dev),
	}
	n := &Node{
		Name:         abs,
		IsDir:        true,
		self:         int64(st.Blocks) * 512,
		selfApparent: fi.Size(),
	}
	s.dir(n, abs)
	s.wg.Wait()
	n.finalize()
	return n, nil
}

func (s *scanner) dir(n *Node, path string) {
	s.prog.current.Store(&path)
	s.prog.Dirs.Add(1)

	entries, err := os.ReadDir(path)
	if err != nil {
		n.Err = err // ReadDir ainda devolve o que conseguiu ler
	}
	n.Children = make([]*Node, 0, len(entries))
	for _, e := range entries {
		c := &Node{Name: e.Name(), Parent: n}
		n.Children = append(n.Children, c)

		fi, err := e.Info() // lstat: não segue symlinks
		if err != nil {
			c.Err = err
			continue
		}
		c.IsDir = fi.IsDir()
		c.IsLink = fi.Mode()&os.ModeSymlink != 0
		st, ok := fi.Sys().(*syscall.Stat_t)
		if !ok {
			c.selfApparent = fi.Size()
			continue
		}
		disk, apparent := int64(st.Blocks)*512, fi.Size()
		if !c.IsDir && st.Nlink > 1 && s.dup(inode{uint64(st.Dev), uint64(st.Ino)}) {
			disk, apparent = 0, 0
		}
		c.self, c.selfApparent = disk, apparent
		s.prog.Bytes.Add(disk)

		if !c.IsDir {
			s.prog.Files.Add(1)
			continue
		}
		if s.opts.OneFS && uint64(st.Dev) != s.rootDev {
			c.OtherFS = true
			continue
		}
		p := filepath.Join(path, e.Name())
		select {
		case s.sem <- struct{}{}:
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				defer func() { <-s.sem }()
				s.dir(c, p)
			}()
		default:
			s.dir(c, p)
		}
	}
}

// dup registra o inode e informa se ele já tinha sido contado (hardlink).
func (s *scanner) dup(i inode) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.seen[i]; ok {
		return true
	}
	s.seen[i] = struct{}{}
	return false
}
