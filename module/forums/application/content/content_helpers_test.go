package content

import (
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
)

// TestContentTextExtractsNestedText verifies text extraction from JSON documents.
func TestContentTextExtractsNestedText(t *testing.T) {
	document := []byte(`{"type":"doc","content":[{"type":"text","text":" Hello "},{"content":[{"text":"World"}]}]}`)
	if got := contentText("", document); got != "Hello World" {
		t.Fatalf("contentText() = %q, want Hello World", got)
	}
	if got := contentText(" explicit ", document); got != "explicit" {
		t.Fatalf("contentText(explicit) = %q, want explicit", got)
	}
}

// TestExtractReferencesFindsSupportedNodes verifies document reference extraction.
func TestExtractReferencesFindsSupportedNodes(t *testing.T) {
	userID := uuid.New()
	assetID := uuid.New()
	postID := uuid.New()
	document := []byte(`{"content":[` +
		`{"type":"mention","attrs":{"user_id":"` + userID.String() + `"}},` +
		`{"type":"attachment","attrs":{"asset_id":"` + assetID.String() + `"}},` +
		`{"type":"quote","attrs":{"post_id":"` + postID.String() + `","excerpt":"hi"}},` +
		`{"type":"link","attrs":{"href":" https://example.test "}}` +
		`]}`)
	references := extractReferences(document)
	if len(references) != 4 {
		t.Fatalf("extractReferences() length = %d, want 4", len(references))
	}
	if references[0].ReferenceType != domain.ReferenceMention || *references[0].TargetUserID != userID {
		t.Fatalf("mention reference = %#v", references[0])
	}
	if references[1].ReferenceType != domain.ReferenceAttachment || *references[1].TargetAssetID != assetID {
		t.Fatalf("attachment reference = %#v", references[1])
	}
}

// TestPrepareReferencesAssignsSource verifies source metadata is set on references.
func TestPrepareReferencesAssignsSource(t *testing.T) {
	sourceID := uuid.New()
	prepared := prepareReferences(sourceID, []domain.PostReference{{ReferenceType: domain.ReferenceLink}})
	if len(prepared) != 1 || prepared[0].ID == uuid.Nil || prepared[0].SourcePostID != sourceID {
		t.Fatalf("prepareReferences() = %#v", prepared)
	}
}
