package content

import (
	"testing"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
)

// benchmarkContentText stores extracted content text benchmark output.
var benchmarkContentText string

// benchmarkPostReferences stores extracted post references benchmark output.
var benchmarkPostReferences []domain.PostReference

// BenchmarkContentText measures nested rich-content text extraction.
func BenchmarkContentText(b *testing.B) {
	document := benchmarkDocument()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkContentText = contentText("", document)
	}
}

// BenchmarkExtractReferences measures supported rich-content reference extraction.
func BenchmarkExtractReferences(b *testing.B) {
	document := benchmarkDocument()

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkPostReferences = extractReferences(document)
	}
}

// benchmarkDocument returns a representative ProseMirror-style document.
func benchmarkDocument() []byte {
	userID := uuid.NewString()
	assetID := uuid.NewString()
	postID := uuid.NewString()
	return []byte(`{"type":"doc","content":[` +
		`{"type":"paragraph","content":[{"type":"text","text":"Welcome "},{"type":"mention","attrs":{"user_id":"` + userID + `"}}]},` +
		`{"type":"attachment","attrs":{"asset_id":"` + assetID + `"}},` +
		`{"type":"quote","attrs":{"post_id":"` + postID + `","excerpt":"quoted text"}},` +
		`{"type":"reply_to","attrs":{"post_id":"` + postID + `"}},` +
		`{"type":"link","attrs":{"href":" https://example.test/guide "}}` +
		`]}`)
}
