package faker

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jimschubert/rumor/internal/store"
	"github.com/santhosh-tekuri/jsonschema/v6"
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
// Supports both simplified schema format and standard JSON Schema
func LoadSchema(r io.Reader) (*Schema, error) {
	var raw map[string]any
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode schema: %w", err)
	}

	if isJSONSchema(raw) {
		return parseJSONSchema(raw)
	}
	return parseSimplifiedSchema(raw)
}

// isJSONSchema detects if the input is a JSON Schema
func isJSONSchema(raw map[string]any) bool {
	if _, hasSchema := raw["$schema"]; hasSchema {
		return true
	}
	if typeVal, hasType := raw["type"]; hasType && typeVal == "object" {
		if _, hasProps := raw["properties"]; hasProps {
			return true
		}
	}
	return false
}

// parseSimplifiedSchema parses the simplified schema format
func parseSimplifiedSchema(raw map[string]any) (*Schema, error) {
	var schema Schema
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to re-marshal schema: %w", err)
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse simplified schema: %w", err)
	}
	return &schema, nil
}

// parseJSONSchema parses standard JSON Schema format
func parseJSONSchema(raw map[string]any) (*Schema, error) {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", raw); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	// just validate it's a real schema
	_, err := compiler.Compile("schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile JSON schema: %w", err)
	}

	properties, ok := raw["properties"].(map[string]any)
	if !ok || len(properties) == 0 {
		return nil, fmt.Errorf("JSON schema must have 'properties' field with at least one property")
	}

	schema := &Schema{
		Fields: make(map[string]Field),
	}

	for fieldName, propRaw := range properties {
		prop, ok := propRaw.(map[string]any)
		if !ok {
			continue
		}

		field := jsonSchemaPropertyToField(prop)
		schema.Fields[fieldName] = field
	}

	if len(schema.Fields) == 0 {
		return nil, fmt.Errorf("schema must define at least one valid field in properties")
	}

	return schema, nil
}

// jsonSchemaPropertyToField converts a JSON Schema property to a Field
func jsonSchemaPropertyToField(prop map[string]any) Field {
	field := Field{}

	if constVal, hasConst := prop["const"]; hasConst {
		field.Value = constVal
		field.Type = inferTypeFromValue(constVal)
		return field
	}

	if defaultVal, hasDefault := prop["default"]; hasDefault {
		field.Value = defaultVal
	}

	typeStr, _ := prop["type"].(string)
	formatStr, _ := prop["format"].(string)

	field.Type = mapJSONSchemaType(typeStr, formatStr)
	return field
}

// inferTypeFromValue infers a SchemaType from a constant value
func inferTypeFromValue(val any) SchemaType {
	switch val.(type) {
	case bool:
		return TypeBool
	case float64:
		return TypeFloat
	case int, int64:
		return TypeInt
	case string:
		return TypeString
	default:
		return TypeString
	}
}

// mapJSONSchemaType maps JSON Schema type and format to SchemaType
func mapJSONSchemaType(typeStr, formatStr string) SchemaType {
	// format > type
	if formatStr != "" {
		switch strings.ToLower(formatStr) {
		case "email", "idn-email":
			return TypeEmail
		case "date-time", "datetime":
			return TypeDatetime
		case "date":
			return TypeDate
		case "time":
			return TypeTime
		case "uuid":
			return TypeUUID
		case "uri", "uri-reference", "url":
			return TypeURL
		case "ipv4":
			return TypeIPv4
		case "ipv6":
			return TypeIPv6

		// these are faker formats which _could_ be used in JSON schema format (but probably aren't?)
		case "name":
			return TypeName
		case "first_name", "firstname":
			return TypeFirstName
		case "last_name", "lastname":
			return TypeLastName
		case "phone", "phone_number":
			return TypePhone
		case "address":
			return TypeAddress
		case "city":
			return TypeCity
		case "state":
			return TypeState
		case "zipcode", "zip", "postal-code":
			return TypeZipCode
		case "country":
			return TypeCountry
		case "company":
			return TypeCompany
		case "job_title", "jobtitle":
			return TypeJobTitle
		case "slug":
			return TypeSlug
		case "word":
			return TypeWord
		case "sentence":
			return TypeSentence
		case "paragraph":
			return TypeParagraph
		case "color":
			return TypeColor
		case "latitude":
			return TypeLatitude
		case "longitude":
			return TypeLongitude
		}
	}

	switch typeStr {
	case "string":
		return TypeString
	case "integer":
		return TypeInt
	case "number":
		return TypeFloat
	case "boolean":
		return TypeBool
	default:
		return TypeString
	}
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
