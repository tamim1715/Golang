package main

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
var collection *mongo.Collection
var ctx = context.TODO()

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

func main(){


	e := echo.New()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
	collection = client.Database("local").Collection("Testing")

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

	_, err := collection.InsertOne(ctx, *body)
	if err != nil{
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, "Save Information")
}

func (data Customer) FindByID(c echo.Context) error {
	id :=c.Param("id")
	var value Customer
	err :=collection.FindOne(ctx, bson.D{{"id", id}}).Decode(&value)
	if err != nil{
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, value)
}

func (data Customer) Update(c echo.Context) error{

	body := &Customer{}
	if err := c.Bind(&body); err !=nil{
		log.Fatal(err)
	}
	ID := c.Param("id")
	body.ID=ID

    filter :=bson.D{{"id",ID}}
	update :=bson.D{{"$set", *body}}

	_, err :=collection.UpdateOne(
		ctx, 
		filter, 
		update,
	)
	if err != nil{
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, "update is complete")
}

func (data Customer) Patch(c echo.Context) error{
	
	body := &Customer{}
	if err := c.Bind(&body); err !=nil{
		log.Fatal(err)
	}

	ID := c.Param("id")
	var value Customer
	err :=collection.FindOne(ctx, bson.D{{"id", ID}}).Decode(&value)
	if err != nil{
		log.Fatal(err)
	}

	action := c.QueryParam("action")

	if action ==string(UPDATE_ADDRESS){
		value.Address=body.Address
	}
	if action ==string(UPDATE_NAME){
		value.Name=body.Name
	}
	

    filter :=bson.D{{"id",ID}}
	update :=bson.D{{"$set", value}}

	_, err =collection.UpdateOne(
		ctx, 
		filter, 
		update,
	)
	if err != nil{
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, "Operation Successfull")
}


func (data Customer) Delete(c echo.Context) error{

	ID := c.Param("id")

	_, err :=collection.DeleteOne(ctx, bson.D{{"id", ID}})
	if err !=nil{
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, "Delete your record")

}