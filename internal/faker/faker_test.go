package faker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSchema(t *testing.T) {
	schemaJSON := `{
		"fields": {
			"email": {"type": "email"},
			"name": {"type": "name"},
			"age": {"type": "int"},
			"active": {"type": "bool"}
		}
	}`

	schema, err := LoadSchema(bytes.NewBufferString(schemaJSON))
	require.NoError(t, err)
	require.NotNil(t, schema)
	assert.Equal(t, 4, len(schema.Fields))
	assert.Equal(t, TypeEmail, schema.Fields["email"].Type)
	assert.Equal(t, TypeName, schema.Fields["name"].Type)
	assert.Equal(t, TypeInt, schema.Fields["age"].Type)
	assert.Equal(t, TypeBool, schema.Fields["active"].Type)
}

func TestLoadSchemaWithStaticValues(t *testing.T) {
	schemaJSON := `{
		"fields": {
			"status": {"type": "string", "value": "active"},
			"count": {"type": "int", "value": 42}
		}
	}`

	schema, err := LoadSchema(bytes.NewBufferString(schemaJSON))
	require.NoError(t, err)
	assert.Equal(t, "active", schema.Fields["status"].Value)
	assert.Equal(t, float64(42), schema.Fields["count"].Value)
}

func TestLoadSchemaInvalidJSON(t *testing.T) {
	invalidJSON := `{"fields": {`
	_, err := LoadSchema(bytes.NewBufferString(invalidJSON))
	assert.Error(t, err)
}

func TestGenerateRecord(t *testing.T) {
	schema := &Schema{
		Fields: map[string]Field{
			"email":     {Type: TypeEmail},
			"name":      {Type: TypeName},
			"age":       {Type: TypeInt},
			"active":    {Type: TypeBool},
			"join_date": {Type: TypeDate},
		},
	}

	record := GenerateRecord(schema)

	assert.Len(t, record, 5)
	assert.NotEmpty(t, record["email"])
	assert.NotEmpty(t, record["name"])
	assert.NotEmpty(t, record["age"])
	assert.NotNil(t, record["active"])
	assert.NotEmpty(t, record["join_date"])

	assert.IsType(t, "", record["email"])
	assert.IsType(t, "", record["name"])
	assert.IsType(t, int64(0), record["age"])
	assert.IsType(t, false, record["active"])
	assert.IsType(t, "", record["join_date"])
}

func TestGenerateRecordWithStaticValues(t *testing.T) {
	fixedPriority := float64(5)
	fixedStatus := "active"
	schema := &Schema{
		Fields: map[string]Field{
			"status":   {Type: TypeString, Value: fixedStatus},
			"priority": {Type: TypeFloat, Value: fixedPriority},
		},
	}

	record := GenerateRecord(schema)

	assert.Equal(t, fixedStatus, record["status"])
	assert.Equal(t, fixedPriority, record["priority"])
}

func TestGenerateFieldValue(t *testing.T) {
	tests := []struct {
		name         string
		fieldSchema  Field
		expectedType string
	}{
		{
			name:         "string",
			fieldSchema:  Field{Type: TypeString},
			expectedType: "string",
		},
		{
			name:         "int",
			fieldSchema:  Field{Type: TypeInt},
			expectedType: "int64",
		},
		{
			name:         "float",
			fieldSchema:  Field{Type: TypeFloat},
			expectedType: "float64",
		},
		{
			name:         "bool",
			fieldSchema:  Field{Type: TypeBool},
			expectedType: "bool",
		},
		{
			name:         "email",
			fieldSchema:  Field{Type: TypeEmail},
			expectedType: "string",
		},
		{
			name:         "name",
			fieldSchema:  Field{Type: TypeName},
			expectedType: "string",
		},
		{
			name:         "first_name",
			fieldSchema:  Field{Type: TypeFirstName},
			expectedType: "string",
		},
		{
			name:         "last_name",
			fieldSchema:  Field{Type: TypeLastName},
			expectedType: "string",
		},
		{
			name:         "phone",
			fieldSchema:  Field{Type: TypePhone},
			expectedType: "string",
		},
		{
			name:         "address",
			fieldSchema:  Field{Type: TypeAddress},
			expectedType: "string",
		},
		{
			name:         "city",
			fieldSchema:  Field{Type: TypeCity},
			expectedType: "string",
		},
		{
			name:         "state",
			fieldSchema:  Field{Type: TypeState},
			expectedType: "string",
		},
		{
			name:         "zipcode",
			fieldSchema:  Field{Type: TypeZipCode},
			expectedType: "string",
		},
		{
			name:         "country",
			fieldSchema:  Field{Type: TypeCountry},
			expectedType: "string",
		},
		{
			name:         "url",
			fieldSchema:  Field{Type: TypeURL},
			expectedType: "string",
		},
		{
			name:         "company",
			fieldSchema:  Field{Type: TypeCompany},
			expectedType: "string",
		},
		{
			name:         "job_title",
			fieldSchema:  Field{Type: TypeJobTitle},
			expectedType: "string",
		},
		{
			name:         "date",
			fieldSchema:  Field{Type: TypeDate},
			expectedType: "string",
		},
		{
			name:         "time",
			fieldSchema:  Field{Type: TypeTime},
			expectedType: "string",
		},
		{
			name:         "datetime",
			fieldSchema:  Field{Type: TypeDatetime},
			expectedType: "string",
		},
		{
			name:         "uuid",
			fieldSchema:  Field{Type: TypeUUID},
			expectedType: "string",
		},
		{
			name:         "slug",
			fieldSchema:  Field{Type: TypeSlug},
			expectedType: "string",
		},
		{
			name:         "word",
			fieldSchema:  Field{Type: TypeWord},
			expectedType: "string",
		},
		{
			name:         "sentence",
			fieldSchema:  Field{Type: TypeSentence},
			expectedType: "string",
		},
		{
			name:         "paragraph",
			fieldSchema:  Field{Type: TypeParagraph},
			expectedType: "string",
		},
		{
			name:         "color",
			fieldSchema:  Field{Type: TypeColor},
			expectedType: "string",
		},
		{
			name:         "latitude",
			fieldSchema:  Field{Type: TypeLatitude},
			expectedType: "float64",
		},
		{
			name:         "longitude",
			fieldSchema:  Field{Type: TypeLongitude},
			expectedType: "float64",
		},
		{
			name:         "ipv4",
			fieldSchema:  Field{Type: TypeIPv4},
			expectedType: "string",
		},
		{
			name:         "ipv6",
			fieldSchema:  Field{Type: TypeIPv6},
			expectedType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := generateFieldValue(tt.fieldSchema)
			assert.NotNil(t, value)
			assert.Equal(t, tt.expectedType, fmt.Sprintf("%T", value))
		})
	}
}

