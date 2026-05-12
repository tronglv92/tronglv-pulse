package security

import (
	"testing"
)

func TestHashPath_Determinism(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"simple get", "GET", "/api/users"},
		{"post request", "POST", "/api/users"},
		{"with params", "DELETE", "/api/users/:id"},
		{"nested path", "GET", "/api/v1/users/profile"},
		{"root path", "GET", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Same input should always produce same hash
			hash1 := HashPath(tt.method, tt.path)
			hash2 := HashPath(tt.method, tt.path)

			if hash1 != hash2 {
				t.Errorf("HashPath not deterministic: got %s and %s", hash1, hash2)
			}

			// Hash should be 32 characters (truncated SHA-256 hex)
			if len(hash1) != 32 {
				t.Errorf("Expected hash length 32, got %d", len(hash1))
			}
		})
	}
}

func TestHashPath_Normalization(t *testing.T) {
	tests := []struct {
		name          string
		method1       string
		path1         string
		method2       string
		path2         string
		shouldBeEqual bool
	}{
		{
			name:          "case insensitive method",
			method1:       "GET",
			path1:         "/api/users",
			method2:       "get",
			path2:         "/api/users",
			shouldBeEqual: true,
		},
		{
			name:          "case insensitive path",
			method1:       "GET",
			path1:         "/API/USERS",
			method2:       "GET",
			path2:         "/api/users",
			shouldBeEqual: true,
		},
		{
			name:          "trailing slash removed",
			method1:       "GET",
			path1:         "/api/users/",
			method2:       "GET",
			path2:         "/api/users",
			shouldBeEqual: true,
		},
		{
			name:          "whitespace trimmed",
			method1:       " GET ",
			path1:         " /api/users ",
			method2:       "GET",
			path2:         "/api/users",
			shouldBeEqual: true,
		},
		{
			name:          "different methods",
			method1:       "GET",
			path1:         "/api/users",
			method2:       "POST",
			path2:         "/api/users",
			shouldBeEqual: false,
		},
		{
			name:          "different paths",
			method1:       "GET",
			path1:         "/api/users",
			method2:       "GET",
			path2:         "/api/posts",
			shouldBeEqual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := HashPath(tt.method1, tt.path1)
			hash2 := HashPath(tt.method2, tt.path2)

			equal := hash1 == hash2
			if equal != tt.shouldBeEqual {
				t.Errorf("Expected equality=%v, got %v for hashes %s and %s",
					tt.shouldBeEqual, equal, hash1, hash2)
			}
		})
	}
}

func TestHashPath_Uniqueness(t *testing.T) {
	paths := []struct {
		method string
		path   string
	}{
		{"GET", "/api/users"},
		{"POST", "/api/users"},
		{"PUT", "/api/users"},
		{"DELETE", "/api/users"},
		{"GET", "/api/users/:id"},
		{"GET", "/api/posts"},
		{"GET", "/api/v1/users"},
		{"GET", "/api/v2/users"},
		{"PATCH", "/api/users/:id"},
	}

	hashes := make(map[string]struct{})
	for _, p := range paths {
		hash := HashPath(p.method, p.path)

		// Check for collisions
		if _, exists := hashes[hash]; exists {
			t.Errorf("Hash collision detected for %s:%s", p.method, p.path)
		}

		hashes[hash] = struct{}{}
	}

	// Should have unique hash for each path
	if len(hashes) != len(paths) {
		t.Errorf("Expected %d unique hashes, got %d", len(paths), len(hashes))
	}
}

func TestHashPath_EmptyAndRoot(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"empty path normalized to root", "GET", ""},
		{"root path", "GET", "/"},
		{"root with trailing slash", "GET", "//"},
	}

	// All should normalize to GET:/
	hashes := make(map[string]int)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashPath(tt.method, tt.path)
			hashes[hash]++

			// Should not be empty
			if hash == "" {
				t.Errorf("Hash should not be empty for %s:%s", tt.method, tt.path)
			}
		})
	}

	// Empty path and "/" should produce same hash
	if len(hashes) != 1 {
		t.Errorf("Expected all root paths to normalize to same hash, got %d different hashes", len(hashes))
	}
}

func BenchmarkHashPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HashPath("GET", "/api/users/:id/profile")
	}
}

func TestPathToServiceName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "path with uid param",
			path:     "/outbound-ticket-svc/api/v1/outbound-tickets/:uid",
			expected: "OutboundTicketsService",
		},
		{
			name:     "assign action",
			path:     "/outbound-ticket-svc/api/v1/outbound-tickets/assign",
			expected: "OutboundTicketsService",
		},
		{
			name:     "nested action",
			path:     "/outbound-ticket-svc/api/v1/outbound-tickets/assign/tests",
			expected: "OutboundTicketsService",
		},
		{
			name:     "base resource path",
			path:     "/outbound-ticket-svc/api/v1/outbound-tickets",
			expected: "OutboundTicketsService",
		},
		{
			name:     "different version",
			path:     "/outbound-ticket-svc/api/v2/outbound-tickets",
			expected: "OutboundTicketsService",
		},
		{
			name:     "no service prefix",
			path:     "/api/v1/outbound-tickets",
			expected: "OutboundTicketsService",
		},
		{
			name:     "unknown path",
			path:     "/",
			expected: "DefaultService",
		},
		{
			name:     "empty path",
			path:     "",
			expected: "DefaultService",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathToServiceName(tt.path)
			if result != tt.expected {
				t.Fatalf("PathToServiceName(%q) = %q, want %q",
					tt.path, result, tt.expected)
			}
		})
	}
}
