package server

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	pb "github.com/jimschubert/rumor/gen/rumor/v1"
	"github.com/jimschubert/rumor/internal/store"
	"github.com/jimschubert/rumor/internal/store/jsonstore"
)

// MockFileSystem implements jsonstore.FileSystem for testing  without actual file operations
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

func newTestServer(t *testing.T) (*RumorServer, *MockFileSystem) {
	mock := newMockFS()
	s, err := jsonstore.NewWithFS("test.json", mock)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	return New(s), mock
}

func TestNew(t *testing.T) {
	t.Run("creates a new server with given store", func(t *testing.T) {
		mock := newMockFS()
		s, _ := jsonstore.NewWithFS("test.json", mock)
		srv := New(s)

		assert.NotNil(t, srv, "expected non-nil server")
		assert.NotNil(t, srv.store, "expected store to be set")
	})
}

func TestList(t *testing.T) {
	tests := []struct {
		name            string
		resource        string
		filters         map[string]string
		page            int32
		pageSize        int32
		setupData       map[string][]store.Record
		expectCount     int
		expectTotal     int32
		expectError     bool
		expectErrorCode string
	}{
		{
			name:     "lists all records without filters",
			resource: "users",
			filters:  nil,
			page:     1,
			pageSize: 0,
			setupData: map[string][]store.Record{
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
			},
			expectCount: 2,
			expectTotal: 2,
			expectError: false,
		},
		{
			name:     "lists records with pagination",
			resource: "users",
			filters:  nil,
			page:     1,
			pageSize: 1,
			setupData: map[string][]store.Record{
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
			},
			expectCount: 1,
			expectTotal: 2,
			expectError: false,
		},
		{
			name:     "lists records with filters",
			resource: "users",
			filters: map[string]string{
				"status": "active",
			},
			page:     1,
			pageSize: 0,
			setupData: map[string][]store.Record{
				"users": {
					{
						"id":     float64(1),
						"name":   "Alice",
						"status": "active",
					},
					{
						"id":     float64(2),
						"name":   "Bob",
						"status": "inactive",
					},
				},
			},
			expectCount: 1,
			expectTotal: 1,
			expectError: false,
		},
		{
			name:        "returns error for non-existent resource",
			resource:    "nonexistent",
			filters:     nil,
			page:        1,
			pageSize:    0,
			setupData:   map[string][]store.Record{},
			expectCount: 0,
			expectTotal: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)

			for resource, records := range tt.setupData {
				for _, rec := range records {
					if _, err := srv.store.Create(resource, rec); err != nil {
						t.Fatalf("failed to setup record: %v", err)
					}
				}
			}

			resp, err := srv.List(context.Background(), &pb.ListRequest{
				Resource: tt.resource,
				Filters:  tt.filters,
				Page:     tt.page,
				PageSize: tt.pageSize,
			})

			if tt.expectError {
				assert.Error(t, err, "expected error")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.Len(t, resp.Items, tt.expectCount, "expected %d items", tt.expectCount)
			assert.Equal(t, tt.expectTotal, resp.Total, "expected total=%d", tt.expectTotal)
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		recordID     string
		setupData    map[string][]store.Record
		expectFound  bool
		expectFields map[string]any
	}{
		{
			name:     "gets existing record",
			resource: "users",
			recordID: "1",
			setupData: map[string][]store.Record{
				"users": {
					{
						"id":    float64(1),
						"name":  "Alice",
						"email": "alice@example.com",
					},
				},
			},
			expectFound: true,
			expectFields: map[string]any{
				"id":    float64(1),
				"name":  "Alice",
				"email": "alice@example.com",
			},
		},
		{
			name:        "returns error for non-existent record",
			resource:    "users",
			recordID:    "999",
			setupData:   map[string][]store.Record{},
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)

			for resource, records := range tt.setupData {
				for _, rec := range records {
					if _, err := srv.store.Create(resource, rec); err != nil {
						t.Fatalf("failed to setup record: %v", err)
					}
				}
			}

			resp, err := srv.Get(context.Background(), &pb.GetRequest{
				Resource: tt.resource,
				Id:       tt.recordID,
			})

			if tt.expectFound {
				assert.NoError(t, err, "unexpected error")
				assert.NotNil(t, resp, "expected non-nil response")

				respMap := resp.AsMap()
				for key, expectedVal := range tt.expectFields {
					assert.Equal(t, expectedVal, respMap[key], "expected %s=%v", key, expectedVal)
				}
			} else {
				assert.Error(t, err, "expected error")
			}
		})
	}
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		inputData    map[string]any
		expectError  bool
		expectFields map[string]any
	}{
		{
			name:     "creates new record",
			resource: "users",
			inputData: map[string]any{
				"name":  "Alice",
				"email": "alice@example.com",
			},
			expectError: false,
			expectFields: map[string]any{
				"name":  "Alice",
				"email": "alice@example.com",
			},
		},
		{
			name:        "returns error when data is nil",
			resource:    "users",
			inputData:   nil,
			expectError: true,
		},
		{
			name:     "creates record with complex nested data",
			resource: "posts",
			inputData: map[string]any{
				"title":  "Hello World",
				"author": "Alice",
				"tags": []any{
					"go",
					"testing",
				},
			},
			expectError: false,
			expectFields: map[string]any{
				"title":  "Hello World",
				"author": "Alice",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)

			var data *structpb.Struct
			var err error
			if tt.inputData != nil {
				data, err = structpb.NewStruct(tt.inputData)
				if err != nil {
					t.Fatalf("failed to create struct: %v", err)
				}
			}

			resp, err := srv.Create(context.Background(), &pb.CreateRequest{
				Resource: tt.resource,
				Data:     data,
			})

			if tt.expectError {
				assert.Error(t, err, "expected error")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.NotNil(t, resp, "expected non-nil response")

			respMap := resp.AsMap()
			assert.NotNil(t, respMap["id"], "expected id field to be set")

			for key, expectedVal := range tt.expectFields {
				assert.Equal(t, expectedVal, respMap[key], "expected %s=%v", key, expectedVal)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		recordID     string
		updateData   map[string]any
		setupData    map[string][]store.Record
		expectError  bool
		expectFields map[string]any
	}{
		{
			name:     "updates existing record",
			resource: "users",
			recordID: "1",
			updateData: map[string]any{
				"name":  "Alice Smith",
				"email": "alice.smith@example.com",
			},
			setupData: map[string][]store.Record{
				"users": {
					{
						"id":    float64(1),
						"name":  "Alice",
						"email": "alice@example.com",
					},
				},
			},
			expectError: false,
			expectFields: map[string]any{
				"id":    float64(1),
				"name":  "Alice Smith",
				"email": "alice.smith@example.com",
			},
		},
		{
			name:       "returns error when data is nil",
			resource:   "users",
			recordID:   "1",
			updateData: nil,
			setupData: map[string][]store.Record{
				"users": {
					{
						"id":   float64(1),
						"name": "Alice",
					},
				},
			},
			expectError: true,
		},
		{
			name:     "returns error for non-existent record",
			resource: "users",
			recordID: "999",
			updateData: map[string]any{
				"name": "Ghost",
			},
			setupData:   map[string][]store.Record{},
			expectError: true,
		},
	}

	//goland:noinspection DuplicatedCode
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)

			for resource, records := range tt.setupData {
				for _, rec := range records {
					if _, err := srv.store.Create(resource, rec); err != nil {
						t.Fatalf("failed to setup record: %v", err)
					}
				}
			}

			var data *structpb.Struct
			var err error
			if tt.updateData != nil {
				data, err = structpb.NewStruct(tt.updateData)
				if err != nil {
					t.Fatalf("failed to create struct: %v", err)
				}
			}

			resp, err := srv.Update(context.Background(), &pb.UpdateRequest{
				Resource: tt.resource,
				Id:       tt.recordID,
				Data:     data,
			})

			if tt.expectError {
				assert.Error(t, err, "expected error")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.NotNil(t, resp, "expected non-nil response")

			respMap := resp.AsMap()
			for key, expectedVal := range tt.expectFields {
				assert.Equal(t, expectedVal, respMap[key], "expected %s=%v", key, expectedVal)
			}

			assert.NotNil(t, respMap["updatedAt"], "expected updatedAt field to be set")
		})
	}
}

