package routers

import (
	controllers "github.com/kube-carbonara/cluster-agent/controllers"
	"github.com/kube-carbonara/cluster-agent/utils"
	"github.com/labstack/echo/v4"
)

type NodesRouter struct{}

func (router NodesRouter) Handle(e *echo.Echo) {
	nodesController := controllers.NodesController{}
	e.GET("/nodes", func(context echo.Context) error {
		return nodesController.Get(context)
	})

	e.GET("/nodesmetrics", func(context echo.Context) error {
		return nodesController.Metrics(context)
	})

	e.GET("/nodes/:id", func(context echo.Context) error {
		return nodesController.GetOne(context, context.Param("id"))
	})

	e.POST("/nodes/:id", func(context echo.Context) error {
		node := utils.JsonBodyToMap(context.Request().Body)
		return nodesController.Create(context, node)
	})

	e.PUT("/nodes/:id", func(context echo.Context) error {
		node := utils.JsonBodyToMap(context.Request().Body)

		return nodesController.Update(context, node)
	})

	e.DELETE("/nodes/:id", func(context echo.Context) error {
		return nodesController.Delete(context, context.Param("id"))
	})
}
