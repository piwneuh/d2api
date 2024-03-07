package api

import (
	"context"
	v1 "d2api/pkg/server/api/v1"

	"github.com/gin-gonic/gin"
)

func RegisterVersion(router *gin.Engine, ctx context.Context) {
	v1.RegisterServer(router, ctx)
}
