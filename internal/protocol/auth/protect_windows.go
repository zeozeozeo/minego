//go:build windows

package auth

import (
	"bytes"
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

var (
	crypt32            = syscall.NewLazyDLL("crypt32.dll")
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	cryptProtectData   = crypt32.NewProc("CryptProtectData")
	cryptUnprotectData = crypt32.NewProc("CryptUnprotectData")
	localFree          = kernel32.NewProc("LocalFree")
)
var dpapiMagic = []byte("MINEGO-DPAPI\x00")

const cryptprotectUIForbidden = 0x1

type dataBlob struct {
	size uint32
	data *byte
}

func blob(data []byte) dataBlob {
	if len(data) == 0 {
		return dataBlob{}
	}
	return dataBlob{uint32(len(data)), &data[0]}
}
func blobBytes(b dataBlob) []byte {
	if b.size == 0 || b.data == nil {
		return nil
	}
	return append([]byte(nil), unsafe.Slice(b.data, b.size)...)
}
func protect(data []byte) ([]byte, error) {
	in := blob(data)
	var out dataBlob
	r, _, e := cryptProtectData.Call(uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, cryptprotectUIForbidden, uintptr(unsafe.Pointer(&out)))
	if r == 0 {
		return nil, fmt.Errorf("CryptProtectData: %w", e)
	}
	defer localFree.Call(uintptr(unsafe.Pointer(out.data)))
	return append(append([]byte(nil), dpapiMagic...), blobBytes(out)...), nil
}
func unprotect(data []byte) ([]byte, error) {
	if !bytes.HasPrefix(data, dpapiMagic) {
		return nil, errors.New("credential file is not DPAPI protected")
	}
	in := blob(data[len(dpapiMagic):])
	var out dataBlob
	r, _, e := cryptUnprotectData.Call(uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, cryptprotectUIForbidden, uintptr(unsafe.Pointer(&out)))
	if r == 0 {
		return nil, fmt.Errorf("CryptUnprotectData: %w", e)
	}
	defer localFree.Call(uintptr(unsafe.Pointer(out.data)))
	return blobBytes(out), nil
}
