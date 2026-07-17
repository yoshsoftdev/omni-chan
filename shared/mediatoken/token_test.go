package mediatoken

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMintParseAndBuildWebhookURL(t *testing.T) {
	issuer := NewIssuer("secret", 5*time.Minute)
	objectID := uuid.New()
	teamID := uuid.New()

	token, _, err := issuer.Mint(objectID, teamID)
	if err != nil {
		t.Fatal(err)
	}
	gotObject, gotTeam, err := issuer.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if gotObject != objectID || gotTeam != teamID {
		t.Fatalf("parsed ids mismatch")
	}

	url := BuildWebhookURL("https://webhook.example.com/", token)
	if !strings.HasPrefix(url, "https://webhook.example.com/media/eyJ") {
		t.Fatalf("url=%q", url)
	}
}
