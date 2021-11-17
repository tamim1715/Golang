package main

import (
	value "myapp/Customer"

	"github.com/labstack/echo/v4"
)



func main() {
	e := echo.New()

	value.Init()

	data := value.Customer{}

	e.POST("/api/v1/customers", data.Save)
	e.GET("/api/v1/customers:id", data.FindByID)
	e.PATCH("/api/v1/customers:id", data.Patch)
	e.PUT("/api/v1/customers:id", data.Update)
	e.DELETE("/api/v1/customers:id", data.Delete)

	e.Logger.Fatal(e.Start(":8000"))
}
