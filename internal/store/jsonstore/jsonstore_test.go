package jsonstore

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jimschubert/rumor/internal/store"
)

// MockFileSystem implements FileSystem for testing without actual file operations
type MockFileSystem struct {
	files map[string][]byte
	err   error
}

func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	data, ok := m.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (m *MockFileSystem) WriteFile(name string, data []byte, _ os.FileMode) error {
	if m.err != nil {
		return m.err
	}
	if m.files == nil {
		m.files = make(map[string][]byte)
	}
	m.files[name] = data
	return nil
}

func newMockFS() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
	}
}

func TestNew(t *testing.T) {
	t.Run("creates empty store when file does not exist", func(t *testing.T) {
		mock := newMockFS()
		s, err := NewWithFS("test.json", mock)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s == nil {
			t.Fatalf("expected non-nil store")
		}
	})

	t.Run("loads existing data from file", func(t *testing.T) {
		initialData := map[string][]store.Record{
			"users": {
				{
					"id":        float64(1),
					"name":      "Alice",
					"createdAt": "2026-03-01T00:00:00Z",
				},
			},
		}
		raw, _ := json.Marshal(initialData)

		mock := newMockFS()
		mock.files["test.json"] = raw

		s, err := NewWithFS("test.json", mock)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify data was loaded by attempting to retrieve the record
		rec, err := s.Get("users", "1")
		if err != nil {
			t.Fatalf("failed to get record: %v", err)
		}
		if rec["name"] != "Alice" {
			t.Errorf("expected name=Alice, got %v", rec["name"])
		}
	})

	t.Run("handles corrupted JSON in file", func(t *testing.T) {
		mock := newMockFS()
		mock.files["test.json"] = []byte("invalid json")

		_, err := NewWithFS("test.json", mock)
		if err == nil {
			t.Errorf("expected error for corrupted JSON")
		}
	})
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		data         store.Record
		expectID     int64
		expectFields map[string]interface{}
	}{
		{
			name:     "creates record with sequential ID",
			resource: "users",
			data: store.Record{
				"name":  "Alice",
				"email": "alice@example.com",
			},
			expectID: 1,
			expectFields: map[string]interface{}{
				"name":  "Alice",
				"email": "alice@example.com",
			},
		},
		{
			name:     "creates record in different resource",
			resource: "posts",
			data: store.Record{
				"title": "Hello World",
				"body":  "This is a test post",
			},
			expectID: 1,
			expectFields: map[string]interface{}{
				"title": "Hello World",
				"body":  "This is a test post",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockFS()
			s, _ := NewWithFS("test.json", mock)

			result, err := s.Create(tt.resource, tt.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result["id"] != tt.expectID {
				t.Errorf("expected id=%d, got %v", tt.expectID, result["id"])
			}

			for key, expectedVal := range tt.expectFields {
				if result[key] != expectedVal {
					t.Errorf("expected %s=%v, got %v", key, expectedVal, result[key])
				}
			}

			if _, ok := result["createdAt"]; !ok {
				t.Errorf("expected createdAt field to be set")
			}
		})
	}

	t.Run("creates multiple records with incremented IDs in same store", func(t *testing.T) {
		mock := newMockFS()
		s, _ := NewWithFS("test.json", mock)

		user1, _ := s.Create("users", store.Record{"name": "Alice"})
		user2, _ := s.Create("users", store.Record{"name": "Bob"})
		user3, _ := s.Create("users", store.Record{"name": "Charlie"})

		if user1["id"] != int64(1) {
			t.Errorf("expected user1 id=1, got %v", user1["id"])
		}
		if user2["id"] != int64(2) {
			t.Errorf("expected user2 id=2, got %v", user2["id"])
		}
		if user3["id"] != int64(3) {
			t.Errorf("expected user3 id=3, got %v", user3["id"])
		}
	})
}

