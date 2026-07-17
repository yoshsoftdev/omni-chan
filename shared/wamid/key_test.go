package wamid

import "testing"

func TestExtractMessageKey_WAToWAReply(t *testing.T) {
	original := "wamid.HBgNNjI4MjI4ODgxMjAwMxUCABIYFDNCMkJGRjRFNTA3RDNBOTJBQUZFAA=="
	replyContext := "wamid.HBgUSUQuMzcwMzA4MjQ4MDk4MzgxNDYVFAASGBQzQjJCRkY0RTUwN0QzQTkyQUFGRQA="
	replyMessage := "wamid.HBgNNjI4MjI4ODgxMjAwMxUCABIYFDNCMkJDOThDMUUwNjI4MEUzREIzAA=="

	originalKey := ExtractMessageKey(original)
	contextKey := ExtractMessageKey(replyContext)
	replyKey := ExtractMessageKey(replyMessage)

	if originalKey != "3B2BFF4E507D3A92AAFE" {
		t.Fatalf("original key=%q", originalKey)
	}
	if contextKey != originalKey {
		t.Fatalf("context key=%q want %q", contextKey, originalKey)
	}
	if replyKey == originalKey {
		t.Fatalf("reply message should have its own key, got %q", replyKey)
	}
	if !KeysMatch(original, replyContext) {
		t.Fatal("expected original and reply context to match")
	}
}

func TestKeysMatch_ExactWAMID(t *testing.T) {
	if !KeysMatch("wamid-1", "wamid-1") {
		t.Fatal("expected exact simple wamid match")
	}
}

func TestExtractMessageKey_Empty(t *testing.T) {
	if ExtractMessageKey("") != "" {
		t.Fatal("expected empty key")
	}
}

func TestExtractMessageKey_LongInstagramMID(t *testing.T) {
	mid := "aWdfZAG1faXRlbToxOklGRXh0OjE3ODQxNDc0MTU5NzAxMDE4OjE4MjU4NzM5MjM3MjQ1MzU0OjE3MjEyMzQ1Njc4OWFiY2RlZgZDZD"
	key := ExtractMessageKey(mid)
	if key == "" {
		t.Fatal("expected non-empty key")
	}
	if len(key) > 32 {
		t.Fatalf("key length %d exceeds wamid_key VARCHAR(32)", len(key))
	}
	if ExtractMessageKey(mid) != key {
		t.Fatal("expected stable hash for same mid")
	}
	if !KeysMatch(mid, mid) {
		t.Fatal("expected KeysMatch for identical long mid")
	}
}
