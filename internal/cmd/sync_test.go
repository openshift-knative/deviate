package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindGitRoot(t *testing.T) {
	// Helper function to create a temporary directory structure for testing
	setupTestDir := func(t *testing.T, gitLocationRelPath string, isDir bool) (string, string) {
		t.Helper()
		baseDir, err := os.MkdirTemp("", "testFindGitRoot_*")
		if err != nil {
			t.Fatalf("Failed to create temp base dir: %v", err)
		}

		// Default startPath deep within a nested structure
		startPath := filepath.Join(baseDir, "project", "subdir1", "subdir2")
		if err := os.MkdirAll(startPath, 0755); err != nil {
			os.RemoveAll(baseDir)
			t.Fatalf("Failed to create startPath %s: %v", startPath, err)
		}

		if gitLocationRelPath != "" {
			// Construct the full path for .git based on baseDir and the relative gitLocationRelPath
			var gitPath string
			// Handle cases where gitLocationRelPath is like ".git" (at baseDir) or "project/.git"
			if filepath.IsAbs(gitLocationRelPath) { // Should not happen with current test cases, but good for robustness
				gitPath = gitLocationRelPath
			} else {
				gitPath = filepath.Join(baseDir, gitLocationRelPath)
			}

			// Ensure the parent directory of the .git entity exists
			parentOfGit := filepath.Dir(gitPath)
			if err := os.MkdirAll(parentOfGit, 0755); err != nil {
				os.RemoveAll(baseDir)
				t.Fatalf("Failed to create parent directory for .git entity at %s: %v", parentOfGit, err)
			}

			if isDir {
				if err := os.Mkdir(gitPath, 0755); err != nil { // Use Mkdir for .git itself
					os.RemoveAll(baseDir)
					t.Fatalf("Failed to create .git directory at %s: %v", gitPath, err)
				}
			} else {
				f, err := os.Create(gitPath)
				if err != nil {
					os.RemoveAll(baseDir)
					t.Fatalf("Failed to create .git file at %s: %v", gitPath, err)
				}
				f.Close()
			}
		}
		return baseDir, startPath
	}

	tests := []struct {
		name                string
		gitLocationRelPath  string // Relative to temp baseDir. E.g., "project/.git" or ".git" (for baseDir/.git)
		isGitADir           bool
		startPathChoice     string // "deep", "projectRoot", "baseDir"
		expectedRootRelPath string // Expected root relative to baseDir. E.g., "project" or "" (for baseDir itself)
		expectError         bool
		expectedErrorMsg    string // Substring to look for in the error message
	}{
		{
			name:                "git in project subdir (start deep)",
			gitLocationRelPath:  "project/.git",
			isGitADir:           true,
			startPathChoice:     "deep", // starts at baseDir/project/subdir1/subdir2
			expectedRootRelPath: "project",
			expectError:         false,
		},
		{
			name:                "git at baseDir (start deep)",
			gitLocationRelPath:  ".git", // i.e. baseDir/.git
			isGitADir:           true,
			startPathChoice:     "deep", // starts at baseDir/project/subdir1/subdir2
			expectedRootRelPath: "",     // Expected to be baseDir
			expectError:         false,
		},
		{
			name:                "git in project subdir (start at projectRoot)",
			gitLocationRelPath:  "project/.git",
			isGitADir:           true,
			startPathChoice:     "projectRoot", // starts at baseDir/project
			expectedRootRelPath: "project",
			expectError:         false,
		},
		{
			name:               "no git directory found (start deep)",
			gitLocationRelPath: "",   // No .git created
			isGitADir:          true, // Irrelevant
			startPathChoice:    "deep",
			expectError:        true,
			expectedErrorMsg:   "'.git' directory not found",
		},
		{
			name:               "git is a file, not a directory (in project)",
			gitLocationRelPath: "project/.git",
			isGitADir:          false, // Create .git as a file
			startPathChoice:    "deep",
			expectError:        true,
			expectedErrorMsg:   ".git found at", // ".git found at ... but it is not a directory"
		},
		{
			name:                "git at baseDir (start at baseDir)",
			gitLocationRelPath:  ".git",
			isGitADir:           true,
			startPathChoice:     "baseDir", // starts at baseDir
			expectedRootRelPath: "",
			expectError:         false,
		},
		{
			name:               "no git at baseDir (start at baseDir)",
			gitLocationRelPath: "",
			isGitADir:          true,
			startPathChoice:    "baseDir",
			expectError:        true,
			expectedErrorMsg:   "'.git' directory not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir, defaultStartPath := setupTestDir(t, tt.gitLocationRelPath, tt.isGitADir)
			defer os.RemoveAll(baseDir)

			var currentStartPath string
			switch tt.startPathChoice {
			case "deep":
				currentStartPath = defaultStartPath
			case "projectRoot":
				currentStartPath = filepath.Join(baseDir, "project")
			case "baseDir":
				currentStartPath = baseDir
			default:
				currentStartPath = defaultStartPath
			}

			var expectedRootAbsPath string
			if tt.expectedRootRelPath == "" {
				expectedRootAbsPath = baseDir
			} else {
				expectedRootAbsPath = filepath.Join(baseDir, tt.expectedRootRelPath)
			}

			actualRoot, err := findGitRoot(currentStartPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("findGitRoot() was expected to return an error, but did not")
				} else if tt.expectedErrorMsg != "" && !strings.Contains(err.Error(), tt.expectedErrorMsg) {
					t.Errorf("findGitRoot() error = \"%v\", expected to contain \"%s\"", err, tt.expectedErrorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("findGitRoot() returned an unexpected error: %v", err)
				}
				normalizedActualRoot := filepath.Clean(actualRoot)
				normalizedExpectedRoot := filepath.Clean(expectedRootAbsPath)

				if normalizedActualRoot != normalizedExpectedRoot {
					t.Errorf("findGitRoot() actualRoot = %s, expectedRoot = %s", normalizedActualRoot, normalizedExpectedRoot)
				}
			}
		})
	}
}
