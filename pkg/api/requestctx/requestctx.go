package requestctx

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	// LocalProfile stores the selected request timeout profile name.
	LocalProfile = "realmkit.request_context.profile"

	// LocalDeadline stores the selected request deadline.
	LocalDeadline = "realmkit.request_context.deadline"
)

// Profile describes one request timeout profile.
type Profile struct {
	// Name is the operator-facing timeout profile identifier.
	Name string

	// Timeout is the maximum duration for the request.
	Timeout time.Duration
}

// Option changes middleware behavior.
type Option func(*settings)

// Skipper reports whether a request should skip timeout wrapping.
type Skipper func(*fiber.Ctx) bool

// routeProfile binds one exact route to a profile.
type routeProfile struct {
	method  string
	path    string
	prefix  bool
	profile Profile
}

// settings contains request context middleware configuration.
type settings struct {
	defaultProfile Profile
	profiles       []routeProfile
	skipper        Skipper
}

// Middleware creates Fiber middleware that installs a deadline-bound user context.
func Middleware(timeout time.Duration, options ...Option) fiber.Handler {
	cfg := settings{defaultProfile: Profile{Name: "standard", Timeout: timeout}}
	for _, option := range options {
		option(&cfg)
	}
	return func(ctx *fiber.Ctx) error {
		if cfg.skipper != nil && cfg.skipper(ctx) {
			return ctx.Next()
		}
		profile := cfg.profileFor(ctx)
		if profile.Timeout <= 0 {
			return ctx.Next()
		}
		parent := ctx.UserContext()
		child, cancel := context.WithTimeout(parent, profile.Timeout)
		defer cancel()
		deadline, _ := child.Deadline()
		ctx.SetUserContext(child)
		ctx.Locals(LocalProfile, profile.Name)
		ctx.Locals(LocalDeadline, deadline)
		return ctx.Next()
	}
}

// WithRouteProfile configures an exact method and path timeout profile.
func WithRouteProfile(method string, path string, profile Profile) Option {
	return func(settings *settings) {
		settings.profiles = append(settings.profiles, routeProfile{
			method:  strings.ToUpper(strings.TrimSpace(method)),
			path:    strings.TrimSpace(path),
			profile: profile,
		})
	}
}

// WithPathPrefixProfile configures a method and path prefix timeout profile.
func WithPathPrefixProfile(method string, path string, profile Profile) Option {
	return func(settings *settings) {
		settings.profiles = append(settings.profiles, routeProfile{
			method:  strings.ToUpper(strings.TrimSpace(method)),
			path:    strings.TrimSpace(path),
			prefix:  true,
			profile: profile,
		})
	}
}

// WithSkipper configures requests that should not receive a timeout wrapper.
func WithSkipper(skipper Skipper) Option {
	return func(settings *settings) {
		settings.skipper = skipper
	}
}

// CurrentDeadline returns the request deadline stored by Middleware.
func CurrentDeadline(ctx *fiber.Ctx) (time.Time, bool) {
	value, ok := ctx.Locals(LocalDeadline).(time.Time)
	return value, ok && !value.IsZero()
}

// CurrentProfile returns the request timeout profile stored by Middleware.
func CurrentProfile(ctx *fiber.Ctx) (string, bool) {
	value, ok := ctx.Locals(LocalProfile).(string)
	return value, ok && value != ""
}

// profileFor returns the configured profile for ctx.
func (settings settings) profileFor(ctx *fiber.Ctx) Profile {
	method := strings.ToUpper(ctx.Method())
	path := ctx.Path()
	routePath := ctx.Route().Path
	for _, route := range settings.profiles {
		if !route.matchesMethod(method) {
			continue
		}
		if route.matchesPath(path, routePath) {
			return route.profile
		}
	}
	return settings.defaultProfile
}

// matchesMethod reports whether route applies to method.
func (route routeProfile) matchesMethod(method string) bool {
	return route.method == "" || route.method == method
}

// matchesPath reports whether route applies to a request or route path.
func (route routeProfile) matchesPath(path string, routePath string) bool {
	if !route.prefix {
		return route.path == path || route.path == routePath
	}
	return strings.HasPrefix(path, route.path) || strings.HasPrefix(routePath, route.path)
}
