package Customer

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
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

	id string `json:"id"`
	name string `json:"Name"`
	address string `json:"Address"`
}

func Init(){
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
	collection = client.Database("local").Collection("Customer")

}


var mpDTO = make(map[string]Customer)


func  (data Customer) Save(c echo.Context) error{

	body := &Customer{}
	if err := c.Bind(&body); err !=nil{
		log.Fatal(err)
	}

	_, err := collection.InsertOne(ctx, *body)
	if err != nil{
		log.Fatal(err)
	}
	
	return c.JSON(http.StatusOK, "Save your value")
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

func (data Customer) Patch(c echo.Context) error{
	
	body := &Customer{}
	if err := c.Bind(&body); err !=nil{
		log.Fatal(err)
	}

	id := c.Param("id")
	var value Customer
	err :=collection.FindOne(ctx, bson.D{{"id", id}}).Decode(&value)
	if err != nil{
		log.Fatal(err)
	}

	action := c.QueryParam("action")

	if action ==string(UPDATE_ADDRESS){
		value.address=body.address
	}
	if action ==string(UPDATE_NAME){
		value.name=body.name
	}
	

    filter :=bson.D{{"id",id}}
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

func (data Customer) Update(c echo.Context) error{

	
	body := &Customer{}
	if err := c.Bind(&body); err !=nil{
		log.Fatal(err)
	}
	id := c.Param("id")
	body.id=id

    filter :=bson.D{{"id",id}}
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

func (data Customer) Delete(c echo.Context) error{

	id := c.Param("id")

	_, err :=collection.DeleteOne(ctx, bson.D{{"id", id}})
	if err !=nil{
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, "Delete your record")

}
