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

type Employee struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

var db *gorm.DB

func main() {
	// Replace these details with your MySQL server information
	dsn := "root:password@tcp(localhost:3306)/empdb"
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

	//Different Handlers
	r.GET("/employees", getEmployeesHandler)
	r.GET("/employees/:id", getEmployeeHandler)
	r.POST("/employees", createEmployeeHandler)
	r.PUT("/employees/:id", updateEmployeeHandler)

	r.Run(":8080")
}

func initDB() error {
	// Create database if not exist
	if err := db.Exec("CREATE DATABASE IF NOT EXISTS empdb").Error; err != nil {
		return err
	}

	// Use the database
	db.Exec("USE empdb")

	// Auto Migrate the schema
	if err := db.AutoMigrate(&Employee{}); err != nil {
		return err
	}

	return nil
}

func getEmployeesHandler(c *gin.Context) {
	var employees []Employee
	if err := db.Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, employees)
}

func getEmployeeHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	var employee Employee
	if err := db.First(&employee, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, employee)
}

func createEmployeeHandler(c *gin.Context) {
	var employee Employee
	if err := c.ShouldBindJSON(&employee); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	if err := db.Create(&employee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": employee.ID})
}

func updateEmployeeHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	var existingEmployee Employee
	if err := db.First(&existingEmployee, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	var updatedEmployee Employee
	if err := c.ShouldBindJSON(&updatedEmployee); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	if err := db.Model(&existingEmployee).Updates(updatedEmployee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Employee updated successfully"})
}
