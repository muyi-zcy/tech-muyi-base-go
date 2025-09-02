package core

import (
	"github.com/gin-gonic/gin"
)

type HealthCheckController struct{}

func NewHealthCheckController() *HealthCheckController {
	return &HealthCheckController{}
}

func (h *HealthCheckController) Ok(c *gin.Context) {
	c.String(200, "ok")
}

func RegisterHealthCheckRoutes(engine *gin.Engine, controller *HealthCheckController) {
	engine.GET("/ok", controller.Ok)
}
