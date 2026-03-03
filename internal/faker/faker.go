package faker

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jimschubert/rumor/internal/store"
)

// SchemaType represents a field type for data generation
type SchemaType string

// TODO: add support for more faker types
const (
	TypeString    SchemaType = "string"
	TypeInt       SchemaType = "int"
	TypeFloat     SchemaType = "float"
	TypeBool      SchemaType = "bool"
	TypeEmail     SchemaType = "email"
	TypeName      SchemaType = "name"
	TypeFirstName SchemaType = "first_name"
	TypeLastName  SchemaType = "last_name"
	TypePhone     SchemaType = "phone"
	TypeAddress   SchemaType = "address"
	TypeCity      SchemaType = "city"
	TypeState     SchemaType = "state"
	TypeZipCode   SchemaType = "zipcode"
	TypeCountry   SchemaType = "country"
	TypeURL       SchemaType = "url"
	TypeCompany   SchemaType = "company"
	TypeJobTitle  SchemaType = "job_title"
	TypeDate      SchemaType = "date"
	TypeTime      SchemaType = "time"
	TypeDatetime  SchemaType = "datetime"
	TypeUUID      SchemaType = "uuid"
	TypeSlug      SchemaType = "slug"
	TypeWord      SchemaType = "word"
	TypeSentence  SchemaType = "sentence"
	TypeParagraph SchemaType = "paragraph"
	TypeColor     SchemaType = "color"
	TypeLatitude  SchemaType = "latitude"
	TypeLongitude SchemaType = "longitude"
	TypeIPv4      SchemaType = "ipv4"
	TypeIPv6      SchemaType = "ipv6"
)

// Field defines how to generate a single field
type Field struct {
	Type  SchemaType `json:"type"`
	Value any        `json:"value,omitempty"`
}

// Schema defines the structure for generating fake data
type Schema struct {
	Fields map[string]Field `json:"fields"`
}

// LoadSchema reads a schema from a JSON reader
func LoadSchema(r io.Reader) (*Schema, error) {
	// TODO: support actual JSON Schema?
	var schema Schema
	if err := json.NewDecoder(r).Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode schema: %w", err)
	}
	return &schema, nil
}

// GenerateRecord generates a single record based on the schema
func GenerateRecord(schema *Schema) store.Record {
	record := make(store.Record)
	for fieldName, fieldSchema := range schema.Fields {
		record[fieldName] = generateFieldValue(fieldSchema)
	}
	return record
}

// generateFieldValue generates a value for a single field based on its schema
func generateFieldValue(fs Field) any {
	switch fs.Type {
	case TypeString:
		if fs.Value != nil {
			return fs.Value
		}
		return gofakeit.Word()

	case TypeInt:
		if fs.Value != nil {
			if v, ok := fs.Value.(float64); ok {
				return int64(v)
			}
			if v, ok := fs.Value.(int); ok {
				return int64(v)
			}
		}
		return rand.Int63n(1000)

	case TypeFloat:
		if fs.Value != nil {
			if v, ok := fs.Value.(float64); ok {
				return v
			}
		}
		return rand.Float64() * 1000

	case TypeBool:
		return rand.Intn(2) == 0

	case TypeEmail:
		return gofakeit.Email()

	case TypeName:
		return gofakeit.Name()

	case TypeFirstName:
		return gofakeit.FirstName()

	case TypeLastName:
		return gofakeit.LastName()

	case TypePhone:
		return gofakeit.Phone()

	case TypeAddress:
		return gofakeit.Address().Address

	case TypeCity:
		return gofakeit.City()

	case TypeState:
		return gofakeit.State()

	case TypeZipCode:
		return gofakeit.Zip()

	case TypeCountry:
		return gofakeit.Country()

	case TypeURL:
		return gofakeit.URL()

	case TypeCompany:
		return gofakeit.Company()

	case TypeJobTitle:
		return gofakeit.JobTitle()

	case TypeDate:
		return gofakeit.Date().Format("2006-01-02")

	case TypeTime:
		return time.Now().Format("15:04:05")

	case TypeDatetime:
		return gofakeit.Date().Format("2006-01-02T15:04:05Z")

	case TypeUUID:
		return gofakeit.UUID()

	case TypeSlug:
		return gofakeit.UrlSlug(rand.Intn(3) + 2)

	case TypeWord:
		return gofakeit.Word()

	case TypeSentence:
		return gofakeit.Sentence()

	case TypeParagraph:
		return gofakeit.Paragraph()

	case TypeColor:
		return gofakeit.Color()

	case TypeLatitude:
		return gofakeit.Latitude()

	case TypeLongitude:
		return gofakeit.Longitude()

	case TypeIPv4:
		return gofakeit.IPv4Address()

	case TypeIPv6:
		return gofakeit.IPv6Address()

	default:
		return gofakeit.Word()
	}
}
