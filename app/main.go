package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Customer struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

var db *gorm.DB

func main() {
	// Replace these details with your MySQL server information
	dsn := "root:password@tcp(localhost:3306)/custdb"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}

	// Initialize database (create if not exist)
	if err := initDB(); err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	r := gin.Default()

	// Use CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	r.Use(cors.New(config))

	// Different Handlers for Customers
	r.GET("/customers", getCustomersHandler)
	r.GET("/customers/:id", getCustomerHandler)
	r.POST("/customers", createCustomerHandler)
	r.PUT("/customers/:id", updateCustomerHandler)

	r.Run(":8080")
}

func initDB() error {
	// Create database if not exist
	if err := db.Exec("CREATE DATABASE IF NOT EXISTS custdb").Error; err != nil {
		return err
	}

	// Use the database
	db.Exec("USE custdb")

	// Auto Migrate the schema
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
