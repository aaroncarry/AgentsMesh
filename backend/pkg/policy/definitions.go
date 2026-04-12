package policy

var (
	// PodPolicy: members see/manage their own pods; admins see all.
	// Use PodResource(orgID, createdByID) to build ResourceContext.
	PodPolicy = ResourcePolicy{Read: ReadOwnerOnly, Write: WriteCreatorAdmin}

	// RunnerPolicy: visibility-controlled read (private = truly private, no admin bypass);
	// only admins may manage runners.
	// Use VisibleResource(orgID, registeredByUserID, visibility) to build ResourceContext.
	RunnerPolicy = ResourcePolicy{Read: ReadVisibility, Write: WriteAdminOnly}

	// RepositoryPolicy: visibility-controlled read; only admins may manage repos.
	// Use VisibleResource(orgID, importedByUserID, visibility) to build ResourceContext.
	RepositoryPolicy = ResourcePolicy{Read: ReadVisibility, Write: WriteAdminOnly}
)
