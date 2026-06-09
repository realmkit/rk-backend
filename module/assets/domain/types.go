package domain

// Namespace separates asset trees by product area or tenant concept.
type Namespace string

// VirtualPath is a slash-separated virtual folder path inside a namespace.
type VirtualPath string

// Filename is the original or canonical asset filename.
type Filename string

// Visibility describes who may resolve an asset URL.
type Visibility string

// Status describes upload and moderation state.
type Status string

const (
	// VisibilityPublic allows public URL resolution.
	VisibilityPublic Visibility = "public"

	// VisibilityAuthenticated requires an authenticated principal.
	VisibilityAuthenticated Visibility = "authenticated"

	// VisibilityPrivate restricts access to owner or authorized staff.
	VisibilityPrivate Visibility = "private"
)

const (
	// StatusPendingUpload means the database row exists but storage object is not confirmed.
	StatusPendingUpload Status = "pending_upload"

	// StatusAvailable means the storage object exists and may be served.
	StatusAvailable Status = "available"

	// StatusQuarantined means the asset is blocked by moderation or scanning.
	StatusQuarantined Status = "quarantined"

	// StatusFailed means upload completion failed.
	StatusFailed Status = "failed"
)
