package store

// Record defines a "dynamic" set of data
type Record = map[string]any

// Store defines the interface for record storage operations.
type Store interface {
	List(resource string, filters map[string]string, page, pageSize int) ([]Record, int, error)
	Get(resource, id string) (Record, error)
	Create(resource string, data Record) (Record, error)
	Update(resource, id string, data Record) (Record, error)
	Patch(resource, id string, patch Record) (Record, error)
	Delete(resource, id string) error
}
