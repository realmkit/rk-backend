package search

import "testing"

// benchmarkTextQuery stores search query benchmark output.
var benchmarkTextQuery TextQuery

// benchmarkCursorToken stores encoded cursor benchmark output.
var benchmarkCursorToken string

// benchmarkCursor stores decoded cursor benchmark output.
var benchmarkCursor Cursor

// BenchmarkNewTextQuery measures user search text normalization and validation.
func BenchmarkNewTextQuery(b *testing.B) {
	raw := "  RealmKit    forum   moderation   "

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		query, err := NewTextQuery(raw, QueryOptions{})
		if err != nil {
			b.Fatalf("NewTextQuery() error = %v", err)
		}
		benchmarkTextQuery = query
	}
}

// BenchmarkCursorRoundTrip measures cursor JSON/base64 encode and decode.
func BenchmarkCursorRoundTrip(b *testing.B) {
	cursor := Cursor{
		FilterHash: HashFilter("forum", "public", 25),
		Sort:       "created_at",
		Direction:  DirectionDesc,
		Values:     []string{"2026-06-20T00:00:00Z"},
		ID:         "018f68e8-8d60-7c0c-a1c6-833f6503ec4d",
	}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		token, err := EncodeCursor(cursor)
		if err != nil {
			b.Fatalf("EncodeCursor() error = %v", err)
		}
		decoded, err := DecodeCursor(token)
		if err != nil {
			b.Fatalf("DecodeCursor() error = %v", err)
		}
		benchmarkCursorToken = token
		benchmarkCursor = decoded
	}
}
