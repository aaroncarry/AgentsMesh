package podgit

import runnerv1 "github.com/anthropics/agentsmesh/proto/gen/go/runner/v1"

type DiffRequest struct {
	Path     string
	Staged   bool
	Context  int32
	MaxBytes int32
}

type CommitAuthor struct {
	Name  string
	Email string
}

type CommitRequest struct {
	Message string
	Paths   []string
	All     bool
	Author  *CommitAuthor
}

type PushAuth struct {
	Username string
	Token    string
}

type PushRequest struct {
	Branch         string
	RemoteURL      string
	SetUpstream    bool
	ForceWithLease bool
	Auth           PushAuth
}

type StatusResponse struct {
	Ok               bool                      `json:"ok"`
	PodKey           string                    `json:"pod_key"`
	Branch           string                    `json:"branch"`
	HeadSHA          string                    `json:"head_sha"`
	HasChanges       bool                      `json:"has_changes"`
	HasStagedChanges bool                      `json:"has_staged_changes"`
	Files            []*runnerv1.GitStatusFile `json:"files"`
	Stats            *runnerv1.GitStatusStats  `json:"stats"`
}

type DiffResponse struct {
	Ok        bool   `json:"ok"`
	PodKey    string `json:"pod_key"`
	Branch    string `json:"branch"`
	Path      string `json:"path"`
	Staged    bool   `json:"staged"`
	Truncated bool   `json:"truncated"`
	MaxBytes  int32  `json:"max_bytes"`
	Diff      string `json:"diff"`
}

type CommitResponse struct {
	Ok             bool     `json:"ok"`
	PodKey         string   `json:"pod_key"`
	Branch         string   `json:"branch"`
	CommitSHA      string   `json:"commit_sha"`
	Message        string   `json:"message"`
	CommittedFiles []string `json:"committed_files"`
}

type PushResponse struct {
	Ok            bool   `json:"ok"`
	PodKey        string `json:"pod_key"`
	Branch        string `json:"branch"`
	RemoteURL     string `json:"remote_url"`
	Pushed        bool   `json:"pushed"`
	UpstreamSet   bool   `json:"upstream_set"`
	RemoteHeadSHA string `json:"remote_head_sha"`
}
