package jsonstore

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jimschubert/rumor/internal/store"
)

var (
	// compile-time verification that JSONStore implements the store.Store interface
	_ store.Store = (*JSONStore)(nil)
)

// FileSystem abstracts file operations for testability, without using afero or other dependency
type FileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
}

// OSFileSystem implements FileSystem using actual os.* functions
type OSFileSystem struct{}

func (OSFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (OSFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// JSONStore represents a file-based datastore of JSON objects/records
type JSONStore struct {
	mu       sync.RWMutex
	data     map[string][]store.Record
	idSeq    map[string]int64
	filePath string
	fs       FileSystem
}

// New JSON "database" at filePath
// If the file does not exist, it will be created on the first persistence operation.
func New(filePath string) (*JSONStore, error) {
	return NewWithFS(filePath, OSFileSystem{})
}

// NewWithFS is an internal-only way to mock out os.* operations without actually reading/writing (or using afero)
func NewWithFS(filePath string, fs FileSystem) (*JSONStore, error) {
	s := &JSONStore{
		data:     make(map[string][]store.Record),
		idSeq:    make(map[string]int64),
		filePath: filePath,
		fs:       fs,
	}
	return s, s.load()
}

// List returns 0..N records of type resource, with optional filter and pagination.
func (s *JSONStore) List(resource string, filters map[string]string, page, pageSize int) ([]store.Record, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records, ok := s.data[resource]
	if !ok {
		return nil, 0, fmt.Errorf("unknown resource %q", resource)
	}

	var out []store.Record
	for _, r := range records {
		if matchesFilters(r, filters) {
			out = append(out, r)
		}
	}

	total := len(out)

	if pageSize > 0 {
		if page < 1 {
			page = 1
		}
		start := (page - 1) * pageSize
		if start >= total {
			return []store.Record{}, total, nil
		}
		end := min(start+pageSize, total)
		out = out[start:end]
	}

	return out, total, nil
}

// Get a resource by id
func (s *JSONStore) Get(resource, id string) (store.Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, r := range s.data[resource] {
		if recordMatchesID(r, id) {
			return r, nil
		}
	}
	return nil, fmt.Errorf("%s/%s not found", resource, id)
}

// Create a resource
func (s *JSONStore) Create(resource string, data store.Record) (store.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.idSeq[resource]++
	data["id"] = s.idSeq[resource]
	data["createdAt"] = time.Now().UTC().Format(time.RFC3339)
	s.data[resource] = append(s.data[resource], data)
	return data, s.persist()
}

// Update a resource with the given id
func (s *JSONStore) Update(resource, id string, data store.Record) (store.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, r := range s.data[resource] {
		if recordMatchesID(r, id) {
			data["id"] = r["id"]
			data["createdAt"] = r["createdAt"]
			data["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
			s.data[resource][i] = data
			return data, s.persist()
		}
	}
	return nil, fmt.Errorf("%s/%s not found", resource, id)
}

// Patch a resource with the given id
func (s *JSONStore) Patch(resource, id string, patch store.Record) (store.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, r := range s.data[resource] {
		if recordMatchesID(r, id) {
			for k, v := range patch {
				if k != "id" && k != "createdAt" {
					r[k] = v
				}
			}
			r["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
			s.data[resource][i] = r
			return r, s.persist()
		}
	}
	return nil, fmt.Errorf("%s/%s not found", resource, id)
}

// Delete a specific resource by id
func (s *JSONStore) Delete(resource, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, r := range s.data[resource] {
		if recordMatchesID(r, id) {
			s.data[resource] = append(s.data[resource][:i], s.data[resource][i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("%s/%s not found", resource, id)
}

// load reads the data representation from the file system
func (s *JSONStore) load() error {
	raw, err := s.fs.ReadFile(s.filePath)
	if os.IsNotExist(err) {
		// nbd, file will be created on save
		return nil
	} else if err != nil {
		return err
	}

	if err = json.Unmarshal(raw, &s.data); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}
	for resource, records := range s.data {
		for _, r := range records {
			// track current ID for the resource
			if id, ok := r["id"].(float64); ok && int64(id) > s.idSeq[resource] {
				s.idSeq[resource] = int64(id)
			}
		}
	}
	return nil
}

// persist saves the data representation to the file system
func (s *JSONStore) persist() error {
	if s.filePath == "" {
		return nil
	}
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	return s.fs.WriteFile(s.filePath, raw, 0644)
}

// matchesFilters checks if a record matches the filters defined in the filter set
// these filters currently match at the top-level keys of the object _only_.
func matchesFilters(r store.Record, filters map[string]string) bool {
	for k, v := range filters {
		if fmt.Sprintf("%v", r[k]) != v {
			return false
		}
	}
	return true
}

// recordMatchesID checks if a record's ID matches the given string ID
func recordMatchesID(r store.Record, id string) bool {
	return fmt.Sprintf("%v", r["id"]) == id
}
