package main

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"os"
	"testing"
)

func TestExitCases(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		shouldErr bool
	}{
		{
			name: "ExitCalled",
			src: `package main

import "os"

func main() {
	os.Exit(1) // want "Direct call to os.Exit in main package"
}`,
			shouldErr: true,
		},
		{
			name: "ExitNotCalled",
			src: `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`,
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := t.TempDir() // Создаем временный каталог
			testFile := testDir + "/main.go"

			if err := os.WriteFile(testFile, []byte(tt.src), 0644); err != nil {
				t.Fatalf("Failed creating test file: %v", err)
			}

			result := analysistest.Run(t, testDir, ExitAnalyzer)

			hasError := len(result[0].Diagnostics) > 0

			if hasError != tt.shouldErr {
				t.Errorf("For test %s: expected error: %v, got: %v", tt.name, tt.shouldErr, hasError)
			}

			if hasError && tt.shouldErr {
				for _, diag := range result[0].Diagnostics {
					if diag.Message == "Direct call to os.Exit in main package" {
						return
					}
				}
				t.Errorf("For test %s: expected diagnostic not found", tt.name)
			}
		})
	}
}