func TestGenerateRecordConsistency(t *testing.T) {
	schema := &Schema{
		Fields: map[string]Field{
			"email": {Type: TypeEmail},
			"name":  {Type: TypeName},
		},
	}

	records := make([]map[string]any, 10)
	for i := range 10 {
		records[i] = GenerateRecord(schema)
	}

	emails := make(map[string]bool)
	for _, r := range records {
		email := r["email"].(string)
		assert.False(t, emails[email], "emails should be unique")
		emails[email] = true
	}
}

func TestGenerateRecordMarshalJSON(t *testing.T) {
	schema := &Schema{
		Fields: map[string]Field{
			"email": {Type: TypeEmail},
			"count": {Type: TypeInt},
		},
	}

	record := GenerateRecord(schema)
	data, err := json.Marshal(record)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	var unmarshaled map[string]any
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.NotNil(t, unmarshaled["email"])
	assert.NotNil(t, unmarshaled["count"])
}

func TestLoadJSONSchema_Basic(t *testing.T) {
	schemaJSON := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"email": {
				"type": "string",
				"format": "email"
			},
			"age": {
				"type": "integer"
			},
			"score": {
				"type": "number"
			},
			"active": {
				"type": "boolean"
			}
		}
	}`

	schema, err := LoadSchema(bytes.NewBufferString(schemaJSON))
	require.NoError(t, err)
	require.NotNil(t, schema)
	assert.Equal(t, 4, len(schema.Fields))
	assert.Equal(t, TypeEmail, schema.Fields["email"].Type)
	assert.Equal(t, TypeInt, schema.Fields["age"].Type)
	assert.Equal(t, TypeFloat, schema.Fields["score"].Type)
	assert.Equal(t, TypeBool, schema.Fields["active"].Type)
}

func TestLoadJSONSchema_Formats(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "string", "format": "uuid"},
			"created_at": {"type": "string", "format": "date-time"},
			"birth_date": {"type": "string", "format": "date"},
			"website": {"type": "string", "format": "uri"},
			"ip": {"type": "string", "format": "ipv4"},
			"ipv6": {"type": "string", "format": "ipv6"}
		}
	}`

	schema, err := LoadSchema(bytes.NewBufferString(schemaJSON))
	require.NoError(t, err)
	assert.Equal(t, TypeUUID, schema.Fields["id"].Type)
	assert.Equal(t, TypeDatetime, schema.Fields["created_at"].Type)
	assert.Equal(t, TypeDate, schema.Fields["birth_date"].Type)
	assert.Equal(t, TypeURL, schema.Fields["website"].Type)
	assert.Equal(t, TypeIPv4, schema.Fields["ip"].Type)
	assert.Equal(t, TypeIPv6, schema.Fields["ipv6"].Type)
}

