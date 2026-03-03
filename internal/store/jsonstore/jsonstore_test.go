package jsonstore

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jimschubert/rumor/internal/store"
	"github.com/stretchr/testify/assert"
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

		assert.NoError(t, err, "unexpected error")
		assert.NotNil(t, s, "expected non-nil store")
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
		assert.NoError(t, err, "unexpected error")

		rec, err := s.Get("users", "1")
		assert.NoError(t, err, "failed to get record")
		assert.Equal(t, "Alice", rec["name"], "expected name=Alice")
	})

	t.Run("handles corrupted JSON in file", func(t *testing.T) {
		mock := newMockFS()
		mock.files["test.json"] = []byte("invalid json")

		_, err := NewWithFS("test.json", mock)
		assert.Error(t, err, "expected error for corrupted JSON")
	})
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		data         store.Record
		expectID     int64
		expectFields map[string]any
	}{
		{
			name:     "creates record with sequential ID",
			resource: "users",
			data: store.Record{
				"name":  "Alice",
				"email": "alice@example.com",
			},
			expectID: 1,
			expectFields: map[string]any{
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
			expectFields: map[string]any{
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
			assert.NoError(t, err, "unexpected error")

			assert.Equal(t, tt.expectID, result["id"], "expected id=%d", tt.expectID)

			for key, expectedVal := range tt.expectFields {
				assert.Equal(t, expectedVal, result[key], "expected %s=%v", key, expectedVal)
			}

			assert.NotNil(t, result["createdAt"], "expected createdAt field to be set")
		})
	}

	t.Run("creates multiple records with incremented IDs in same store", func(t *testing.T) {
		mock := newMockFS()
		s, _ := NewWithFS("test.json", mock)

		user1, _ := s.Create("users", store.Record{"name": "Alice"})
		user2, _ := s.Create("users", store.Record{"name": "Bob"})
		user3, _ := s.Create("users", store.Record{"name": "Charlie"})

		assert.Equal(t, int64(1), user1["id"], "expected user1 id=1")
		assert.Equal(t, int64(2), user2["id"], "expected user2 id=2")
		assert.Equal(t, int64(3), user3["id"], "expected user3 id=3")
	})
}

