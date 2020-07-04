package routes

import (
	"app/application/controllers"
	"github.com/labstack/echo"
)

func Init(e *echo.Echo) {
	g := e.Group("/api")
	{
		g.GET("/chart", controllers.GetAllCandle())
	}
}
