package v1

import (
	"net/http"

	"github.com/anthropics/agentsmesh/backend/internal/middleware"
	podgitservice "github.com/anthropics/agentsmesh/backend/internal/service/podgit"
	"github.com/anthropics/agentsmesh/backend/pkg/apierr"
	"github.com/gin-gonic/gin"
)

type PodGitHandler struct {
	service *podgitservice.Service
}

func NewPodGitHandler(service *podgitservice.Service) *PodGitHandler {
	return &PodGitHandler{service: service}
}

type podGitDiffQuery struct {
	Path     string `form:"path"`
	Staged   bool   `form:"staged"`
	Context  int32  `form:"context"`
	MaxBytes int32  `form:"max_bytes"`
}

type podGitCommitBody struct {
	Message string   `json:"message" binding:"required"`
	Paths   []string `json:"paths"`
	All     bool     `json:"all"`
	Author  *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
}

type podGitPushBody struct {
	Branch         string `json:"branch" binding:"required"`
	RemoteURL      string `json:"remote_url" binding:"required"`
	SetUpstream    bool   `json:"set_upstream"`
	ForceWithLease bool   `json:"force_with_lease"`
	Auth           struct {
		Username string `json:"username" binding:"required"`
		Token    string `json:"token" binding:"required"`
	} `json:"auth" binding:"required"`
}

func (h *PodGitHandler) Status(c *gin.Context) {
	tenant := middleware.GetTenant(c)
	resp, err := h.service.Status(c.Request.Context(), tenant.OrganizationID, c.Param("key"))
	if err != nil {
		respondPodGitError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PodGitHandler) Diff(c *gin.Context) {
	var query podGitDiffQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}

	tenant := middleware.GetTenant(c)
	resp, err := h.service.Diff(c.Request.Context(), tenant.OrganizationID, c.Param("key"), podgitservice.DiffRequest{
		Path:     query.Path,
		Staged:   query.Staged,
		Context:  query.Context,
		MaxBytes: query.MaxBytes,
	})
	if err != nil {
		respondPodGitError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PodGitHandler) Commit(c *gin.Context) {
	var body podGitCommitBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}

	var author *podgitservice.CommitAuthor
	if body.Author != nil {
		author = &podgitservice.CommitAuthor{
			Name:  body.Author.Name,
			Email: body.Author.Email,
		}
	}

	tenant := middleware.GetTenant(c)
	resp, err := h.service.Commit(c.Request.Context(), tenant.OrganizationID, c.Param("key"), podgitservice.CommitRequest{
		Message: body.Message,
		Paths:   body.Paths,
		All:     body.All,
		Author:  author,
	})
	if err != nil {
		respondPodGitError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PodGitHandler) Push(c *gin.Context) {
	var body podGitPushBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.ValidationError(c, err.Error())
		return
	}

	tenant := middleware.GetTenant(c)
	resp, err := h.service.Push(c.Request.Context(), tenant.OrganizationID, c.Param("key"), podgitservice.PushRequest{
		Branch:         body.Branch,
		RemoteURL:      body.RemoteURL,
		SetUpstream:    body.SetUpstream,
		ForceWithLease: body.ForceWithLease,
		Auth: podgitservice.PushAuth{
			Username: body.Auth.Username,
			Token:    body.Auth.Token,
		},
	})
	if err != nil {
		respondPodGitError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func respondPodGitError(c *gin.Context, err error) {
	if apiErr, ok := err.(*podgitservice.Error); ok {
		apierr.Respond(c, apiErr.HTTPStatus, apiErr.Code, apiErr.Message)
		return
	}
	apierr.InternalError(c, "Failed to execute pod git command")
}
