package output

import (
	"testing"
)

func TestPrintTable(t *testing.T) {
	out := NewOutput(true, map[string]interface{}{
		"message": "test message",
		"count":   42,
	})

	err := PrintTable(out)
	if err != nil {
		t.Fatalf("PrintTable() error = %v", err)
	}
}

func TestPrintTableWithNestedData(t *testing.T) {
	out := NewOutput(true, map[string]interface{}{
		"user": map[string]interface{}{
			"name":  "test",
			"email": "test@example.com",
		},
		"items": []interface{}{
			map[string]interface{}{"id": 1, "name": "item1"},
			map[string]interface{}{"id": 2, "name": "item2"},
		},
	})

	err := PrintTable(out)
	if err != nil {
		t.Fatalf("PrintTable() error = %v", err)
	}
}

func TestPrintTableError(t *testing.T) {
	out := NewErrorOutput("TABLE_ERR", "Table error message")

	err := PrintTable(out)
	if err != nil {
		t.Fatalf("PrintTable() error = %v", err)
	}
}

func TestPrintTableEmptyData(t *testing.T) {
	out := NewOutput(true, nil)

	err := PrintTable(out)
	if err != nil {
		t.Fatalf("PrintTable() error = %v", err)
	}
}

func TestPrintTableComplexStructure(t *testing.T) {
	out := NewOutput(true, map[string]interface{}{
		"files": []interface{}{
			map[string]interface{}{
				"name":   "file1.txt",
				"size":   int64(1024),
				"is_dir": false,
			},
			map[string]interface{}{
				"name":   "folder1",
				"size":   int64(0),
				"is_dir": true,
			},
		},
		"count": 2,
		"path":  "/test/path",
	})

	err := PrintTable(out)
	if err != nil {
		t.Fatalf("PrintTable() error = %v", err)
	}
}

func TestPrintTableCapacity(t *testing.T) {
	out := NewOutput(true, map[string]interface{}{
		"account": "test@example.com",
		"personal": map[string]interface{}{
			"total":    int64(1073741824),
			"used":     int64(536870912),
			"free":     int64(536870912),
			"total_gb": float64(1.0),
			"used_gb":  float64(0.5),
		},
	})

	err := PrintTable(out)
	if err != nil {
		t.Fatalf("PrintTable() error = %v", err)
	}
}

func TestGetInt64(t *testing.T) {
	m := map[string]interface{}{
		"int":   100,
		"int64": int64(200),
		"float": float64(300.5),
	}

	if getInt64(m, "int") != 100 {
		t.Errorf("getInt64(int) = %d, want 100", getInt64(m, "int"))
	}

	if getInt64(m, "int64") != 200 {
		t.Errorf("getInt64(int64) = %d, want 200", getInt64(m, "int64"))
	}

	if getInt64(m, "float") != 300 {
		t.Errorf("getInt64(float) = %d, want 300", getInt64(m, "float"))
	}

	if getInt64(m, "nonexistent") != 0 {
		t.Error("getInt64(nonexistent) should return 0")
	}
}

func TestGetFloat64(t *testing.T) {
	m := map[string]interface{}{
		"float": float64(1.5),
		"int":   2,
	}

	if getFloat64(m, "float") != 1.5 {
		t.Errorf("getFloat64(float) = %f, want 1.5", getFloat64(m, "float"))
	}

	if getFloat64(m, "int") != 2.0 {
		t.Errorf("getFloat64(int) = %f, want 2.0", getFloat64(m, "int"))
	}
}

func TestGetString(t *testing.T) {
	m := map[string]interface{}{
		"str": "test",
	}

	if getString(m, "str") != "test" {
		t.Errorf("getString(str) = %s, want test", getString(m, "str"))
	}

	if getString(m, "nonexistent") != "" {
		t.Error("getString(nonexistent) should return empty string")
	}
}

func TestGetBool(t *testing.T) {
	m := map[string]interface{}{
		"bool": true,
	}

	if !getBool(m, "bool") {
		t.Error("getBool(bool) should return true")
	}

	if getBool(m, "nonexistent") {
		t.Error("getBool(nonexistent) should return false")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.00 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
	}

	for _, tt := range tests {
		result := formatSize(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatSize(%d) = %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}
