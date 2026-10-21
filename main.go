package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"udirstats/internal/scan"
	"udirstats/internal/ui"
)

func main() {
	oneFS := flag.Bool("x", false, "não cruzar para outros filesystems (ex.: ignora /mnt/c no WSL)")
	apparent := flag.Bool("apparent", false, "iniciar exibindo o tamanho aparente em vez do uso em disco")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "uso: %s [flags] [caminho]\n\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	path := "."
	if flag.NArg() > 0 {
		path = flag.Arg(0)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "udirstats:", err)
		os.Exit(1)
	}
	if fi, err := os.Stat(abs); err != nil || !fi.IsDir() {
		if err == nil {
			err = fmt.Errorf("%s não é um diretório", abs)
		}
		fmt.Fprintln(os.Stderr, "udirstats:", err)
		os.Exit(1)
	}

	final, err := tea.NewProgram(ui.New(abs, scan.Options{OneFS: *oneFS}, *apparent), tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "udirstats:", err)
		os.Exit(1)
	}
	if m, ok := final.(ui.Model); ok && m.Err() != nil {
		fmt.Fprintln(os.Stderr, "udirstats:", m.Err())
		os.Exit(1)
	}
}