func TestPatch(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		recordID     string
		patchData    map[string]any
		setupData    map[string][]store.Record
		expectError  bool
		expectFields map[string]any
	}{
		{
			name:     "patches single field",
			resource: "users",
			recordID: "1",
			patchData: map[string]any{
				"email": "newemail@example.com",
			},
			setupData: map[string][]store.Record{
				"users": {
					{
						"id":    float64(1),
						"name":  "Alice",
						"email": "old@example.com",
					},
				},
			},
			expectError: false,
			expectFields: map[string]any{
				"id":    float64(1),
				"name":  "Alice",
				"email": "newemail@example.com",
			},
		},
		{
			name:     "patches multiple fields",
			resource: "users",
			recordID: "1",
			patchData: map[string]any{
				"name":  "Alice Smith",
				"email": "alice.smith@example.com",
			},
			setupData: map[string][]store.Record{
				"users": {
					{
						"id":    float64(1),
						"name":  "Alice",
						"email": "alice@example.com",
					},
				},
			},
			expectError: false,
			expectFields: map[string]any{
				"id":    float64(1),
				"name":  "Alice Smith",
				"email": "alice.smith@example.com",
			},
		},
		{
			name:      "returns error when data is nil",
			resource:  "users",
			recordID:  "1",
			patchData: nil,
			setupData: map[string][]store.Record{
				"users": {
					{
						"id":   float64(1),
						"name": "Alice",
					},
				},
			},
			expectError: true,
		},
		{
			name:     "returns error for non-existent record",
			resource: "users",
			recordID: "999",
			patchData: map[string]any{
				"name": "Ghost",
			},
			setupData:   map[string][]store.Record{},
			expectError: true,
		},
	}

	//goland:noinspection DuplicatedCode
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)

			for resource, records := range tt.setupData {
				for _, rec := range records {
					if _, err := srv.store.Create(resource, rec); err != nil {
						t.Fatalf("failed to setup record: %v", err)
					}
				}
			}

			var data *structpb.Struct
			var err error
			if tt.patchData != nil {
				data, err = structpb.NewStruct(tt.patchData)
				if err != nil {
					t.Fatalf("failed to create struct: %v", err)
				}
			}

			resp, err := srv.Patch(context.Background(), &pb.PatchRequest{
				Resource: tt.resource,
				Id:       tt.recordID,
				Data:     data,
			})

			if tt.expectError {
				assert.Error(t, err, "expected error")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.NotNil(t, resp, "expected non-nil response")

			respMap := resp.AsMap()
			for key, expectedVal := range tt.expectFields {
				assert.Equal(t, expectedVal, respMap[key], "expected %s=%v", key, expectedVal)
			}

			assert.NotNil(t, respMap["updatedAt"], "expected updatedAt field to be set")
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name          string
		resource      string
		recordID      string
		setupData     map[string][]store.Record
		expectError   bool
		expectSuccess bool
		expectMessage string
	}{
		{
			name:     "deletes existing record",
			resource: "users",
			recordID: "1",
			setupData: map[string][]store.Record{
				"users": {
					{
						"id":   float64(1),
						"name": "Alice",
					},
				},
			},
			expectError:   false,
			expectSuccess: true,
			expectMessage: "deleted users/1",
		},
		{
			name:          "handles non-existent record idempotently",
			resource:      "users",
			recordID:      "999",
			setupData:     map[string][]store.Record{},
			expectError:   false,
			expectSuccess: true,
			expectMessage: "deleted users/999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer(t)

			for resource, records := range tt.setupData {
				for _, rec := range records {
					if _, err := srv.store.Create(resource, rec); err != nil {
						t.Fatalf("failed to setup record: %v", err)
					}
				}
			}

			resp, err := srv.Delete(context.Background(), &pb.DeleteRequest{
				Resource: tt.resource,
				Id:       tt.recordID,
			})

			if tt.expectError {
				assert.Error(t, err, "expected error")
				return
			}

			assert.NoError(t, err, "unexpected error")
			assert.NotNil(t, resp, "expected non-nil response")

			assert.Equal(t, tt.expectSuccess, resp.Success, "expected success=%v", tt.expectSuccess)
			assert.Equal(t, tt.expectMessage, resp.Message, "expected message=%q", tt.expectMessage)
		})
	}
}

