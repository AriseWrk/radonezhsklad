package repository

import (
"encoding/json"
"errors"
"fmt"
"os"
"path/filepath"
"sort"
"strings"
"sync"
"time"

"github.com/google/uuid"
)

type FileMeta struct {
Name           string     `json:"name"`
Category       string     `json:"category"`
SizeBytes      int64      `json:"size_bytes"`
UploadedBy     *uuid.UUID `json:"uploaded_by,omitempty"`
UploadedByName string     `json:"uploaded_by_name,omitempty"`
CreatedAt      time.Time  `json:"created_at"`
}

type FileRepo struct {
dir string
mu  sync.Mutex
}

func NewFileRepo(dir string) (*FileRepo, error) {
if err := os.MkdirAll(dir, 0o755); err != nil {
return nil, err
}
return &FileRepo{dir: dir}, nil
}

func (r *FileRepo) metaPath() string { return filepath.Join(r.dir, ".meta.json") }

func (r *FileRepo) readMeta() (map[string]FileMeta, error) {
out := map[string]FileMeta{}
b, err := os.ReadFile(r.metaPath())
if errors.Is(err, os.ErrNotExist) {
return out, nil
}
if err != nil {
return nil, err
}
_ = json.Unmarshal(b, &out)
return out, nil
}

func (r *FileRepo) writeMeta(m map[string]FileMeta) error {
b, _ := json.MarshalIndent(m, "", "  ")
return os.WriteFile(r.metaPath(), b, 0o644)
}

func (r *FileRepo) List() ([]FileMeta, error) {
r.mu.Lock()
defer r.mu.Unlock()

meta, err := r.readMeta()
if err != nil {
return nil, err
}
entries, err := os.ReadDir(r.dir)
if err != nil {
return nil, err
}
out := []FileMeta{}
for _, e := range entries {
if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
continue
}
info, err := e.Info()
if err != nil {
continue
}
m, ok := meta[e.Name()]
if !ok {
m = FileMeta{Name: e.Name(), CreatedAt: info.ModTime()}
}
m.SizeBytes = info.Size()
out = append(out, m)
}
sort.Slice(out, func(i, j int) bool {
return out[i].CreatedAt.After(out[j].CreatedAt)
})
return out, nil
}

func (r *FileRepo) Save(name, category string, content []byte, userID *uuid.UUID, userName string) (FileMeta, error) {
r.mu.Lock()
defer r.mu.Unlock()

name = strings.TrimSpace(name)
if name == "" {
return FileMeta{}, fmt.Errorf("empty name")
}
if strings.ContainsAny(name, `/\`) || strings.HasPrefix(name, ".") {
return FileMeta{}, fmt.Errorf("invalid file name")
}
full := filepath.Join(r.dir, name)
if err := os.WriteFile(full, content, 0o644); err != nil {
return FileMeta{}, err
}
meta, _ := r.readMeta()
if meta == nil {
meta = map[string]FileMeta{}
}
m := FileMeta{
Name:           name,
Category:       category,
SizeBytes:      int64(len(content)),
UploadedBy:     userID,
UploadedByName: userName,
CreatedAt:      time.Now(),
}
meta[name] = m
_ = r.writeMeta(meta)
return m, nil
}

func (r *FileRepo) Get(name string) (string, error) {
if strings.ContainsAny(name, `/\`) || strings.HasPrefix(name, ".") {
return "", fmt.Errorf("invalid file name")
}
full := filepath.Join(r.dir, name)
if _, err := os.Stat(full); err != nil {
return "", err
}
return full, nil
}

func (r *FileRepo) Delete(name string) error {
r.mu.Lock()
defer r.mu.Unlock()

if strings.ContainsAny(name, `/\`) || strings.HasPrefix(name, ".") {
return fmt.Errorf("invalid file name")
}
full := filepath.Join(r.dir, name)
if err := os.Remove(full); err != nil && !errors.Is(err, os.ErrNotExist) {
return err
}
meta, _ := r.readMeta()
if meta != nil {
delete(meta, name)
_ = r.writeMeta(meta)
}
return nil
}
// UpdateCategory — обновляет категорию файла в meta.json.
// Если записи нет — создаёт с этим файлом и текущим временем.
func (r *FileRepo) UpdateCategory(name, category string) error {
r.mu.Lock()
defer r.mu.Unlock()

if strings.ContainsAny(name, `/\`) || strings.HasPrefix(name, ".") {
return fmt.Errorf("invalid file name")
}
full := filepath.Join(r.dir, name)
info, err := os.Stat(full)
if err != nil {
return err
}

meta, _ := r.readMeta()
if meta == nil {
meta = map[string]FileMeta{}
}
m, ok := meta[name]
if !ok {
m = FileMeta{Name: name, CreatedAt: info.ModTime()}
}
m.Category = category
m.SizeBytes = info.Size()
meta[name] = m
return r.writeMeta(meta)
}