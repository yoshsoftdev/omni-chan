package storage

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestValidateObjectKey(t *testing.T) {
	cases := []struct {
		key   string
		valid bool
	}{
		{"attachments/team/2026/06/file.jpg", true},
		{"", false},
		{"/absolute/path", false},
		{"../escape", false},
		{"attachments/../secret", false},
	}
	for _, tc := range cases {
		err := ValidateObjectKey(tc.key)
		if tc.valid && err != nil {
			t.Fatalf("expected valid key %q, got %v", tc.key, err)
		}
		if !tc.valid && err == nil {
			t.Fatalf("expected invalid key %q", tc.key)
		}
	}
}

func TestBuildObjectKey(t *testing.T) {
	teamID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	key, err := BuildObjectKey("attachments", teamID, "photo.JPG")
	if err != nil {
		t.Fatalf("BuildObjectKey() error = %v", err)
	}
	if !strings.HasPrefix(key, "attachments/11111111-1111-4111-8111-111111111111/") {
		t.Fatalf("unexpected prefix: %s", key)
	}
	if !strings.HasSuffix(key, ".jpg") {
		t.Fatalf("expected lowercase extension, got %s", key)
	}
}