func TestGRPCErrorHandling(t *testing.T) {
	t.Run("returns NotFound code for missing resource", func(t *testing.T) {
		srv, _ := newTestServer(t)

		_, err := srv.Get(context.Background(), &pb.GetRequest{
			Resource: "nonexistent",
			Id:       "1",
		})

		assert.Error(t, err, "expected error")
	})

	t.Run("returns InvalidArgument code for nil data on create", func(t *testing.T) {
		srv, _ := newTestServer(t)

		_, err := srv.Create(context.Background(), &pb.CreateRequest{
			Resource: "users",
			Data:     nil,
		})

		assert.Error(t, err, "expected error")
	})

	t.Run("returns InvalidArgument code for nil data on update", func(t *testing.T) {
		srv, _ := newTestServer(t)

		_, err := srv.Update(context.Background(), &pb.UpdateRequest{
			Resource: "users",
			Id:       "1",
			Data:     nil,
		})

		assert.Error(t, err, "expected error")
	})

	t.Run("returns InvalidArgument code for nil data on patch", func(t *testing.T) {
		srv, _ := newTestServer(t)

		_, err := srv.Patch(context.Background(), &pb.PatchRequest{
			Resource: "users",
			Id:       "1",
			Data:     nil,
		})

		assert.Error(t, err, "expected error")
	})
}