func TestGet(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		recordID     string
		initialData  map[string][]store.Record
		expectFound  bool
		expectFields map[string]interface{}
	}{
		{
			name:     "gets existing record",
			resource: "users",
			recordID: "1",
			initialData: map[string][]store.Record{
				"users": {
					{
						"id":   float64(1),
						"name": "Alice",
					},
				},
			},
			expectFound: true,
			expectFields: map[string]interface{}{
				"id":   float64(1),
				"name": "Alice",
			},
		},
		{
			name:     "returns error for non-existent record",
			resource: "users",
			recordID: "999",
			initialData: map[string][]store.Record{
				"users": {
					{
						"id":   float64(1),
						"name": "Alice",
					},
				},
			},
			expectFound: false,
		},
		{
			name:        "returns error for non-existent resource",
			resource:    "posts",
			recordID:    "1",
			initialData: map[string][]store.Record{},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, _ := json.Marshal(tt.initialData)
			mock := newMockFS()
			mock.files["test.json"] = raw

			s, _ := NewWithFS("test.json", mock)

			result, err := s.Get(tt.resource, tt.recordID)

			if tt.expectFound {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				for key, expectedVal := range tt.expectFields {
					if result[key] != expectedVal {
						t.Errorf("expected %s=%v, got %v", key, expectedVal, result[key])
					}
				}
			} else {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
			}
		})
	}
}

