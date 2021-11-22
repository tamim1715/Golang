package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ACTION string
const (
	UPDATE_ADDRESS = ACTION("UPDATE_ADDRESS")
	UPDATE_NAME = ACTION("UPDATE_NAME")
)

type Customer struct{

	ID string `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
	Address string `json:"address" bson:"address"`
}
var mpDTO = make(map[string]Customer)

func main(){


	e := echo.New()

	data := Customer{}

	e.POST("/api/v1/customers", data.Save)
	e.GET("/api/v1/customers:id", data.FindByID)
	e.PUT("/api/v1/customers:id", data.Update)
	e.PATCH("/api/v1/customers:id", data.Patch)
	e.DELETE("/api/v1/customers:id", data.Delete)

	e.Logger.Fatal(e.Start(":8000"))

}

func  (data Customer) Save(c echo.Context) error{

	body := &Customer{}
	if err := c.Bind(&body); err !=nil{
		log.Fatal(err)
	}
	mpDTO[body.ID]=*body

	return c.JSON(http.StatusOK, "Save Information")
}

func (data Customer) FindByID(c echo.Context) error {
	id :=c.Param("id")
	var value Customer

	value = mpDTO[id]

	return c.JSON(http.StatusOK, value)
}

func (data Customer) Update(c echo.Context) error{

	body := &Customer{}
	if err := c.Bind(&body); err !=nil{
		log.Fatal(err)
	}
	ID := c.Param("id")
	body.ID=ID
	mpDTO[ID]=*body

	return c.JSON(http.StatusOK, "update is complete")
}

func (data Customer) Patch(c echo.Context) error{
	
	body := &Customer{}
	if err := c.Bind(&body); err !=nil{
		log.Fatal(err)
	}

	ID := c.Param("id")
	var value Customer
	value=mpDTO[ID]
	

	action := c.QueryParam("action")

	if action ==string(UPDATE_ADDRESS){
		value.Address=body.Address
	}
	if action ==string(UPDATE_NAME){
		value.Name=body.Name
	}
	mpDTO[ID]=value

	return c.JSON(http.StatusOK, "Operation Successfull")
}


func (data Customer) Delete(c echo.Context) error{

	ID := c.Param("id")

	delete(mpDTO, ID)

	return c.JSON(http.StatusOK, "Delete your record")

}