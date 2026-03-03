// Package faker generates realistic fake data based on JSON schema definitions.
//
// # Overview
//
// The faker package provides a flexible data generation library built on top of gofakeit.
// It allows users to define data structures using schema files and generate realistic records
// with fields like emails, names, addresses, and more.
//
// # Schema Formats
//
// Two schema formats are supported with automatic detection:
//
// Simplified format:
//
//	{
//	  "fields": {
//	    "email": {"type": "email"},
//	    "name": {"type": "name"},
//	    "status": {"type": "string", "value": "active"}
//	  }
//	}
//
// Standard JSON Schema format:
//
//	{
//	  "$schema": "http://json-schema.org/draft-07/schema#",
//	  "type": "object",
//	  "properties": {
//	    "email": {"type": "string", "format": "email"},
//	    "name": {"type": "string", "format": "name"},
//	    "status": {"type": "string", "const": "active"}
//	  }
//	}
//
// # Supported Field Types
//
// The following field types are supported:
//
// Basic Types:
//   - string - Random word (or static value if provided)
//   - int - Random integer (0-1000)
//   - float - Random float (0-1000)
//   - bool - Random boolean
//
// Person Information:
//   - name - Full name
//   - first_name - First name
//   - last_name - Last name
//   - email - Email address
//   - phone - Phone number
//
// Location Information:
//   - address - Street address
//   - city - City name
//   - state - State/province
//   - zipcode - ZIP/postal code
//   - country - Country name
//   - latitude - Latitude coordinate
//   - longitude - Longitude coordinate
//
// Professional Information:
//   - company - Company name
//   - job_title - Job title
//
// Digital Information:
//   - url - URL
//   - ipv4 - IPv4 address
//   - ipv6 - IPv6 address
//   - uuid - UUID
//
// Text Content:
//   - word - Single word
//   - sentence - Single sentence of multiple words
//   - paragraph - Multi-sentence paragraph
//   - slug - URL slug
//
// Date/Time:
//   - date - Date (YYYY-MM-DD format)
//   - time - Time (HH:MM:SS format)
//   - datetime - DateTime (RFC3339 format)
//
// Other:
//   - color - Color name
//
// # Usage Example
//
// Load a schema from a file:
//
//	schemaFile, _ := os.Open("users.schema.json")
//	schema, _ := faker.LoadSchema(schemaFile)
//
// Generate records:
//
//	for i := 0; i < 100; i++ {
//	  record := faker.GenerateRecord(schema)
//	  // Use record...
//	}
//
// # Static Values
//
// Fields can have static values that don't vary across records:
//
//	"fields": {
//	  "status": {"type": "string", "value": "active"},
//	  "priority": {"type": "int", "value": 5}
//	}
package faker
