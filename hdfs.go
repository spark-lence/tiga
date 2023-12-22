package tiga

import (
	"fmt"
	"path/filepath"

	"github.com/colinmarc/hdfs/v2"
)

type HdfsDao struct {
	client *hdfs.Client
}

func NewHdfsDao(addr string) *HdfsDao {
	client, err := hdfs.New(addr)
	if err != nil {
		panic(err)
	}
	return &HdfsDao{
		client: client,
	}
}

func (h *HdfsDao) Put(src string, dst string) error {
	dir := filepath.Dir(dst)
	err := h.client.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("创建目录失败:%w", err)
	}
	return h.client.CopyToRemote(src, dst)
}
func (h *HdfsDao) Get(src string, dst string) error {
	return h.client.CopyToLocal(src, dst)
}
func (h *HdfsDao) Delete(src string) error {
	return h.client.Remove(src)
}
func (h *HdfsDao) Mkdir(src string) error {
	return h.client.Mkdir(src, 0755)
}
