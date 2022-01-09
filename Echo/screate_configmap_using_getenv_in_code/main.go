package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)
var collection *mongo.Collection
var collection1 *mongo.Collection
var ctx = context.TODO()


type ACTION string
const (
	UPDATE_ADDRESS = ACTION("UPDATE_ADDRESS")
	UPDATE_NAME = ACTION("UPDATE_NAME")
)

type Customer struct{

	CID string `json:"cid" bson:"cid"`
	CName string `json:"cname" bson:"cname"`
	CAddress string `json:"caddress" bson:"caddress"`
}

type CustomersProduct struct{
	CID string `json:"cid" bson:"cid"`
	PID string `json:"pid" bson:"pid"`
	Quantity int `json:"amount" bson:"amount"`
	
}

type Response struct {
	PID string `json:"pid" bson:"pid"`
	PName string `json:"pname" bson:"pname"`
	Quantity int `json:"quantity" bson:"quantity"`
	Price int `json:"price" bson:"price"`
}

type store struct {
	Id   interface{} `json:"_id" bson:"_id"`
	Count int `json:"count" bson:"count"`
}

func main(){


	e := echo.New()
	url := fmt.Sprintf("mongodb://%s:%s@%s:%s",os.Getenv("MONGO_INITDB_ROOT_USERNAME"), os.Getenv("MONGO_INITDB_ROOT_PASSWORD"), os.Getenv("MONGO_SERVICE"), os.Getenv("MONGO_PORT"))
	log.Println("\n\nhello bro= ",url)
	clientOptions := options.Client().ApplyURI(url)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
	collection = client.Database("User").Collection("Customer")
	collection1 = client.Database("User").Collection("CustomerProduct")

	data := Customer{}
	cp := CustomersProduct{}

	e.POST("/api/v1/customers", data.Save)
	e.GET("/api/v1/customers/:id", data.FindByID)
	e.PUT("/api/v1/customers/:id", data.Update)
	e.PATCH("/api/v1/customers/:id", data.Patch)
	e.DELETE("/api/v1/customers/:id", data.Delete)
	e.POST("/api/v1/addCustomerProduct", cp.addCustomerProduct)

	e.Logger.Fatal(e.Start(":8081"))

}

func (val CustomersProduct) addCustomerProduct(c echo.Context) error{

	body := &CustomersProduct{}
	x := store{}
	var result []bson.M
	if err := c.Bind(&body); err !=nil{
		log.Fatal(err)
	}

	product, err := getProduct(body.PID)
	if err != nil {
		log.Println(fmt.Sprintf("[Error] get product from product service: %s", err.Error()))
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	productJson, _ := json.Marshal(product)
	fmt.Println("product info: " + string(productJson))

    pipeline := []bson.M{
        {
            "$match": bson.M{
                "pid": body.PID,
            },
        },
        {
            "$group": bson.M{
                "_id":                bson.M{"pid": "$pid"},
                "count":              bson.M{"$sum": "$amount"},
            },
        },
    }

    cur, err := collection1.Aggregate(ctx, pipeline)
    if err != nil {
        log.Println("[ERROR]", err)
    }
    _ = cur.All(ctx, &result)


	if result != nil && len(result) > 0 {
		bsonBytes, _ := bson.Marshal(result[0])
		bson.Unmarshal(bsonBytes, &x)
	
		if(product.Quantity-body.Quantity-x.Count<0){
			return c.JSON(http.StatusUnavailableForLegalReasons, "Not Availabe This Product")
		}
	}


	_, err1 := collection1.InsertOne(ctx, *body)
	if err1 != nil{
		log.Fatal(err)
	}
	
	return c.JSON(http.StatusOK, "wow! you buy this product")

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
	cid :=c.Param("id")
	var value Customer
	err :=collection.FindOne(ctx, bson.D{{"cid", cid}}).Decode(&value)
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
	CID := c.Param("id")
	body.CID=CID

    filter :=bson.D{{"cid",CID}}
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

	CID := c.Param("id")
	var value Customer
	err :=collection.FindOne(ctx, bson.D{{"cid", CID}}).Decode(&value)
	if err != nil{
		log.Fatal(err)
	}

	action := c.QueryParam("action")

	if action ==string(UPDATE_ADDRESS){
		value.CAddress=body.CAddress
	}
	if action ==string(UPDATE_NAME){
		value.CName=body.CName
	}
	

    filter :=bson.D{{"cid",CID}}
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

	CID := c.Param("id")

	_, err :=collection.DeleteOne(ctx, bson.D{{"cid", CID}})
	if err !=nil{
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, "Delete your record")

}

func getProduct(id string) (*Response, error) {
    client := &http.Client{}
    req, err := http.NewRequest(http.MethodGet, "http://productservice:8000/api/v1/products/"+ id, nil)
    if req != nil {
        req.Header.Add("Content-Type", "application/json")
    }
    resp, err := client.Do(req)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("error in connecting service: %s", err.Error()))
    }
    defer resp.Body.Close()
    if resp.Body != nil {
        jsonDataFromHttp, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            return nil, errors.New(fmt.Sprintf("error in getting response body: %s", err.Error()))
        }

		response := Response{}
		err = json.Unmarshal([]byte(jsonDataFromHttp), &response) // here!
        if err != nil {
            return nil, errors.New(fmt.Sprintf("error in parsing response body: %s", err.Error()))
        }
		// fmt.Println(jsonDataFromHttp)

		// if resp.StatusCode == http.StatusOK {
			
		// }

			return &response, nil
        
    } else {
        return nil, errors.New("something went wrong")
    }
}