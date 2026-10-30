package scan

import (
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

const (
	fileAttributeSparse     = 0x200
	fileAttributeCompressed = 0x800
	invalidFileSize         = 0xFFFFFFFF
	defaultCluster          = 4096
)

var (
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procGetCompressedFileSize = kernel32.NewProc("GetCompressedFileSizeW")
	procGetDiskFreeSpace      = kernel32.NewProc("GetDiskFreeSpaceW")

	clusters sync.Map // volume -> tamanho do cluster em bytes
)

// statOf estima o uso em disco arredondando o tamanho para o cluster do volume.
// O Windows não expõe dispositivo/inode/nlink sem abrir o arquivo, então
// hardlinks não são deduplicados e -x não tem efeito (pontos de montagem em
// pastas são junções, que já não são seguidas).
func statOf(path string, fi os.FileInfo) (fileStat, bool) {
	d, ok := fi.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return fileStat{}, false
	}
	size := fi.Size()
	if d.FileAttributes&(fileAttributeSparse|fileAttributeCompressed) != 0 {
		if n, err := compressedSize(path); err == nil {
			size = n
		}
	}
	c := clusterSize(path)
	return fileStat{disk: (size + c - 1) / c * c, nlink: 1}, true
}

// compressedSize devolve o espaço realmente alocado de arquivos comprimidos ou esparsos.
func compressedSize(path string) (int64, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var high uint32
	low, _, errno := procGetCompressedFileSize.Call(uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(&high)))
	if uint32(low) == invalidFileSize && errno != syscall.Errno(0) {
		return 0, errno
	}
	return int64(high)<<32 | int64(uint32(low)), nil
}

func clusterSize(path string) int64 {
	vol := filepath.VolumeName(path)
	if c, ok := clusters.Load(vol); ok {
		return c.(int64)
	}
	c := int64(defaultCluster)
	if p, err := syscall.UTF16PtrFromString(vol + `\`); err == nil {
		var spc, bps, free, total uint32
		r, _, _ := procGetDiskFreeSpace.Call(uintptr(unsafe.Pointer(p)),
			uintptr(unsafe.Pointer(&spc)), uintptr(unsafe.Pointer(&bps)),
			uintptr(unsafe.Pointer(&free)), uintptr(unsafe.Pointer(&total)))
		if r != 0 && spc*bps > 0 {
			c = int64(spc) * int64(bps)
		}
	}
	clusters.Store(vol, c)
	return c
}
