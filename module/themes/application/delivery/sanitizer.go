package delivery

import (
	"slices"

	"github.com/realmkit/rk-backend/module/themes/domain"
	"github.com/realmkit/rk-backend/module/themes/port"
)

// SanitizerProfile describes one backend rich-text boundary policy.
type SanitizerProfile struct {
	Profile            domain.RichTextProfile // Profile stores the profile value.
	AllowedElements    []string               // AllowedElements stores the allowed elements value.
	AllowedAttributes  map[string][]string    // AllowedAttributes stores the allowed attributes value.
	RealmKitImagesOnly bool                   // RealmKitImagesOnly stores the realm kit images only value.
}

// SanitizerProfiles returns reusable rich-text sanitizer policies.
func SanitizerProfiles() []SanitizerProfile {
	return []SanitizerProfile{
		profile(domain.ProfileForumPost, forumPostElements()),
		profile(domain.ProfileForumDescription, []string{"br", "strong", "em", "u", "s", "a", "code"}),
		profile(domain.ProfileStaticPage, append(forumPostElements(), "section", "article", "aside")),
		profile(domain.ProfileTicketText, []string{"p", "br", "strong", "em", "a", "ul", "ol", "li", "blockquote", "code"}),
		profile(domain.ProfilePunishmentText, []string{"p", "br", "strong", "em", "a", "ul", "ol", "li", "code"}),
		profile(domain.ProfileSignature, []string{"br", "strong", "em", "a", "code"}),
	}
}

// SanitizerProfileFor returns one sanitizer profile.
func SanitizerProfileFor(profileName domain.RichTextProfile) (SanitizerProfile, error) {
	for _, profile := range SanitizerProfiles() {
		if profile.Profile == profileName {
			return profile, nil
		}
	}
	return SanitizerProfile{}, port.ErrNotFound
}

// AllowsElement reports whether the sanitizer allows an element.
func (profile SanitizerProfile) AllowsElement(element string) bool {
	return slices.Contains(profile.AllowedElements, element)
}

// profile creates a sanitizer policy with shared safe attributes.
func profile(profileName domain.RichTextProfile, elements []string) SanitizerProfile {
	return SanitizerProfile{
		Profile:         profileName,
		AllowedElements: elements,
		AllowedAttributes: map[string][]string{
			"a":    {"href", "title", "target", "rel"},
			"img":  {"src", "alt", "width", "height"},
			"code": {"data-language"},
			"pre":  {"data-language"},
		},
		RealmKitImagesOnly: true,
	}
}

// forumPostElements returns the broad forum-rich-text allowlist.
func forumPostElements() []string {
	return []string{"p", "br", "strong", "em", "u", "s", "a", "ul", "ol", "li", "blockquote", "code", "pre", "h2", "h3", "h4", "table", "thead", "tbody", "tr", "th", "td", "img"}
}