func TestLoadJSONSchema_CustomFormats(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"full_name": {"type": "string", "format": "name"},
			"first_name": {"type": "string", "format": "first_name"},
			"last_name": {"type": "string", "format": "last_name"},
			"phone": {"type": "string", "format": "phone"},
			"address": {"type": "string", "format": "address"},
			"city": {"type": "string", "format": "city"},
			"company": {"type": "string", "format": "company"},
			"title": {"type": "string", "format": "job_title"}
		}
	}`

	schema, err := LoadSchema(bytes.NewBufferString(schemaJSON))
	require.NoError(t, err)
	assert.Equal(t, TypeName, schema.Fields["full_name"].Type)
	assert.Equal(t, TypeFirstName, schema.Fields["first_name"].Type)
	assert.Equal(t, TypeLastName, schema.Fields["last_name"].Type)
	assert.Equal(t, TypePhone, schema.Fields["phone"].Type)
	assert.Equal(t, TypeAddress, schema.Fields["address"].Type)
	assert.Equal(t, TypeCity, schema.Fields["city"].Type)
	assert.Equal(t, TypeCompany, schema.Fields["company"].Type)
	assert.Equal(t, TypeJobTitle, schema.Fields["title"].Type)
}

func TestLoadJSONSchema_ConstValues(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"version": {
				"type": "string",
				"const": "1.0.0"
			},
			"environment": {
				"type": "string",
				"const": "production"
			},
			"max_retries": {
				"type": "integer",
				"const": 3
			}
		}
	}`

	schema, err := LoadSchema(bytes.NewBufferString(schemaJSON))
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", schema.Fields["version"].Value)
	assert.Equal(t, "production", schema.Fields["environment"].Value)
	assert.Equal(t, float64(3), schema.Fields["max_retries"].Value)
}

func TestLoadJSONSchema_DefaultValues(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"status": {
				"type": "string",
				"default": "pending"
			},
			"priority": {
				"type": "integer",
				"default": 5
			}
		}
	}`

	schema, err := LoadSchema(bytes.NewBufferString(schemaJSON))
	require.NoError(t, err)
	assert.Equal(t, "pending", schema.Fields["status"].Value)
	assert.Equal(t, float64(5), schema.Fields["priority"].Value)
}

func TestLoadJSONSchema_WithoutSchema(t *testing.T) {
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"}
		}
	}`

	schema, err := LoadSchema(bytes.NewBufferString(schemaJSON))
	require.NoError(t, err)
	assert.Equal(t, 1, len(schema.Fields))
	assert.Equal(t, TypeString, schema.Fields["name"].Type)
}

func TestLoadJSONSchema_GenerateRecords(t *testing.T) {
	schemaJSON := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"email": {"type": "string", "format": "email"},
			"name": {"type": "string", "format": "name"},
			"age": {"type": "integer"},
			"active": {"type": "boolean"},
			"status": {"type": "string", "const": "verified"}
		}
	}`

	schema, err := LoadSchema(bytes.NewBufferString(schemaJSON))
	require.NoError(t, err)

	record := GenerateRecord(schema)
	assert.Len(t, record, 5)
	assert.NotEmpty(t, record["email"])
	assert.NotEmpty(t, record["name"])
	assert.NotEmpty(t, record["age"])
	assert.NotNil(t, record["active"])
	assert.Equal(t, "verified", record["status"])

	assert.IsType(t, "", record["email"])
	assert.IsType(t, "", record["name"])
	assert.IsType(t, int64(0), record["age"])
	assert.IsType(t, false, record["active"])
	assert.IsType(t, "", record["status"])
}

func TestLoadJSONSchema_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		schema string
	}{
		{
			name:   "empty properties",
			schema: `{"type": "object", "properties": {}}`,
		},
		{
			name:   "invalid schema",
			schema: `{"$schema": "http://json-schema.org/draft-07/schema#", "type": "object", "properties": {}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadSchema(bytes.NewBufferString(tt.schema))
			assert.Error(t, err)
		})
	}
}

func TestIsJSONSchema(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected bool
	}{
		{
			name:     "has $schema",
			input:    map[string]any{"$schema": "http://json-schema.org/draft-07/schema#"},
			expected: true,
		},
		{
			name:     "has type object and properties",
			input:    map[string]any{"type": "object", "properties": map[string]any{}},
			expected: true,
		},
		{
			name:     "simplified schema with fields",
			input:    map[string]any{"fields": map[string]any{}},
			expected: false,
		},
		{
			name:     "has type but no properties",
			input:    map[string]any{"type": "object"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isJSONSchema(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBackwardCompatibility(t *testing.T) {
	simplifiedFormat := `{
		"fields": {
			"email": {"type": "email"},
			"name": {"type": "name"},
			"age": {"type": "int"}
		}
	}`

	schema, err := LoadSchema(bytes.NewBufferString(simplifiedFormat))
	require.NoError(t, err)
	assert.Equal(t, 3, len(schema.Fields))
	assert.Equal(t, TypeEmail, schema.Fields["email"].Type)
	assert.Equal(t, TypeName, schema.Fields["name"].Type)
	assert.Equal(t, TypeInt, schema.Fields["age"].Type)

	record := GenerateRecord(schema)
	assert.Len(t, record, 3)
	assert.NotEmpty(t, record["email"])
	assert.NotEmpty(t, record["name"])
	assert.NotEmpty(t, record["age"])
}