func TestList(t *testing.T) {
	initialData := map[string][]store.Record{
		"users": {
			{
				"id":     float64(1),
				"name":   "Alice",
				"status": "active",
			},
			{
				"id":     float64(2),
				"name":   "Bob",
				"status": "active",
			},
			{
				"id":     float64(3),
				"name":   "Charlie",
				"status": "inactive",
			},
		},
	}

	tests := []struct {
		name        string
		resource    string
		filters     map[string]string
		page        int
		pageSize    int
		expectCount int
		expectTotal int
		expectIDs   []int64
	}{
		{
			name:        "lists all records without filters",
			resource:    "users",
			filters:     nil,
			page:        1,
			pageSize:    0,
			expectCount: 3,
			expectTotal: 3,
			expectIDs:   []int64{1, 2, 3},
		},
		{
			name:     "filters records by status",
			resource: "users",
			filters: map[string]string{
				"status": "active",
			},
			page:        1,
			pageSize:    0,
			expectCount: 2,
			expectTotal: 2,
			expectIDs:   []int64{1, 2},
		},
		{
			name:        "paginates records with page size",
			resource:    "users",
			filters:     nil,
			page:        1,
			pageSize:    2,
			expectCount: 2,
			expectTotal: 3,
			expectIDs:   []int64{1, 2},
		},
		{
			name:        "retrieves second page",
			resource:    "users",
			filters:     nil,
			page:        2,
			pageSize:    2,
			expectCount: 1,
			expectTotal: 3,
			expectIDs:   []int64{3},
		},
		{
			name:        "returns empty for out-of-range page",
			resource:    "users",
			filters:     nil,
			page:        10,
			pageSize:    2,
			expectCount: 0,
			expectTotal: 3,
			expectIDs:   []int64{},
		},
		{
			name:        "defaults page to 1 when less than 1",
			resource:    "users",
			filters:     nil,
			page:        0,
			pageSize:    2,
			expectCount: 2,
			expectTotal: 3,
			expectIDs:   []int64{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, _ := json.Marshal(initialData)
			mock := newMockFS()
			mock.files["test.json"] = raw

			s, _ := NewWithFS("test.json", mock)

			results, total, err := s.List(tt.resource, tt.filters, tt.page, tt.pageSize)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != tt.expectCount {
				t.Errorf("expected %d results, got %d", tt.expectCount, len(results))
			}

			if total != tt.expectTotal {
				t.Errorf("expected total=%d, got %d", tt.expectTotal, total)
			}

			for i, expectedID := range tt.expectIDs {
				if i >= len(results) {
					t.Fatalf("expected at least %d results, got %d", i+1, len(results))
				}
				if results[i]["id"] != float64(expectedID) {
					t.Errorf("expected id=%d at index %d, got %v", expectedID, i, results[i]["id"])
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	initialData := map[string][]store.Record{
		"users": {
			{
				"id":        float64(1),
				"name":      "Alice",
				"email":     "alice@example.com",
				"createdAt": "2026-03-01T00:00:00Z",
			},
		},
	}

	tests := []struct {
		name         string
		resource     string
		recordID     string
		updateData   store.Record
		expectFound  bool
		expectFields map[string]interface{}
	}{
		{
			name:     "updates existing record",
			resource: "users",
			recordID: "1",
			updateData: store.Record{
				"name":  "Alice Smith",
				"email": "alice.smith@example.com",
			},
			expectFound: true,
			expectFields: map[string]interface{}{
				"id":    float64(1),
				"name":  "Alice Smith",
				"email": "alice.smith@example.com",
			},
		},
		{
			name:     "preserves original id and createdAt",
			resource: "users",
			recordID: "1",
			updateData: store.Record{
				"name": "Bob",
			},
			expectFound: true,
			expectFields: map[string]interface{}{
				"id": float64(1),
			},
		},
		{
			name:     "returns error for non-existent record",
			resource: "users",
			recordID: "999",
			updateData: store.Record{
				"name": "Ghost",
			},
			expectFound: false,
		},
	}

	//goland:noinspection DuplicatedCode
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, _ := json.Marshal(initialData)
			mock := newMockFS()
			mock.files["test.json"] = raw

			s, _ := NewWithFS("test.json", mock)

			result, err := s.Update(tt.resource, tt.recordID, tt.updateData)

			if tt.expectFound {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				for key, expectedVal := range tt.expectFields {
					if result[key] != expectedVal {
						t.Errorf("expected %s=%v, got %v", key, expectedVal, result[key])
					}
				}

				if _, ok := result["updatedAt"]; !ok {
					t.Errorf("expected updatedAt field to be set")
				}
			} else {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
			}
		})
	}
}

func TestPatch(t *testing.T) {
	initialData := map[string][]store.Record{
		"users": {
			{
				"id":        float64(1),
				"name":      "Alice",
				"email":     "alice@example.com",
				"age":       30,
				"createdAt": "2026-03-01T00:00:00Z",
			},
		},
	}

	tests := []struct {
		name          string
		resource      string
		recordID      string
		patchData     store.Record
		expectFound   bool
		expectFields  map[string]interface{}
		shouldNotHave []string
	}{
		{
			name:     "patches single field",
			resource: "users",
			recordID: "1",
			patchData: store.Record{
				"age": 31,
			},
			expectFound: true,
			expectFields: map[string]interface{}{
				"id":    float64(1),
				"name":  "Alice",
				"email": "alice@example.com",
				"age":   31,
			},
		},
		{
			name:     "patches multiple fields",
			resource: "users",
			recordID: "1",
			patchData: store.Record{
				"name":  "Alice Smith",
				"email": "asmith@example.com",
			},
			expectFound: true,
			expectFields: map[string]interface{}{
				"id":    float64(1),
				"name":  "Alice Smith",
				"email": "asmith@example.com",
			},
		},
		{
			name:     "prevents patching id field",
			resource: "users",
			recordID: "1",
			patchData: store.Record{
				"id": float64(999),
			},
			expectFound: true,
			expectFields: map[string]interface{}{
				"id": float64(1),
			},
		},
		{
			name:     "prevents patching createdAt field",
			resource: "users",
			recordID: "1",
			patchData: store.Record{
				"createdAt": "2026-01-01T00:00:00Z",
			},
			expectFound: true,
			expectFields: map[string]interface{}{
				"createdAt": "2026-03-01T00:00:00Z",
			},
		},
		{
			name:     "returns error for non-existent record",
			resource: "users",
			recordID: "999",
			patchData: store.Record{
				"age": 25,
			},
			expectFound: false,
		},
	}

	//goland:noinspection DuplicatedCode
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, _ := json.Marshal(initialData)
			mock := newMockFS()
			mock.files["test.json"] = raw

			s, _ := NewWithFS("test.json", mock)

			result, err := s.Patch(tt.resource, tt.recordID, tt.patchData)

			if tt.expectFound {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				for key, expectedVal := range tt.expectFields {
					if result[key] != expectedVal {
						t.Errorf("expected %s=%v, got %v", key, expectedVal, result[key])
					}
				}

				if _, ok := result["updatedAt"]; !ok {
					t.Errorf("expected updatedAt field to be set")
				}
			} else {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	initialData := map[string][]store.Record{
		"users": {
			{
				"id":   float64(1),
				"name": "Alice",
			},
			{
				"id":   float64(2),
				"name": "Bob",
			},
		},
	}

	tests := []struct {
		name            string
		resource        string
		recordID        string
		expectSuccess   bool
		expectRemaining int
	}{
		{
			name:            "deletes existing record",
			resource:        "users",
			recordID:        "1",
			expectSuccess:   true,
			expectRemaining: 1,
		},
		{
			name:          "returns error for non-existent record",
			resource:      "users",
			recordID:      "999",
			expectSuccess: false,
		},
		{
			name:          "returns error for non-existent resource",
			resource:      "posts",
			recordID:      "1",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, _ := json.Marshal(initialData)
			mock := newMockFS()
			mock.files["test.json"] = raw

			s, _ := NewWithFS("test.json", mock)

			err := s.Delete(tt.resource, tt.recordID)

			if tt.expectSuccess {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				// Verify record was actually deleted
				remaining, _, _ := s.List(tt.resource, nil, 1, 0)
				if len(remaining) != tt.expectRemaining {
					t.Errorf("expected %d remaining records, got %d", tt.expectRemaining, len(remaining))
				}
			} else {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
			}
		})
	}
}

func TestPersistence(t *testing.T) {
	t.Run("persists data to filesystem on create", func(t *testing.T) {
		mock := newMockFS()
		s, _ := NewWithFS("test.json", mock)

		_, err := s.Create("users", store.Record{"name": "Alice"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := mock.files["test.json"]; !ok {
			t.Errorf("expected data to be persisted to file")
		}
	})

	t.Run("persists data to filesystem on update", func(t *testing.T) {
		initialData := map[string][]store.Record{
			"users": {
				{
					"id":   float64(1),
					"name": "Alice",
				},
			},
		}
		raw, _ := json.Marshal(initialData)
		mock := newMockFS()
		mock.files["test.json"] = raw

		s, _ := NewWithFS("test.json", mock)
		_, err := s.Update("users", "1", store.Record{"name": "Alice Smith"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify the update was persisted
		var persisted map[string][]store.Record
		err = json.Unmarshal(mock.files["test.json"], &persisted)
		if err != nil {
			t.Fatalf("failed to unmarshal persisted data: %v", err)
		}

		if persisted["users"][0]["name"] != "Alice Smith" {
			t.Errorf("expected persisted name=Alice Smith, got %v", persisted["users"][0]["name"])
		}
	})

	t.Run("persists data to filesystem on delete", func(t *testing.T) {
		initialData := map[string][]store.Record{
			"users": {
				{
					"id":   float64(1),
					"name": "Alice",
				},
				{
					"id":   float64(2),
					"name": "Bob",
				},
			},
		}
		raw, _ := json.Marshal(initialData)
		mock := newMockFS()
		mock.files["test.json"] = raw

		s, _ := NewWithFS("test.json", mock)
		err := s.Delete("users", "1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var persisted map[string][]store.Record
		err = json.Unmarshal(mock.files["test.json"], &persisted)
		if err != nil {
			t.Fatalf("failed to unmarshal persisted data: %v", err)
		}

		if len(persisted["users"]) != 1 {
			t.Errorf("expected 1 remaining record, got %d", len(persisted["users"]))
		}
	})

	t.Run("skips persistence when file path is empty", func(t *testing.T) {
		mock := newMockFS()
		s, _ := NewWithFS("", mock)

		_, err := s.Create("users", store.Record{"name": "Alice"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(mock.files) > 0 {
			t.Errorf("expected no file writes when path is empty")
		}
	})
}

func TestIDSequencing(t *testing.T) {
	t.Run("tracks separate ID sequences per resource", func(t *testing.T) {
		mock := newMockFS()
		s, _ := NewWithFS("test.json", mock)

		user, err := s.Create("users", store.Record{"name": "Alice"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		post, err := s.Create("posts", store.Record{"title": "Hello"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		user2, err := s.Create("users", store.Record{"name": "Bob"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if user["id"] != int64(1) {
			t.Errorf("expected user id=1, got %v", user["id"])
		}
		if post["id"] != int64(1) {
			t.Errorf("expected post id=1, got %v", post["id"])
		}
		if user2["id"] != int64(2) {
			t.Errorf("expected user2 id=2, got %v", user2["id"])
		}
	})

	t.Run("resumes ID sequence from loaded data", func(t *testing.T) {
		initialData := map[string][]store.Record{
			"users": {
				{
					"id":   float64(5),
					"name": "Alice",
				},
			},
		}
		raw, _ := json.Marshal(initialData)
		mock := newMockFS()
		mock.files["test.json"] = raw

		s, _ := NewWithFS("test.json", mock)
		newUser, _ := s.Create("users", store.Record{"name": "Bob"})

		if newUser["id"] != int64(6) {
			t.Errorf("expected new user id=6, got %v", newUser["id"])
		}
	})
}

func TestTimestamps(t *testing.T) {
	t.Run("sets createdAt on create", func(t *testing.T) {
		mock := newMockFS()
		s, _ := NewWithFS("test.json", mock)

		before := time.Now().UTC()
		result, err := s.Create("users", store.Record{"name": "Alice"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		after := time.Now().UTC().Add(1 * time.Second)

		createdAt, ok := result["createdAt"].(string)
		if !ok {
			t.Fatalf("expected createdAt to be a string")
		}

		createdTime, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			t.Fatalf("failed to parse createdAt: %v", err)
		}

		// RFC3339 has second precision, so just verify it's within a reasonable window
		if createdTime.Before(before.Add(-1*time.Second)) || createdTime.After(after) {
			t.Errorf("expected createdAt within range, got %v (before=%v, after=%v)", createdTime, before, after)
		}
	})

	t.Run("sets updatedAt on update", func(t *testing.T) {
		initialData := map[string][]store.Record{
			"users": {
				{
					"id":        float64(1),
					"name":      "Alice",
					"createdAt": "2026-03-01T00:00:00Z",
				},
			},
		}
		raw, _ := json.Marshal(initialData)
		mock := newMockFS()
		mock.files["test.json"] = raw

		s, _ := NewWithFS("test.json", mock)

		before := time.Now().UTC()
		result, _ := s.Update("users", "1", store.Record{"name": "Alice Smith"})
		after := time.Now().UTC().Add(1 * time.Second)

		updatedAt, ok := result["updatedAt"].(string)
		if !ok {
			t.Fatalf("expected updatedAt to be a string")
		}

		updatedTime, err := time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			t.Fatalf("failed to parse updatedAt: %v", err)
		}

		// RFC3339 has second precision, so just verify it's within a reasonable window
		if updatedTime.Before(before.Add(-1*time.Second)) || updatedTime.After(after) {
			t.Errorf("expected updatedAt within range, got %v (before=%v, after=%v)", updatedTime, before, after)
		}
	})

	t.Run("sets updatedAt on patch", func(t *testing.T) {
		initialData := map[string][]store.Record{
			"users": {
				{
					"id":        float64(1),
					"name":      "Alice",
					"createdAt": "2026-03-01T00:00:00Z",
				},
			},
		}
		raw, _ := json.Marshal(initialData)
		mock := newMockFS()
		mock.files["test.json"] = raw

		s, _ := NewWithFS("test.json", mock)

		result, _ := s.Patch("users", "1", store.Record{"name": "Alice Smith"})

		if _, ok := result["updatedAt"]; !ok {
			t.Errorf("expected updatedAt to be set")
		}
	})
}
