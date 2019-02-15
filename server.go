package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Customer struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func getConnection() *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("can't connect to database : ", err)
	}
	return db
}

func getCustomerHandler(c *gin.Context) {
	db := getConnection()
	defer db.Close()

	stmt, err := db.Prepare("select id,name,email,status from customers order by id,name asc")
	if err != nil {
		c.JSON(500, gin.H{
			"errorCode": 500,
			"errorDesc": err,
		})
	}

	rows, err := stmt.Query()
	if err != nil {
		c.JSON(500, gin.H{
			"errorCode": 500,
			"errorDesc": err,
		})
	}

	customers := []Customer{}
	for rows.Next() {
		customer := Customer{}
		rows.Scan(&customer.Id, &customer.Name, &customer.Email, &customer.Status)
		customers = append(customers, customer)
	}

	fmt.Println("query all customers success")
	c.JSON(200, customers)
}
func getCustomerByIdHandler(c *gin.Context) {
	db := getConnection()
	defer db.Close()

	stmt, err := db.Prepare("select id,name,email,status from customers where id=$1")
	if err != nil {
		c.JSON(500, gin.H{
			"errorCode": 500,
			"errorDesc": err.Error(),
		})
		return
	}

	rowId := c.Param("id")
	row := stmt.QueryRow(rowId)
	var customer = Customer{}

	err = row.Scan(&customer.Id, &customer.Name, &customer.Email, &customer.Status)

	if err != nil {
		c.JSON(500, gin.H{
			"errorCode": 500,
			"errorDesc": err.Error(),
		})
		return
	}
	fmt.Println("one row ", customer.Id, customer.Name, customer.Email, customer.Status)

	c.JSON(200, customer)
}
func updateCustomerHandler(c *gin.Context) {
	db := getConnection()
	defer db.Close()

	stmt, err := db.Prepare("update customers set name=$2, email=$3, status=$4 where id=$1;")

	if err != nil {
		c.JSON(500, gin.H{
			"errorCode": 500,
			"errorDesc": err.Error(),
		})
		return
	}

	rowId := c.Param("id")
	var json Customer
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := stmt.Exec(rowId, json.Name, json.Email, json.Status); err != nil {
		c.JSON(500, gin.H{
			"errorCode": 500,
			"errorDesc": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"id":     rowId,
		"name":   json.Name,
		"email":  json.Email,
		"status": json.Status,
	})
}
func deleteCustomerHandler(c *gin.Context) {
	db := getConnection()
	defer db.Close()

	stmt, err := db.Prepare("delete from customers where id=$1;")

	if err != nil {
		c.JSON(500, gin.H{
			"errorCode": 500,
			"errorDesc": err.Error(),
		})
		return
	}

	rowId := c.Param("id")
	if _, err := stmt.Exec(rowId); err != nil {
		c.JSON(500, gin.H{
			"errorCode": 500,
			"errorDesc": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"id":      rowId,
		"message": "customer deleted",
		"status":  "success",
	})
}
func createCustomerHandler(c *gin.Context) {
	db := getConnection()
	defer db.Close()

	var json Customer
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := db.QueryRow("insert into customers(name,email,status) values($1,$2,$3) returning id", json.Name, json.Email, json.Status)
	var id int
	err := row.Scan(&id)

	if err != nil {
		c.JSON(500, gin.H{
			"errorCode": 500,
			"errorDesc": err.Error(),
		})
		return
	}

	c.JSON(201, gin.H{
		"id":     id,
		"name":   json.Name,
		"email":  json.Email,
		"status": json.Status,
	})
}

func init() {
	db := getConnection()
	defer db.Close()

	createTb := `
	create table if not exists customers(
		id serial primary key,
		name text,
		email text,
		status text
	);
	`

	_, err := db.Exec(createTb)
	if err != nil {
		log.Fatal("can't create table", err)
	}
	fmt.Println("create table success")
}

func authMiddleware(c *gin.Context) {
	log.Println("start middleware")
	authKey := c.GetHeader("Authorization")
	if authKey != "token2019" {
		c.JSON(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		c.Abort()
		return
	}

	c.Next()
	log.Println("end middleware")
}

func main() {
	r := gin.Default()
	r.Use(authMiddleware)
	r.GET("customers", getCustomerHandler)
	r.GET("customers/:id", getCustomerByIdHandler)
	r.PUT("customers/:id", updateCustomerHandler)
	r.DELETE("customers/:id", deleteCustomerHandler)
	r.POST("customers", createCustomerHandler)
	r.Run(":2019")
}
