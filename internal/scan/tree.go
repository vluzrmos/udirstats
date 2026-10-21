package scan

import (
	"path/filepath"
	"sort"
)

// Node é um arquivo ou diretório da árvore escaneada.
// Para diretórios, Size, Apparent e Files incluem todo o conteúdo.
type Node struct {
	Name     string
	Size     int64 // uso em disco (st_blocks * 512)
	Apparent int64 // tamanho aparente (st_size)
	Files    int64 // quantidade de arquivos (não diretórios) contidos
	IsDir    bool
	OtherFS  bool // diretório em outro filesystem, não percorrido (-x)
	Err      error
	Children []*Node
	Parent   *Node

	self         int64 // tamanho em disco da própria entrada
	selfApparent int64
}

// Path devolve o caminho completo do nó.
func (n *Node) Path() string {
	if n.Parent == nil {
		return n.Name
	}
	return filepath.Join(n.Parent.Path(), n.Name)
}

// ReplaceWith substitui o conteúdo de n pelo de um novo scan do mesmo caminho,
// mantendo Name e Parent, e recalcula os totais dos ancestrais.
func (n *Node) ReplaceWith(fresh *Node) {
	n.Children = fresh.Children
	for _, c := range n.Children {
		c.Parent = n
	}
	n.IsDir = fresh.IsDir
	n.OtherFS = fresh.OtherFS
	n.Err = fresh.Err
	n.self = fresh.self
	n.selfApparent = fresh.selfApparent
	n.Size = fresh.Size
	n.Apparent = fresh.Apparent
	n.Files = fresh.Files
	for p := n.Parent; p != nil; p = p.Parent {
		p.sum()
	}
}

// sum recalcula os totais de n a partir dos filhos diretos.
func (n *Node) sum() {
	n.Size, n.Apparent, n.Files = n.self, n.selfApparent, 0
	for _, c := range n.Children {
		n.Size += c.Size
		n.Apparent += c.Apparent
		if c.IsDir {
			n.Files += c.Files
		} else {
			n.Files++
		}
	}
}

// finalize calcula os totais da subárvore e ordena os filhos por tamanho decrescente.
func (n *Node) finalize() {
	if !n.IsDir {
		n.Size, n.Apparent = n.self, n.selfApparent
		return
	}
	for _, c := range n.Children {
		c.finalize()
	}
	n.sum()
	sort.SliceStable(n.Children, func(i, j int) bool {
		return n.Children[i].Size > n.Children[j].Size
	})
}
