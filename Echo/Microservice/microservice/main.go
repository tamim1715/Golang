package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Product struct {
	PID string `json:"pid" bson:"pid"`
	PName string `json:"pname" bson:"pname"`
	Quantity int `json:"quantity" bson:"quantity"`
	Price int `json:"price" bson:"price"`

	
}

var collection *mongo.Collection
var ctx = context.TODO()

func main(){
	e :=echo.New()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
	collection = client.Database("Product").Collection("allProduct")
	data := &Product{}
	e.POST("/api/v1/products", data.SaveProduct)
	e.GET("/api/v1/products/:id", data.FindByID)
	e.PUT("/api/v1/products/:id", data.Update)
	e.DELETE("/api/v1/products/:id", data.Delete)

	e.Logger.Fatal(e.Start(":8000"))
}


func (data *Product) add(c echo.Context) error {
	body := &Product{}
	var value Product
	if err := c.Bind(&body); err !=nil{
		log.Fatal(err)
	}

	PID := c.Param("id")
	PID = PID[1:]
	amount := c.Param("amount")
	amount = amount[1:]
	val, _ := strconv.Atoi(amount)
	// fmt.Println(PID, amount, val)
	
	err :=collection.FindOne(ctx, bson.D{{"pid", PID}}).Decode(&value)
	if err != nil{
		log.Fatal(err)
	}

	if value.Quantity-val < 0 {
		return c.JSON(http.StatusOK, "NOT Enough This Product")
	}
	value.Quantity -= val
    filter :=bson.D{{"pid",PID}}
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

func  (data *Product) SaveProduct(c echo.Context) error{

	if err := c.Bind(&data); err !=nil{
		log.Fatal(err)
	}
	_, err := collection.InsertOne(ctx, *data)
	if err != nil{
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, "Save Information")
}

func (data Product) FindByID(c echo.Context) error {
	PID :=c.Param("id")
	var value Product
	err :=collection.FindOne(ctx, bson.D{{"pid", PID}}).Decode(&value)
	// fmt.Println();
	if err != nil{
		log.Fatal(err)
	}
	return c.JSON(http.StatusOK, value)
}

func (data *Product) Update(c echo.Context) error{
	body := &Product{}
	if err := c.Bind(&body); err !=nil{
		log.Fatal(err)
	}
	PID := c.Param("id")
	body.PID=PID

    filter :=bson.D{{"pid",PID}}
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

func (data *Product) Delete(c echo.Context) error{

	PID := c.Param("id")

	_, err :=collection.DeleteOne(ctx, bson.D{{"pid", PID}})
	if err !=nil{
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, "Delete your record")

}