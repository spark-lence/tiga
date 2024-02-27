package tiga

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func Decompress(compressFile string, destDir string) (error) {

	file, err := os.Open(compressFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// 读取文件头部的几个字节来判断文件类型
	buff := make([]byte, 512) // 足够大以读取魔数
	_, err = file.Read(buff)
	if err != nil {
		return err
	}

	// 重置读取指针到文件开头
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}
	fileType := http.DetectContentType(buff)
	switch {
	case fileType == "application/zip":
		return Unzip(file, destDir)
	case fileType == "application/x-gzip":
		return UnTarGz(file, destDir)
	default:
		return fmt.Errorf("unsupported file type")
	}
}

// 解压缩zip文件
func Unzip(file *os.File, destDir string) error {
	reader, err := zip.OpenReader(file.Name())
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, f := range reader.File {
		filePath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		srcFile, err := f.Open()
		if err != nil {
			dstFile.Close()
			return err
		}

		_, err = io.Copy(dstFile, srcFile)

		dstFile.Close()
		srcFile.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// 解压缩tar.gz文件
func UnTarGz(file *os.File, destDir string) error {
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		filePath := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}
