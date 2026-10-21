package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
}

func child(t *testing.T, n *Node, name string) *Node {
	t.Helper()
	for _, c := range n.Children {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("%s não encontrado em %s", name, n.Path())
	return nil
}

// contentApparent soma o tamanho aparente só dos arquivos da subárvore.
func contentApparent(n *Node) int64 {
	if !n.IsDir {
		return n.Apparent
	}
	var t int64
	for _, c := range n.Children {
		t += contentApparent(c)
	}
	return t
}

func TestScanSizesAndOrder(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "small.txt"), 100)
	write(t, filepath.Join(root, "big", "a.bin"), 200_000)
	write(t, filepath.Join(root, "big", "b.bin"), 50_000)
	write(t, filepath.Join(root, "mid", "c.bin"), 30_000)

	n, err := Scan(root, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if n.Files != 4 {
		t.Errorf("Files = %d, esperado 4", n.Files)
	}
	if got := contentApparent(n); got != 280_100 {
		t.Errorf("tamanho aparente dos arquivos = %d, esperado 280100", got)
	}
	if n.Apparent < 280_100 {
		t.Errorf("Apparent da raiz = %d, menor que o conteúdo", n.Apparent)
	}
	if got := child(t, n, "big").Files; got != 2 {
		t.Errorf("big.Files = %d, esperado 2", got)
	}

	names := []string{}
	for _, c := range n.Children {
		names = append(names, c.Name)
	}
	if names[0] != "big" || names[1] != "mid" || names[2] != "small.txt" {
		t.Errorf("ordem = %v, esperado [big mid small.txt]", names)
	}
	for i := 1; i < len(n.Children); i++ {
		if n.Children[i-1].Size < n.Children[i].Size {
			t.Errorf("filhos não ordenados por tamanho: %v", names)
		}
	}
}

func TestScanSymlinkNotFollowed(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "data", "big.bin"), 100_000)
	if err := os.Symlink(filepath.Join(root, "data"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}

	n, err := Scan(root, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	l := child(t, n, "link")
	if !l.IsLink || l.IsDir || len(l.Children) != 0 {
		t.Errorf("symlink foi seguido")
	}
	if got := contentApparent(n) - l.Apparent; got != 100_000 {
		t.Errorf("conteúdo sem o link = %d, esperado 100000", got)
	}
}

func TestScanHardlinkCountedOnce(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a.bin")
	write(t, a, 100_000)
	if err := os.Link(a, filepath.Join(root, "b.bin")); err != nil {
		t.Fatal(err)
	}

	n, err := Scan(root, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := contentApparent(n); got != 100_000 {
		t.Errorf("hardlink contado duas vezes: %d", got)
	}
	if n.Files != 2 {
		t.Errorf("Files = %d, esperado 2", n.Files)
	}
}

func TestReplaceWithUpdatesAncestors(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "sub", "a.bin"), 10_000)

	n, err := Scan(root, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	sub := child(t, n, "sub")
	before := contentApparent(n)

	write(t, filepath.Join(root, "sub", "b.bin"), 40_000)
	fresh, err := Scan(sub.Path(), Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	sub.ReplaceWith(fresh)

	if sub.Name != "sub" || sub.Parent != n {
		t.Errorf("ReplaceWith alterou nome/pai")
	}
	if contentApparent(n) != before+40_000 || n.Apparent < before+40_000 {
		t.Errorf("raiz Apparent = %d, esperado %d", n.Apparent, before+40_000)
	}
	if n.Files != 2 {
		t.Errorf("raiz Files = %d, esperado 2", n.Files)
	}
}

func TestScanNotDir(t *testing.T) {
	f := filepath.Join(t.TempDir(), "f")
	write(t, f, 1)
	if _, err := Scan(f, Options{}, nil); err == nil {
		t.Error("esperado erro ao escanear arquivo")
	}
}
