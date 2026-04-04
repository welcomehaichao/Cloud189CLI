package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNewOutput(t *testing.T) {
	data := map[string]string{"key": "value"}
	out := NewOutput(true, data)

	if !out.Success {
		t.Error("Success should be true")
	}

	if out.Data == nil {
		t.Error("Data should not be nil")
	}
}

func TestNewErrorOutput(t *testing.T) {
	out := NewErrorOutput("ERR001", "Test error message")

	if out.Success {
		t.Error("Success should be false for error output")
	}

	if out.Error == nil {
		t.Fatal("Error should not be nil")
	}

	if out.Error.Code != "ERR001" {
		t.Errorf("Error.Code = %s, want ERR001", out.Error.Code)
	}

	if out.Error.Message != "Test error message" {
		t.Errorf("Error.Message = %s", out.Error.Message)
	}
}

func TestNewErrorOutputWithDetails(t *testing.T) {
	details := map[string]interface{}{
		"field": "username",
	}
	out := NewErrorOutputWithDetails("ERR002", "Validation error", details)

	if out.Error.Details == nil {
		t.Fatal("Error.Details should not be nil")
	}

	if out.Error.Details["field"] != "username" {
		t.Errorf("Details[field] = %v, want username", out.Error.Details["field"])
	}
}

func TestPrintJSON(t *testing.T) {
	out := NewOutput(true, map[string]string{"test": "data"})

	var buf bytes.Buffer
	err := PrintJSONToWriter(&buf, out)
	if err != nil {
		t.Fatalf("PrintJSONToWriter() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if result["success"] != true {
		t.Error("JSON should contain success: true")
	}
}

func TestPrintJSONError(t *testing.T) {
	out := NewErrorOutput("CODE", "message")

	var buf bytes.Buffer
	err := PrintJSONToWriter(&buf, out)
	if err != nil {
		t.Fatalf("PrintJSONToWriter() error = %v", err)
	}

	var result Output
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	if result.Success {
		t.Error("Success should be false for error")
	}
}

func TestPrintYAML(t *testing.T) {
	out := NewOutput(true, map[string]interface{}{
		"message": "test",
		"count":   42,
	})

	var buf bytes.Buffer
	err := PrintYAMLToWriter(&buf, out)
	if err != nil {
		t.Fatalf("PrintYAMLToWriter() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("YAML output should not be empty")
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "bool true",
			input:    true,
			expected: "true",
		},
		{
			name:     "bool false",
			input:    false,
			expected: "false",
		},
		{
			name:     "int",
			input:    42,
			expected: "42",
		},
		{
			name:     "int64",
			input:    int64(123456789),
			expected: "123456789",
		},
		{
			name:     "float64",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "nil",
			input:    nil,
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			if result != tt.expected {
				t.Errorf("formatValue() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestGetIndent(t *testing.T) {
	tests := []struct {
		level    int
		expected string
	}{
		{0, ""},
		{2, "  "},
		{4, "    "},
		{6, "      "},
	}

	for _, tt := range tests {
		result := getIndent(tt.level)
		if result != tt.expected {
			t.Errorf("getIndent(%d) = %q, want %q", tt.level, result, tt.expected)
		}
	}
}

func TestOutputFormatConstants(t *testing.T) {
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %s, want json", FormatJSON)
	}

	if FormatYAML != "yaml" {
		t.Errorf("FormatYAML = %s, want yaml", FormatYAML)
	}

	if FormatTable != "table" {
		t.Errorf("FormatTable = %s, want table", FormatTable)
	}
}

func TestToYAMLMap(t *testing.T) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"nested": map[string]interface{}{
			"inner": "value",
		},
	}

	result, err := toYAML(data, 0)
	if err != nil {
		t.Fatalf("toYAML() error = %v", err)
	}

	if result == "" {
		t.Error("toYAML should return non-empty string")
	}
}

func TestToYAMLArray(t *testing.T) {
	data := []interface{}{
		"item1",
		"item2",
		map[string]interface{}{"key": "value"},
	}

	result, err := toYAML(data, 0)
	if err != nil {
		t.Fatalf("toYAML() error = %v", err)
	}

	if result == "" {
		t.Error("toYAML should return non-empty string for arrays")
	}
}