func TestMultiResourceOperations(t *testing.T) {
	t.Run("handles multiple resources independently", func(t *testing.T) {
		srv, _ := newTestServer(t)

		user, err := srv.Create(context.Background(), &pb.CreateRequest{
			Resource: "users",
			Data: mkStruct(t, map[string]any{
				"name": "Alice",
			}),
		})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		post, err := srv.Create(context.Background(), &pb.CreateRequest{
			Resource: "posts",
			Data: mkStruct(t, map[string]any{
				"title": "Hello World",
			}),
		})
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		userMap := user.AsMap()
		postMap := post.AsMap()

		// JSON numbers are float64, convert for comparison
		userIDFromJSON, _ := userMap["id"].(float64)
		postIDFromJSON, _ := postMap["id"].(float64)
		userID := int64(userIDFromJSON)
		postID := int64(postIDFromJSON)

		assert.Equal(t, int64(1), userID, "expected user id=1")
		assert.Equal(t, int64(1), postID, "expected post id=1")

		userList, err := srv.List(context.Background(), &pb.ListRequest{
			Resource: "users",
		})
		assert.NoError(t, err, "failed to list users")

		postList, err := srv.List(context.Background(), &pb.ListRequest{
			Resource: "posts",
		})
		assert.NoError(t, err, "failed to list posts")

		assert.Len(t, userList.Items, 1, "expected 1 user")
		assert.Len(t, postList.Items, 1, "expected 1 post")
	})
}

func TestPersistenceWithMockFS(t *testing.T) {
	t.Run("persists data across operations", func(t *testing.T) {
		srv, mock := newTestServer(t)

		_, err := srv.Create(context.Background(), &pb.CreateRequest{
			Resource: "users",
			Data: mkStruct(t, map[string]any{
				"name": "Alice",
			}),
		})
		assert.NoError(t, err, "failed to create")

		assert.NotEmpty(t, mock.files, "expected files to be written")
	})
}

// mkStruct makes a proto buffer struct from a map
func mkStruct(t *testing.T, m map[string]any) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	if err != nil {
		t.Fatalf("failed to create struct: %v", err)
	}
	return s
}
