// go run main.go -dbhostname=localhost -dbname=custdb -dbuser=root -dbpass=password -dbport=3306

package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	dbHostname string
	dbName     string
	dbUser     string
	dbPass     string
	dbPort     int
)

func init() {
	flag.StringVar(&dbHostname, "dbhostname", "localhost", "MySQL database hostname")
	flag.StringVar(&dbName, "dbname", "custdb", "MySQL database name")
	flag.StringVar(&dbUser, "dbuser", "root", "MySQL database user")
	flag.StringVar(&dbPass, "dbpass", "password", "MySQL database password")
	flag.IntVar(&dbPort, "dbport", 3306, "MySQL database port")
	flag.Parse()
}

type Customer struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

var db *gorm.DB

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", dbUser, dbPass, dbHostname, dbPort)
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}

	// Create the database if it does not exist
	if err := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName).Error; err != nil {
		fmt.Println("Error creating database:", err)
		return
	}

	dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbUser, dbPass, dbHostname, dbPort, dbName)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}

	if err := initDB(); err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	r.Use(cors.New(config))

	r.GET("/customers", getCustomersHandler)
	r.GET("/customers/:id", getCustomerHandler)
	r.POST("/customers", createCustomerHandler)
	r.PUT("/customers/:id", updateCustomerHandler)

	r.Run(":8080")
}

func initDB() error {
	db.Exec("USE " + dbName)

	if err := db.AutoMigrate(&Customer{}); err != nil {
		return err
	}

	return nil
}

func getCustomersHandler(c *gin.Context) {
	var customers []Customer
	if err := db.Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, customers)
}

func getCustomerHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var customer Customer
	if err := db.First(&customer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func createCustomerHandler(c *gin.Context) {
	var customer Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	if err := db.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": customer.ID})
}

func updateCustomerHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var existingCustomer Customer
	if err := db.First(&existingCustomer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	var updatedCustomer Customer
	if err := c.ShouldBindJSON(&updatedCustomer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	if err := db.Model(&existingCustomer).Updates(updatedCustomer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer updated successfully"})
}
