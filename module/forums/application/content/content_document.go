package content

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/realmkit/rk-backend/module/forums/domain"
)

// checksum supports package behavior.
func checksum(provided string, content []byte) string {
	if strings.TrimSpace(provided) != "" {
		return strings.TrimSpace(provided)
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// prepareReferences supports package behavior.
func prepareReferences(
	sourcePostID uuid.UUID,
	references []domain.PostReference,
) []domain.PostReference {
	prepared := make([]domain.PostReference, 0, len(references))
	for _, reference := range references {
		reference.ID = uuid.New()
		reference.SourcePostID = sourcePostID
		prepared = append(prepared, reference)
	}
	return prepared
}

// contentText supports package behavior.
func contentText(explicit string, document []byte) string {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit)
	}
	var payload any
	if err := json.Unmarshal(document, &payload); err != nil {
		return ""
	}
	var parts []string
	collectText(payload, &parts)
	return strings.TrimSpace(strings.Join(parts, " "))
}

// collectText supports package behavior.
func collectText(value any, parts *[]string) {
	switch typed := value.(type) {
	case map[string]any:
		if text, ok := typed["text"].(string); ok && strings.TrimSpace(text) != "" {
			*parts = append(*parts, strings.TrimSpace(text))
		}
		for _, nested := range typed {
			collectText(nested, parts)
		}
	case []any:
		for _, nested := range typed {
			collectText(nested, parts)
		}
	}
}

// extractReferences supports package behavior.
func extractReferences(document []byte) []domain.PostReference {
	var payload any
	if err := json.Unmarshal(document, &payload); err != nil {
		return nil
	}
	var references []domain.PostReference
	collectReferences(payload, &references)
	return references
}

// collectReferences supports package behavior.
func collectReferences(value any, references *[]domain.PostReference) {
	switch typed := value.(type) {
	case map[string]any:
		appendNodeReference(typed, references)
		for _, nested := range typed {
			collectReferences(nested, references)
		}
	case []any:
		for _, nested := range typed {
			collectReferences(nested, references)
		}
	}
}

// appendNodeReference supports package behavior.
func appendNodeReference(
	node map[string]any,
	references *[]domain.PostReference,
) {
	nodeType, _ := node["type"].(string)
	attrs, _ := node["attrs"].(map[string]any)
	switch nodeType {
	case "mention":
		appendMentionReference(attrs, references)
	case "attachment":
		appendAttachmentReference(attrs, references)
	case "quote":
		appendQuoteReference(attrs, references)
	case "reply_to":
		appendReplyReference(attrs, references)
	case "link":
		appendLinkReference(attrs, references)
	}
}

// appendMentionReference supports package behavior.
func appendMentionReference(
	attrs map[string]any,
	references *[]domain.PostReference,
) {
	if id := uuidFromAttr(attrs, "id", "user_id"); id != uuid.Nil {
		*references = append(*references, domain.PostReference{
			TargetUserID:  &id,
			ReferenceType: domain.ReferenceMention,
		})
	}
}

// appendAttachmentReference supports package behavior.
func appendAttachmentReference(
	attrs map[string]any,
	references *[]domain.PostReference,
) {
	if id := uuidFromAttr(attrs, "asset_id", "id"); id != uuid.Nil {
		*references = append(*references, domain.PostReference{
			TargetAssetID: &id,
			ReferenceType: domain.ReferenceAttachment,
		})
	}
}

// appendQuoteReference supports package behavior.
func appendQuoteReference(
	attrs map[string]any,
	references *[]domain.PostReference,
) {
	if id := uuidFromAttr(attrs, "post_id", "id"); id != uuid.Nil {
		excerpt, _ := attrs["excerpt"].(string)
		*references = append(*references, domain.PostReference{
			TargetPostID:  &id,
			ReferenceType: domain.ReferenceQuote,
			QuoteExcerpt:  excerpt,
		})
	}
}

// appendReplyReference supports package behavior.
func appendReplyReference(
	attrs map[string]any,
	references *[]domain.PostReference,
) {
	if id := uuidFromAttr(attrs, "post_id", "id"); id != uuid.Nil {
		*references = append(*references, domain.PostReference{
			TargetPostID:  &id,
			ReferenceType: domain.ReferenceReplyTo,
		})
	}
}

// appendLinkReference supports package behavior.
func appendLinkReference(
	attrs map[string]any,
	references *[]domain.PostReference,
) {
	href, _ := attrs["href"].(string)
	if strings.TrimSpace(href) != "" {
		*references = append(*references, domain.PostReference{
			ReferenceType: domain.ReferenceLink,
			LinkURL:       strings.TrimSpace(href),
		})
	}
}

// uuidFromAttr supports package behavior.
func uuidFromAttr(attrs map[string]any, keys ...string) uuid.UUID {
	for _, key := range keys {
		if raw, ok := attrs[key].(string); ok {
			parsed, err := uuid.Parse(raw)
			if err == nil {
				return parsed
			}
		}
	}
	return uuid.Nil
}