func TestGet(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		recordID     string
		initialData  map[string][]store.Record
		expectFound  bool
		expectFields map[string]any
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
			expectFields: map[string]any{
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
				assert.NoError(t, err, "unexpected error")
				for key, expectedVal := range tt.expectFields {
					assert.Equal(t, expectedVal, result[key], "expected %s=%v", key, expectedVal)
				}
			} else {
				assert.Error(t, err, "expected error")
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

			assert.NoError(t, err, "unexpected error")
			assert.Len(t, results, tt.expectCount, "expected %d results", tt.expectCount)
			assert.Equal(t, tt.expectTotal, total, "expected total=%d", tt.expectTotal)

			for i, expectedID := range tt.expectIDs {
				assert.Greater(t, len(results), i, "expected at least %d results", i+1)
				assert.Equal(t, float64(expectedID), results[i]["id"], "expected id=%d at index %d", expectedID, i)
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
		expectFields map[string]any
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
			expectFields: map[string]any{
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
			expectFields: map[string]any{
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
				assert.NoError(t, err, "unexpected error")

				for key, expectedVal := range tt.expectFields {
					assert.Equal(t, expectedVal, result[key], "expected %s=%v", key, expectedVal)
				}

				assert.NotNil(t, result["updatedAt"], "expected updatedAt field to be set")
			} else {
				assert.Error(t, err, "expected error")
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
		expectFields  map[string]any
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
			expectFields: map[string]any{
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
			expectFields: map[string]any{
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
			expectFields: map[string]any{
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
			expectFields: map[string]any{
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
				assert.NoError(t, err, "unexpected error")

				for key, expectedVal := range tt.expectFields {
					assert.Equal(t, expectedVal, result[key], "expected %s=%v", key, expectedVal)
				}

				assert.NotNil(t, result["updatedAt"], "expected updatedAt field to be set")
			} else {
				assert.Error(t, err, "expected error")
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
				assert.NoError(t, err, "unexpected error")

				remaining, _, _ := s.List(tt.resource, nil, 1, 0)
				assert.Len(t, remaining, tt.expectRemaining, "expected %d remaining records", tt.expectRemaining)
			} else {
				assert.Error(t, err, "expected error")
			}
		})
	}
}

func TestPersistence(t *testing.T) {
	t.Run("persists data to filesystem on create", func(t *testing.T) {
		mock := newMockFS()
		s, _ := NewWithFS("test.json", mock)

		_, err := s.Create("users", store.Record{"name": "Alice"})
		assert.NoError(t, err, "unexpected error")
		assert.NotEmpty(t, mock.files["test.json"], "expected data to be persisted to file")
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
		assert.NoError(t, err, "unexpected error")

		var persisted map[string][]store.Record
		err = json.Unmarshal(mock.files["test.json"], &persisted)
		assert.NoError(t, err, "failed to unmarshal persisted data")
		assert.Equal(t, "Alice Smith", persisted["users"][0]["name"], "expected persisted name=Alice Smith")
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
		assert.NoError(t, err, "unexpected error")

		var persisted map[string][]store.Record
		err = json.Unmarshal(mock.files["test.json"], &persisted)
		assert.NoError(t, err, "failed to unmarshal persisted data")
		assert.Len(t, persisted["users"], 1, "expected 1 remaining record")
	})

	t.Run("skips persistence when file path is empty", func(t *testing.T) {
		mock := newMockFS()
		s, _ := NewWithFS("", mock)

		_, err := s.Create("users", store.Record{"name": "Alice"})
		assert.NoError(t, err, "unexpected error")
		assert.Empty(t, mock.files, "expected no file writes when path is empty")
	})
}

func TestIDSequencing(t *testing.T) {
	t.Run("tracks separate ID sequences per resource", func(t *testing.T) {
		mock := newMockFS()
		s, _ := NewWithFS("test.json", mock)

		user, err := s.Create("users", store.Record{"name": "Alice"})
		assert.NoError(t, err, "unexpected error")
		post, err := s.Create("posts", store.Record{"title": "Hello"})
		assert.NoError(t, err, "unexpected error")
		user2, err := s.Create("users", store.Record{"name": "Bob"})
		assert.NoError(t, err, "unexpected error")

		assert.Equal(t, int64(1), user["id"], "expected user id=1")
		assert.Equal(t, int64(1), post["id"], "expected post id=1")
		assert.Equal(t, int64(2), user2["id"], "expected user2 id=2")
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
		assert.Equal(t, int64(6), newUser["id"], "expected new user id=6")
	})
}

func TestTimestamps(t *testing.T) {
	t.Run("sets createdAt on create", func(t *testing.T) {
		mock := newMockFS()
		s, _ := NewWithFS("test.json", mock)

		before := time.Now().UTC()
		result, err := s.Create("users", store.Record{"name": "Alice"})
		assert.NoError(t, err, "unexpected error")
		after := time.Now().UTC().Add(1 * time.Second)

		createdAt, ok := result["createdAt"].(string)
		assert.True(t, ok, "expected createdAt to be a string")

		createdTime, err := time.Parse(time.RFC3339, createdAt)
		assert.NoError(t, err, "failed to parse createdAt")
		assert.True(t, createdTime.After(before.Add(-1*time.Second)), "createdAt should be after before-1s")
		assert.True(t, createdTime.Before(after), "createdAt should be before after")
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
		assert.True(t, ok, "expected updatedAt to be a string")

		updatedTime, err := time.Parse(time.RFC3339, updatedAt)
		assert.NoError(t, err, "failed to parse updatedAt")
		assert.True(t, updatedTime.After(before.Add(-1*time.Second)), "updatedAt should be after before-1s")
		assert.True(t, updatedTime.Before(after), "updatedAt should be before after")
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

		assert.NotNil(t, result["updatedAt"], "expected updatedAt to be set")
	})
}
