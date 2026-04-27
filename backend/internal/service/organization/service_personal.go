package organization

import (
	"context"
	"errors"
	"fmt"

	orgDomain "github.com/anthropics/agentsmesh/backend/internal/domain/organization"
	"github.com/anthropics/agentsmesh/backend/pkg/slugkit"
)

// CreatePersonal creates a personal workspace organization for a user.
//
// Slug derivation strategy (in order; first non-error wins):
//
//  1. rawSlug ("<sanitized-username>-workspace"): GenerateUnique walks raw,
//     raw-2, raw-3... until SlugExists returns false. The username is
//     pre-sanitized so Unicode-only inputs (e.g. "用户名") don't collapse to
//     a meaningless seed like "workspace".
//  2. rawSlug again: race-retry. If attempt 1 lost the SlugExists→Insert race
//     (DB unique-violation surfaced as ErrSlugAlreadyExists), this attempt
//     calls GenerateUnique with a fresh DB view, naturally advancing past
//     the now-taken candidates.
//  3. fallbackSlug ("user-<id>-workspace"): final guarantee path. Also used
//     when sanitized username is empty (Unicode-only / emoji-only input).
//
// Returns the last meaningful error if every attempt fails.
//
// Callers pass username and displayName so this method does not depend on
// the user service.
func (s *Service) CreatePersonal(ctx context.Context, ownerID int64, username, displayName string) (*orgDomain.Organization, error) {
	fallbackSlug := fmt.Sprintf("user-%d-workspace", ownerID)
	seeds := personalSlugSeeds(username, fallbackSlug)
	check := slugkit.FromExistsCheck(s.repo.SlugExists)
	name := personalWorkspaceName(displayName, username)

	var lastErr error
	for i, seed := range seeds {
		isFinal := i == len(seeds)-1

		finalSlug, err := slugkit.GenerateUnique(ctx, seed, check)
		if err != nil {
			lastErr = err
			if isFinal {
				return nil, err
			}
			continue
		}

		org, err := s.Create(ctx, ownerID, &CreateRequest{
			Name: name,
			Slug: finalSlug,
		})
		if err == nil {
			return org, nil
		}
		if !errors.Is(err, ErrSlugAlreadyExists) {
			return nil, err
		}
		lastErr = err
	}
	return nil, lastErr
}

// personalSlugSeeds returns the candidate seeds for personal workspace creation.
// When the username sanitizes to empty (Unicode-only / emoji-only input), all
// seeds use fallbackSlug — avoids every such user racing for the same
// degenerate "workspace" slug.
func personalSlugSeeds(username, fallbackSlug string) []string {
	sanitized := slugkit.Sanitize(username)
	if sanitized == "" {
		return []string{fallbackSlug}
	}
	rawSlug := fmt.Sprintf("%s-workspace", sanitized)
	return []string{rawSlug, rawSlug, fallbackSlug}
}

func personalWorkspaceName(displayName, username string) string {
	name := displayName
	if name == "" {
		name = username
	}
	return fmt.Sprintf("%s's Workspace", name)
}
