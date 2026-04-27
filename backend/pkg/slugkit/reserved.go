package slugkit

// reserved enumerates slug values that conflict with built-in routes,
// public marketing pages, or system endpoints. Both the Go and TypeScript
// reserved sets MUST stay in sync — see web/src/lib/slug/reserved.ts.
//
// Categories:
//   - Auth flows: auth, login, logout, register, verify-email, ...
//   - Dashboard chrome: settings, support, dashboard, billing
//   - Marketing pages (web/src/app/*): about, blog, careers, ...
//   - API/system paths: api, www, app
//   - Onboarding/special endpoints: onboarding, personal, invite
//   - Polite booby-traps that look like literals: me, new, null, true, false, undefined
var reserved = map[string]struct{}{
	// Auth flows
	"auth":             {},
	"forgot-password":  {},
	"invite":           {},
	"login":            {},
	"logout":           {},
	"onboarding":       {},
	"register":         {},
	"reset-password":   {},
	"verify-email":     {},
	// Dashboard chrome
	"admin":     {},
	"billing":   {},
	"dashboard": {},
	"settings":  {},
	"support":   {},
	// Marketing / static pages
	"about":         {},
	"blog":          {},
	"careers":       {},
	"changelog":     {},
	"demo":          {},
	"docs":          {},
	"enterprise":    {},
	"mock-checkout": {},
	"offline":       {},
	"popout":        {},
	"privacy":       {},
	"terms":         {},
	// API / system
	"api": {},
	"app": {},
	"www": {},
	// Special endpoints / collection-level paths
	"organizations": {},
	"orgs":          {},
	"personal":      {},
	"runners":       {},
	"agents":        {},
	// Literal-looking traps
	"me":        {},
	"new":       {},
	"null":      {},
	"true":      {},
	"false":     {},
	"undefined": {},
}

func IsReserved(s string) bool {
	_, ok := reserved[s]
	return ok
}

// ReservedList returns a deterministic slice of all reserved slugs.
// Used by cross-language drift tests.
func ReservedList() []string {
	out := make([]string, 0, len(reserved))
	for k := range reserved {
		out = append(out, k)
	}
	return out
}
