//go:build unix

package scan

import (
	"os"
	"syscall"
)

// statOf extrai do FileInfo o uso em disco e a identidade do arquivo.
func statOf(_ string, fi os.FileInfo) (fileStat, bool) {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return fileStat{}, false
	}
	return fileStat{
		disk:  int64(st.Blocks) * 512,
		dev:   uint64(st.Dev),
		ino:   uint64(st.Ino),
		nlink: uint64(st.Nlink),
	}, true
}
