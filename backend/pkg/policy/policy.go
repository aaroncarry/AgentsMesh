package policy

// Subject is the acting user.
type Subject struct {
	OrgID  int64
	UserID int64
	Role   string // "owner", "admin", "member", "apikey"
}

// NewSubject constructs a Subject from raw tenant context fields.
func NewSubject(orgID, userID int64, role string) Subject {
	return Subject{OrgID: orgID, UserID: userID, Role: role}
}

func (s Subject) isAdmin() bool  { return s.Role == "owner" || s.Role == "admin" }
func (s Subject) isAPIKey() bool { return s.Role == "apikey" }

// ResourceContext carries the access-relevant fields of a resource instance.
type ResourceContext struct {
	OrgID          int64
	OwnerID        int64
	Visibility     string
	GrantedUserIDs []int64
}

// PodResource builds a context for a ReadOwnerOnly resource (Pod).
func PodResource(orgID, createdByID int64) ResourceContext {
	return ResourceContext{OrgID: orgID, OwnerID: createdByID}
}

// VisibleResource builds a context for a ReadVisibility resource (Runner, Repository).
// Handles the nullable *int64 owner field internally.
func VisibleResource(orgID int64, ownerIDPtr *int64, visibility string) ResourceContext {
	var ownerID int64
	if ownerIDPtr != nil {
		ownerID = *ownerIDPtr
	}
	return ResourceContext{OrgID: orgID, OwnerID: ownerID, Visibility: visibility}
}

// WithGrants returns a copy with explicit grants attached (Phase 2).
func (rc ResourceContext) WithGrants(userIDs []int64) ResourceContext {
	rc.GrantedUserIDs = userIDs
	return rc
}

func (rc ResourceContext) isGranted(userID int64) bool {
	for _, id := range rc.GrantedUserIDs {
		if id == userID {
			return true
		}
	}
	return false
}

// Visibility constants for ReadVisibility policy mode.
const (
	VisibilityOrganization = "organization"
	VisibilityPrivate      = "private"
)
type ReadAccess int

const (
	ReadOrgOpen    ReadAccess = iota // any org member
	ReadOwnerOnly                    // members see own; admins/apikeys see all
	ReadVisibility                   // Visibility field controls; no admin bypass
)

// WriteAccess enumerates write access modes.
type WriteAccess int

const (
	WriteOrgOpen      WriteAccess = iota // any org member
	WriteCreatorAdmin                    // creator or admin
	WriteAdminOnly                       // admin only
)

// ResourcePolicy declares access rules for a resource type.
type ResourcePolicy struct {
	Read  ReadAccess
	Write WriteAccess
}

// AllowRead returns true if subject may read the resource.
// Grants augment within each mode (not as a universal bypass).
func (p ResourcePolicy) AllowRead(s Subject, res ResourceContext) bool {
	if res.OrgID != s.OrgID {
		return false
	}
	switch p.Read {
	case ReadOrgOpen:
		return true
	case ReadOwnerOnly:
		return s.isAdmin() || s.isAPIKey() || res.OwnerID == s.UserID || res.isGranted(s.UserID)
	case ReadVisibility:
		if res.Visibility == VisibilityPrivate {
			return res.OwnerID == s.UserID || res.isGranted(s.UserID)
		}
		return true
	}
	return false
}

// AllowWrite returns true if subject may mutate the resource.
func (p ResourcePolicy) AllowWrite(s Subject, res ResourceContext) bool {
	if res.OrgID != s.OrgID {
		return false
	}
	switch p.Write {
	case WriteOrgOpen:
		return true
	case WriteCreatorAdmin:
		return s.isAdmin() || res.OwnerID == s.UserID || res.isGranted(s.UserID)
	case WriteAdminOnly:
		return s.isAdmin()
	}
	return false
}

// AllowAdmin returns true if the subject is an admin in the given org.
// Use for pre-fetch admin-gate checks (Create operations, or as optimization
// before fetching a resource for WriteAdminOnly policies).
func AllowAdmin(s Subject, orgID int64) bool {
	return s.OrgID == orgID && s.isAdmin()
}

// ListFilter describes how to narrow a list query for the subject.
type ListFilter struct {
	// OwnerOnly: non-zero restricts results to this owner (ReadOwnerOnly, non-admin).
	OwnerOnly int64
	// VisibilityUserID: non-zero means apply visibility filtering for this user.
	// ReadVisibility always sets this (no admin bypass for truly private resources).
	VisibilityUserID int64
	// GrantUserID: non-zero means also include resources explicitly granted to this user.
	// Always set to the subject's UserID so repos can include grant-based access.
	GrantUserID int64
}

// ListFilter returns filter parameters for list queries under this policy.
func (p ResourcePolicy) ListFilter(s Subject) ListFilter {
	switch p.Read {
	case ReadOwnerOnly:
		if !s.isAdmin() && !s.isAPIKey() {
			return ListFilter{OwnerOnly: s.UserID, GrantUserID: s.UserID}
		}
		return ListFilter{}
	case ReadVisibility:
		return ListFilter{VisibilityUserID: s.UserID, GrantUserID: s.UserID}
	}
	return ListFilter{}
}
