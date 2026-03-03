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
