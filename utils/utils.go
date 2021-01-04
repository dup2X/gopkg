// Package utils ...
package utils

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	_ float64 = 1 << (10 * iota)
	// KBytes ...
	KBytes
	// MBytes ...
	MBytes
	// GBytes ...
	GBytes
	// TBytes ...
	TBytes
	// PBytes ...
	PBytes
)

// GetBinName ...
func GetBinName() string {
	return filepath.Base(os.Args[0])
}

// DumpHex ...
func DumpHex(data []byte) string {
	return hex.Dump(data)
}

const (
	tmpDir = "/tmp"
)

// WritePid writes pid in /tmp/proc-name.pid
// if procName == "", it will call GetBinName that
// return exec_bin as name
func WritePid(procName string) (pid int, err error) {
	if procName == "" {
		procName = GetBinName()
	}
	pid = os.Getpid()
	pidFile := filepath.Join(tmpDir, fmt.Sprintf("%s.pid", procName))
	err = ioutil.WriteFile(pidFile, []byte(fmt.Sprint(pid)), 0644)
	return
}

// BandWidth ...
func BandWidth(bytes float64) string {
	switch {
	case bytes > PBytes:
		return fmt.Sprintf("%.3fPB", bytes/PBytes)
	case bytes > TBytes:
		return fmt.Sprintf("%.3fTB", bytes/TBytes)
	case bytes > GBytes:
		return fmt.Sprintf("%.3fGB", bytes/GBytes)
	case bytes > MBytes:
		return fmt.Sprintf("%.3fMB", bytes/MBytes)
	case bytes > KBytes:
		return fmt.Sprintf("%.3fKB", bytes/KBytes)
	case bytes > 0:
		return fmt.Sprintf("%.3fB", bytes)
	default:
		return "0B"
	}
}

// GetClientAddr ...
func GetClientAddr(req *http.Request) string {
	if addr := req.Header.Get("HTTP_CLIENT_IP"); addr != "" {
		return addr
	} else if addr := req.Header.Get("HTTP_X_FORWARDED_FOR"); addr != "" {
		return addr
	}
	return req.RemoteAddr
}

// GetFilesByDir ...
func GetFilesByDir(dir string) ([]string, error) {
	var fileList []string
	var walkFunc = func(path string, f os.FileInfo, err error) error {
		var strRet string
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		strRet += path
		fileList = append(fileList, strRet)
		return nil
	}
	err := filepath.Walk(dir, walkFunc)
	if err != nil {
		return nil, err
	}
	return fileList, nil
}

// GetURILastIndex ...
func GetURILastIndex(uri string) (string, error) {
	ruri, err := url.ParseRequestURI(uri)
	if err != nil {
		return "", err
	}
	idx := strings.LastIndex(ruri.Path, "/")
	if idx > -1 && idx+1 < len(ruri.Path) {
		return ruri.Path[idx+1:], nil
	}
	return "", nil
}
