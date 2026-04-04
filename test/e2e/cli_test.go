//go:build e2e
// +build e2e

package e2e_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

var binaryPath = "../cloud189"

func TestMain(m *testing.M) {
	if os.Getenv("CLOUD189_BINARY") != "" {
		binaryPath = os.Getenv("CLOUD189_BINARY")
	}

	os.Exit(m.Run())
}

func runCloud189(args ...string) (string, error) {
	cmd := exec.Command(binaryPath, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestE2EVersion(t *testing.T) {
	output, err := runCloud189("version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	if !strings.Contains(output, "cloud189 CLI") {
		t.Errorf("version output missing expected text: %s", output)
	}

	t.Logf("Version output: %s", strings.TrimSpace(output))
}

func TestE2EWhoamiNotLoggedIn(t *testing.T) {
	if os.Getenv("CLOUD189_TEST_LOGGED_IN") != "" {
		t.Skip("Skipping because logged in")
	}

	output, _ := runCloud189("whoami")

	if strings.Contains(output, `"success": true`) {
		t.Error("whoami should show not logged in")
	}

	t.Logf("whoami output: %s", output)
}

func TestE2EInfoCommand(t *testing.T) {
	if os.Getenv("CLOUD189_TEST_LOGGED_IN") == "" {
		t.Skip("Set CLOUD189_TEST_LOGGED_IN=1 when logged in")
	}

	output, err := runCloud189("info", "-o", "json")
	if err != nil {
		t.Fatalf("info command failed: %v", err)
	}

	if !strings.Contains(output, "personal") && !strings.Contains(output, "family") {
		t.Errorf("info output missing expected fields: %s", output)
	}

	t.Logf("info output: %s", output)
}

func TestE2EListRoot(t *testing.T) {
	if os.Getenv("CLOUD189_TEST_LOGGED_IN") == "" {
		t.Skip("Set CLOUD189_TEST_LOGGED_IN=1 when logged in")
	}

	output, err := runCloud189("ls", "/", "-o", "json")
	if err != nil {
		t.Fatalf("ls command failed: %v", err)
	}

	if !strings.Contains(output, "files") {
		t.Errorf("ls output missing 'files' field: %s", output)
	}

	t.Logf("ls / output: %s", output)
}

func TestE2EListWithYAML(t *testing.T) {
	if os.Getenv("CLOUD189_TEST_LOGGED_IN") == "" {
		t.Skip("Set CLOUD189_TEST_LOGGED_IN=1 when logged in")
	}

	output, err := runCloud189("ls", "/", "-o", "yaml")
	if err != nil {
		t.Fatalf("ls yaml command failed: %v", err)
	}

	if !strings.Contains(output, "files:") {
		t.Errorf("yaml output missing expected format: %s", output)
	}
}

func TestE2EListWithTable(t *testing.T) {
	if os.Getenv("CLOUD189_TEST_LOGGED_IN") == "" {
		t.Skip("Set CLOUD189_TEST_LOGGED_IN=1 when logged in")
	}

	output, err := runCloud189("ls", "/", "-o", "table")
	if err != nil {
		t.Fatalf("ls table command failed: %v", err)
	}

	t.Logf("Table output:\n%s", output)
}

func TestE2EMkdirAndRm(t *testing.T) {
	if os.Getenv("CLOUD189_TEST_LOGGED_IN") == "" {
		t.Skip("Set CLOUD189_TEST_LOGGED_IN=1 when logged in")
	}

	folderName := "e2e_test_folder"

	output, err := runCloud189("mkdir", folderName, "-o", "json")
	if err != nil {
		t.Fatalf("mkdir command failed: %v, output: %s", err, output)
	}

	t.Logf("mkdir output: %s", output)

	output, err = runCloud189("rm", folderName, "-f", "-o", "json")
	if err != nil {
		t.Fatalf("rm command failed: %v, output: %s", err, output)
	}

	t.Logf("rm output: %s", output)
}

func TestE2EFamilyList(t *testing.T) {
	if os.Getenv("CLOUD189_TEST_LOGGED_IN") == "" {
		t.Skip("Set CLOUD189_TEST_LOGGED_IN=1 when logged in")
	}

	output, err := runCloud189("family", "list", "-o", "json")
	if err != nil {
		t.Fatalf("family list command failed: %v", err)
	}

	t.Logf("family list output: %s", output)
}

func TestE2EHelpCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"root help", []string{"--help"}},
		{"ls help", []string{"ls", "--help"}},
		{"mkdir help", []string{"mkdir", "--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCloud189(tt.args...)
			if err != nil {
				t.Errorf("help command failed: %v", err)
			}

			if !strings.Contains(output, "Usage") && !strings.Contains(output, "Flags") {
				t.Errorf("Help output missing expected sections: %s", output)
			}
		})
	}
}

func TestE2EInvalidCommand(t *testing.T) {
	output, _ := runCloud189("invalid_command_xyz")

	if output == "" {
		t.Error("Invalid command should produce error output")
	}
}

func TestE2EOutputFormats(t *testing.T) {
	if os.Getenv("CLOUD189_TEST_LOGGED_IN") == "" {
		t.Skip("Set CLOUD189_TEST_LOGGED_IN=1 when logged in")
	}

	tests := []struct {
		name   string
		format string
		check  func(string) bool
	}{
		{
			name:   "json format",
			format: "json",
			check: func(s string) bool {
				return strings.Contains(s, "{") && strings.Contains(s, "}")
			},
		},
		{
			name:   "yaml format",
			format: "yaml",
			check: func(s string) bool {
				return strings.Contains(s, ":")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCloud189("whoami", "-o", tt.format)
			if err != nil {
				t.Logf("Command may have failed (expected if not logged in): %v", err)
			}

			if !tt.check(output) {
				t.Errorf("Output format check failed for %s: %s", tt.format, output)
			}
		})
	}
}
