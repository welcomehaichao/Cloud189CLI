package updater

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected *Version
		hasError bool
	}{
		{"v1.5.0", &Version{Major: 1, Minor: 5, Patch: 0}, false},
		{"1.5.0", &Version{Major: 1, Minor: 5, Patch: 0}, false},
		{"v1.2.3", &Version{Major: 1, Minor: 2, Patch: 3}, false},
		{"v2.0.0", &Version{Major: 2, Minor: 0, Patch: 0}, false},
		{"invalid", nil, true},
		{"v1.2", nil, true},
		{"v1.2.3.4", nil, true},
	}

	for _, tt := range tests {
		ver, err := ParseVersion(tt.input)

		if tt.hasError {
			if err == nil {
				t.Errorf("ParseVersion(%s) 应该返回错误，但没有", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ParseVersion(%s) 返回错误: %v", tt.input, err)
			}
			if ver.Major != tt.expected.Major || ver.Minor != tt.expected.Minor || ver.Patch != tt.expected.Patch {
				t.Errorf("ParseVersion(%s) = %v, want %v", tt.input, ver, tt.expected)
			}
		}
	}
}

func TestVersionString(t *testing.T) {
	ver := &Version{Major: 1, Minor: 5, Patch: 0}
	expected := "v1.5.0"

	if ver.String() != expected {
		t.Errorf("Version.String() = %s, want %s", ver.String(), expected)
	}
}

func TestVersionLessThan(t *testing.T) {
	tests := []struct {
		v1       *Version
		v2       *Version
		expected bool
	}{
		{&Version{1, 0, 0}, &Version{1, 0, 1}, true},
		{&Version{1, 0, 1}, &Version{1, 1, 0}, true},
		{&Version{1, 5, 0}, &Version{2, 0, 0}, true},
		{&Version{1, 5, 0}, &Version{1, 5, 0}, false},
		{&Version{2, 0, 0}, &Version{1, 5, 0}, false},
		{&Version{1, 5, 1}, &Version{1, 5, 0}, false},
	}

	for _, tt := range tests {
		result := tt.v1.LessThan(tt.v2)
		if result != tt.expected {
			t.Errorf("%s.LessThan(%s) = %v, want %v", tt.v1.String(), tt.v2.String(), result, tt.expected)
		}
	}
}

func TestVersionEqual(t *testing.T) {
	tests := []struct {
		v1       *Version
		v2       *Version
		expected bool
	}{
		{&Version{1, 5, 0}, &Version{1, 5, 0}, true},
		{&Version{1, 5, 0}, &Version{1, 5, 1}, false},
		{&Version{1, 5, 0}, &Version{2, 0, 0}, false},
	}

	for _, tt := range tests {
		result := tt.v1.Equal(tt.v2)
		if result != tt.expected {
			t.Errorf("%s.Equal(%s) = %v, want %v", tt.v1.String(), tt.v2.String(), result, tt.expected)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		current  string
		latest   string
		expected bool
		hasError bool
	}{
		{"v1.5.0", "v1.5.0", true, false},
		{"v1.4.0", "v1.5.0", false, false},
		{"v1.5.0", "v1.4.0", true, false},
		{"v1.5.0", "v2.0.0", false, false},
		{"invalid", "v1.5.0", false, true},
		{"v1.5.0", "invalid", false, true},
	}

	for _, tt := range tests {
		isOlderOrEqual, err := CompareVersions(tt.current, tt.latest)

		if tt.hasError {
			if err == nil {
				t.Errorf("CompareVersions(%s, %s) 应该返回错误", tt.current, tt.latest)
			}
		} else {
			if err != nil {
				t.Errorf("CompareVersions(%s, %s) 返回错误: %v", tt.current, tt.latest, err)
			}
			if isOlderOrEqual != tt.expected {
				t.Errorf("CompareVersions(%s, %s) = %v, want %v", tt.current, tt.latest, isOlderOrEqual, tt.expected)
			}
		}
	}
}
