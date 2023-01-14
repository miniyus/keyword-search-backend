package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/swagger"
	"github.com/miniyus/keyword-search-backend/internal/core/container"
	"github.com/miniyus/keyword-search-backend/internal/core/router"
	"github.com/miniyus/keyword-search-backend/internal/utils"
)

func External(c container.Container) {
	extRouter := router.New(c.App(), "/", "external")
	extRouter.Route("/", func(router fiber.Router) {
		router.Get("/metrics", monitor.New(monitor.Config{Title: "MyService Metrics Page"}))
		router.Get("/swagger/*", swagger.HandlerDefault)
		router.Get("/health-check", healthCheck)
	})

}

// healthCheck
// @Summary health check your server
// @Description health check your server
// @Success 200 {object} utils.StatusResponse
// @Tags HealthCheck
// @Accept */*
// @Produce json
// @Router /health-check [get]
func healthCheck(ctx *fiber.Ctx) error {

	err := ctx.JSON(utils.StatusResponse{Status: true})
	if err != nil {
		return ctx.JSON(utils.StatusResponse{Status: false})
	}

	return err
}