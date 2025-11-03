package storage

import (
	"log"
	"os"
	"path/filepath"
)

// FileStore persists documents to disk.
type FileStore struct {
	FilePath string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{FilePath: filepath.Clean(path)}
}

// Read returns the current document content.
func (s *FileStore) Read() string {
	data, err := os.ReadFile(s.FilePath)
	if err != nil {
		log.Fatalf("error reading the file")
		return ""
	}
	return string(data)
}

// Write persists the provided document content.
func (s *FileStore) Write(value []byte) error {
	log.Printf("writing to file %s", s.FilePath)
	err := os.WriteFile(s.FilePath, value, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
