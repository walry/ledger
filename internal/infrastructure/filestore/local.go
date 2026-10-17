// Package filestore 提供一个基于本地磁盘的简单文件存储实现，
// 用于保存上传的报告原始文件，供后续解析读取。
package filestore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// LocalStore 将文件保存在本地目录下。
type LocalStore struct {
	dir string
}

// NewLocalStore 创建本地文件存储。
func NewLocalStore(dir string) (*LocalStore, error) {
	if strings.TrimSpace(dir) == "" {
		dir = filepath.Join(os.TempDir(), "ledger-uploads")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("filestore: create dir: %w", err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("filestore: resolve dir: %w", err)
	}
	return &LocalStore{dir: abs}, nil
}

// Save 保存文件并返回其存储路径。
func (s *LocalStore) Save(_ context.Context, name string, data []byte) (string, error) {
	ext := filepath.Ext(name)
	key := uuid.NewString() + ext
	full := filepath.Join(s.dir, key)
	if err := os.WriteFile(full, data, 0o644); err != nil {
		return "", fmt.Errorf("filestore: write: %w", err)
	}
	return full, nil
}

// Load 按存储路径读取文件内容。路径必须位于存储目录之内。
func (s *LocalStore) Load(_ context.Context, path string) ([]byte, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("filestore: resolve path: %w", err)
	}
	if abs != s.dir && !strings.HasPrefix(abs, s.dir+string(os.PathSeparator)) {
		return nil, errors.New("filestore: path escapes storage directory")
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("filestore: read: %w", err)
	}
	return data, nil
}

// Dir 返回存储根目录。
func (s *LocalStore) Dir() string { return s.dir }
