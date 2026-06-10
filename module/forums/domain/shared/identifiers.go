package shared

import "github.com/google/uuid"

// Key is a stable forum machine key.
type Key string

// Slug is a URL-friendly forum slug.
type Slug string

// rootForumObjectID is the reserved permission object for structure administration.
const rootForumObjectID = "00000000-0000-0000-0000-000000000101"

// RootForumObjectID returns the reserved forum permission target.
func RootForumObjectID() uuid.UUID {
	return uuid.MustParse(rootForumObjectID)
}
