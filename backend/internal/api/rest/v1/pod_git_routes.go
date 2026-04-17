package v1

import "github.com/gin-gonic/gin"

func registerPodGitRoutes(pods *gin.RouterGroup, handler *PodGitHandler) {
	if handler == nil {
		return
	}

	pods.GET("/:key/git/status", handler.Status)
	pods.GET("/:key/git/diff", handler.Diff)
	pods.POST("/:key/git/commit", handler.Commit)
	pods.POST("/:key/git/push", handler.Push)
}
