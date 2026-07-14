//go:build !windows

package auth

// Unix token files are protected by their 0600 mode and 0700 parent. These
// helpers mirror the Windows DPAPI boundary without adding a dependency.
func protect(data []byte) ([]byte, error)   { return data, nil }
func unprotect(data []byte) ([]byte, error) { return data, nil }
