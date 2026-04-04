package utils

import (
	"testing"
)

func TestRandom(t *testing.T) {
	result := Random()

	if result == "" {
		t.Error("Random() returned empty string")
	}

	if len(result) < 10 {
		t.Errorf("Random() length = %d, expected at least 10", len(result))
	}
}

func TestRandomRange(t *testing.T) {
	tests := []struct {
		name string
		min  int
		max  int
	}{
		{
			name: "range 1-10",
			min:  1,
			max:  10,
		},
		{
			name: "range 100-200",
			min:  100,
			max:  200,
		},
		{
			name: "same min max",
			min:  5,
			max:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				result := RandomRange(tt.min, tt.max)
				if result < tt.min || result > tt.max {
					t.Errorf("RandomRange() = %d, want between %d and %d", result, tt.min, tt.max)
				}
			}
		})
	}
}

func TestRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "length 10",
			length: 10,
		},
		{
			name:   "length 32",
			length: 32,
		},
		{
			name:   "length 0",
			length: 0,
		},
		{
			name:   "length 100",
			length: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandomString(tt.length)

			if len(result) != tt.length {
				t.Errorf("RandomString() length = %d, want %d", len(result), tt.length)
			}

			for _, c := range result {
				isValid := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
				if !isValid {
					t.Errorf("RandomString() contains invalid character: %c", c)
					break
				}
			}
		})
	}
}

func TestRandomStringUniqueness(t *testing.T) {
	results := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		s := RandomString(16)
		if results[s] {
			t.Logf("Warning: RandomString() generated duplicate (acceptable but rare)")
		}
		results[s] = true
	}

	if len(results) < 990 {
		t.Errorf("RandomString() produced too many duplicates: %d unique out of 1000", len(results))
	}
}

func TestRandomUUID(t *testing.T) {
	result := RandomUUID()

	if len(result) != 32 {
		t.Errorf("RandomUUID() length = %d, want 32", len(result))
	}

	for _, c := range result {
		isValid := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		if !isValid {
			t.Errorf("RandomUUID() contains invalid character: %c", c)
			break
		}
	}
}

func TestGenerateUUID(t *testing.T) {
	result := GenerateUUID()

	if len(result) != 32 {
		t.Errorf("GenerateUUID() length = %d, want 32", len(result))
	}
}

func TestRandomWithPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{
			name:    "x pattern",
			pattern: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		},
		{
			name:    "mixed pattern",
			pattern: "xxxx-xyxy-xyxy-xxxx",
		},
		{
			name:    "simple pattern",
			pattern: "xxxx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RandomWithPattern(tt.pattern)

			if len(result) != len(tt.pattern) {
				t.Errorf("RandomWithPattern() length = %d, want %d", len(result), len(tt.pattern))
			}

			for i, c := range result {
				origChar := rune(tt.pattern[i])
				if origChar == 'x' || origChar == 'y' {
					isHex := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
					if !isHex {
						t.Errorf("RandomWithPattern() hex position %d has invalid char: %c", i, c)
					}
				} else {
					if c != origChar {
						t.Errorf("RandomWithPattern() non-pattern char changed at position %d", i)
					}
				}
			}
		})
	}
}

func TestRandomWithPatternYConstraint(t *testing.T) {
	pattern := "yyyy"
	for i := 0; i < 100; i++ {
		result := RandomWithPattern(pattern)
		if len(result) != 4 {
			t.Errorf("Result length = %d, want 4", len(result))
		}
		for _, c := range result {
			if c != '8' && c != '9' && c != 'a' && c != 'b' {
				t.Errorf("y pattern produced invalid hex digit: %c (should be 8-b)", c)
			}
		}
	}
}
