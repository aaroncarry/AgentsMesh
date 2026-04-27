package slugkit

import (
	"context"
	"fmt"
	"strings"
)

// UniquenessChecker returns true when candidate is available
// (not currently used in the relevant scope).
type UniquenessChecker func(ctx context.Context, candidate string) (available bool, err error)

// FromExistsCheck adapts a "does X exist" query to a UniquenessChecker.
// Use it to plug repository.SlugExists-style functions into GenerateUnique:
//
//	slugkit.GenerateUnique(ctx, raw, slugkit.FromExistsCheck(repo.SlugExists))
func FromExistsCheck(exists func(context.Context, string) (bool, error)) UniquenessChecker {
	return func(ctx context.Context, candidate string) (bool, error) {
		e, err := exists(ctx, candidate)
		if err != nil {
			return false, err
		}
		return !e, nil
	}
}

const (
	generatorMaxAttempts = 100
	// suffixReserve = len("-100"); ensures "%s-%d" up to N=100 stays within MaxLen.
	suffixReserve = 4
)

// GenerateUnique sanitizes raw, then probes candidates `base`, `base-2`, `base-3`...
// until one is available or attempts are exhausted.
//
// If raw sanitizes to a base whose length leaves no room for the "-N" suffix,
// the base is truncated so suffixed candidates still satisfy MaxLen.
//
// Candidates that happen to collide with reserved words (e.g. "ap-2" if it were
// ever reserved) are skipped, not fatal.
func GenerateUnique(ctx context.Context, raw string, check UniquenessChecker) (string, error) {
	base, err := SanitizeAndValidate(raw)
	if err != nil {
		return "", err
	}
	if len(base) > MaxLen-suffixReserve {
		truncated := strings.TrimRight(base[:MaxLen-suffixReserve], "-")
		if err := Validate(truncated); err != nil {
			return "", fmt.Errorf("base %q too long to suffix safely (truncation invalid): %w", base, err)
		}
		base = truncated
	}

	for i := 0; i < generatorMaxAttempts; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}
		if err := Validate(candidate); err != nil {
			continue
		}
		ok, err := check(ctx, candidate)
		if err != nil {
			return "", err
		}
		if ok {
			return candidate, nil
		}
	}
	return "", ErrCollisionExhausted
}
