package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// executeCommand is a helper to run a cobra command and capture its output
func executeCommand(args ...string) (string, string, error) {
	// Redirect log output for capture
	var out, errOut bytes.Buffer
	log.SetOutput(&errOut)
	defer log.SetOutput(os.Stderr) // Restore default logger

	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errOut)
	rootCmd.SetArgs(args)

	err := rootCmd.Execute()

	return out.String(), errOut.String(), err
}

func TestRootCmd_SingleArgument(t *testing.T) {
	t.Run("when argument is a directory, creates config file", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.yaml")

		_, errOut, err := executeCommand(tempDir)
		if err != nil {
			t.Fatalf("command execution failed: %v, output: %s", err, errOut)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Errorf("expected config file to be created at %s, but it wasn't", configPath)
		}

		// Check for log output indicating creation
		if !strings.Contains(errOut, "Creating default config") {
			t.Errorf("expected log output to contain 'Creating default config', but got: %s", errOut)
		}
	})

	t.Run("when argument is a file, assumes it's a config and tries to run", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "test-config.yaml")
		dbPath := filepath.Join(tempDir, "test.db")
		imagesPath := filepath.Join(tempDir, "images")

		// Create a valid config file
		validConfig := `
meta:
  description: "Sample annotation project."
auth:
  admin: { password: "changeme" }
  annotator: { password: "changeme" }
tasks:
  - id: quality
    name: "Image Quality Assessment"
    classes:
      good: { name: "Good" }
      bad: { name: "Bad" }
`
		os.WriteFile(configPath, []byte(validConfig), 0644)
		os.Mkdir(imagesPath, 0755)

		// We expect an error because the server can't actually start in a test env,
		// but we can check the log output to see if it tried.
		// The error "listen tcp :8080: bind: address already in use" is a good sign.
		// Or, if we get a timeout, that's also a sign it tried to start.
		// For this test, we'll just check that it doesn't fail immediately with a config error.
		// The most reliable check is for the log output saying it's starting the server.
		_, errOut, err := executeCommand(configPath, "--database", dbPath, "--images", imagesPath, "--addr", ":8082")

		// We expect an error because the server will be interrupted or fail to bind in test.
		// The key is that it shouldn't be a "config file not found" or "database flag required" error.
		if err == nil {
			t.Log("command did not return an error, which is unexpected but could be ok if it timed out")
		}

		if strings.Contains(errOut, "database flag is required") {
			t.Errorf("should not have prompted for database flag, got: %s", errOut)
		}

		if !strings.Contains(errOut, "Starting server on: :8082") {
			t.Errorf("expected log to show server starting, but it didn't. Got: %s", errOut)
		}
	})

	t.Run("when argument is an invalid path, returns an error", func(t *testing.T) {
		invalidPath := "/path/to/some/nonexistent/dir"
		_, _, err := executeCommand(invalidPath, "-d", "a", "-i", "b")

		if err == nil {
			t.Fatal("expected an error for invalid path, but got none")
		}

		// The error should be about the config file, since it assumes the arg is a config file
		if !strings.Contains(err.Error(), "failed to load config") {
			t.Errorf("expected error to be about loading config, but got: %v", err)
		}
	})
}
