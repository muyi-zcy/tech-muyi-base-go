package myUtils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		return os.IsExist(err)
	}
	return true
}

func DirectoryExists(dirPath string) bool {
	if s, err := os.Stat(dirPath); err != nil {
		return false
	} else {
		return s.IsDir()
	}
}

func ExtractFilePath(filePath string) string {
	idx := strings.LastIndex(filePath, string(os.PathSeparator))
	return filePath[:idx]
}

func ExtractFileName(filePath string) string {
	idx := strings.LastIndex(filePath, string(os.PathSeparator))
	return filePath[idx+1:]
}

func ExtractFileExt(filePath string) string {
	idx := strings.LastIndex(filePath, ".")
	if idx != -1 {
		return filePath[idx+1:]
	}
	return ""
}

func ChangeFileExt(filePath string, ext string) string {
	mext := ExtractFileExt(filePath)
	if mext == "" {
		return filePath + "." + ext
	} else {
		return filePath[:len(filePath)-len(mext)] + ext
	}
}

func MkDirs(path string) bool {
	if !DirectoryExists(path) {
		return os.MkdirAll(path, os.ModePerm) == nil
	} else {
		return false
	}
}

func DeleteDirs(path string) bool {
	return os.RemoveAll(path) == nil
}

func DeleteFile(filePath string) bool {
	return os.Remove(filePath) == nil
}

func ReadFile(filePath string) string {
	var ret = ""
	if b, err := os.ReadFile(filePath); err == nil {
		ret = string(b)
	}
	return ret
}

func ReadFileBytes(filePath string) []byte {
	var ret []byte
	if b, err := os.ReadFile(filePath); err == nil {
		ret = b
	}
	return ret
}

func ReadFileLines(filePath string) []string {
	var ret []string
	if b, err := os.ReadFile(filePath); err == nil {
		ret = strings.Split(string(b), "\n")
	}
	return ret
}

func WriteFile(filePath string, text string) bool {
	return WriteFileBytes(filePath, []byte(text))
}

func WriteFileBytes(filePath string, data []byte) bool {
	p0 := ExtractFilePath(filePath)
	if !DirectoryExists(p0) {
		MkDirs(p0)
	}
	if FileExists(filePath) {
		DeleteFile(filePath)
	}
	if fl, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return false
	} else {
		_, err := fl.Write(data)
		_ = fl.Close()
		return err == nil
	}
}

func AppendFile(filePath string, text string) bool {
	return AppendFileBytes(filePath, []byte(text))
}

func AppendFileBytes(filePath string, data []byte) bool {
	p0 := ExtractFilePath(filePath)
	if !DirectoryExists(p0) {
		MkDirs(p0)
	}
	if fl, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return false
	} else {
		_, err := fl.Write(data)
		_ = fl.Close()
		return err == nil
	}
}

// Deprecated:this function is replaced by CopyFileWithError, which will return the reason for copy failure.
func CopyFile(srcFilePath string, destFilePath string) bool {
	p0 := ExtractFilePath(destFilePath)
	if !DirectoryExists(p0) {
		MkDirs(p0)
	}
	src, _ := os.Open(srcFilePath)
	defer func(src *os.File) { _ = src.Close() }(src)
	dst, _ := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	defer func(dst *os.File) { _ = dst.Close() }(dst)
	_, err := io.Copy(dst, src)
	return err == nil
}

func RenameFile(srcFilePath string, destFilePath string) bool {
	p0 := ExtractFilePath(destFilePath)
	if !DirectoryExists(p0) {
		MkDirs(p0)
	}
	return os.Rename(srcFilePath, destFilePath) == nil
}

func CreateFile(filePath string) bool {
	if FileExists(filePath) {
		return true
	}

	p0 := ExtractFilePath(filePath)
	if !DirectoryExists(p0) {
		MkDirs(p0)
	}

	if _, err := os.OpenFile(filePath, os.O_CREATE, 0644); err != nil {
		return false
	} else {
		return true
	}
}

func Child(filePath string) ([]os.DirEntry, error) {
	return os.ReadDir(filePath)
}

// Size 返回文件/目录的大小
func Size(filePath string) int64 {
	if !DirectoryExists(filePath) {
		fi, err := os.Stat(filePath)
		if err == nil {
			return fi.Size()
		}
		return 0
	} else {
		var size int64
		err := filepath.Walk(filePath, func(_ string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				size += info.Size()
			}
			return err
		})
		if err != nil {
			return 0
		}
		return size
	}
}

func Sizes(filePaths ...string) int64 {
	var count int64
	for _, filePath := range filePaths {
		count += Size(filePath)
	}
	return count
}

func SizeList(filePaths []string) int64 {
	var count int64
	for _, filePath := range filePaths {
		count += Size(filePath)
	}
	return count
}

func SizeFormat(filePath string) string {
	return FormatSize(Size(filePath))
}

func FormatSize(fileSize int64) (size string) {
	if fileSize < KB {
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(B))
	} else if fileSize < MB {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(KB))
	} else if fileSize < GB {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(MB))
	} else if fileSize < TB {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(GB))
	} else if fileSize < PB {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(TB))
	} else if fileSize < EB {
		return fmt.Sprintf("%.2fPB", float64(fileSize)/float64(PB))
	} else {
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(EB))
	}
}
func CopyFileWithError(srcFilePath string, destFilePath string) error {
	dirPath := ExtractFilePath(destFilePath)
	var err error
	if !DirectoryExists(dirPath) {
		err = os.MkdirAll(dirPath, os.ModePerm)
	}
	if err != nil {
		return err
	}
	src, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer func(src *os.File) { _ = src.Close() }(src)
	dst, err := os.OpenFile(destFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer func(dst *os.File) { _ = dst.Close() }(dst)
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
	PB = 1024 * TB
	EB = 1024 * PB
)
