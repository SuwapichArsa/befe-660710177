package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Menu struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type Order struct {
	ID         int     `json:"id"`
	Items      []int   `json:"items"`
	TotalPrice float64 `json:"total_price"`
}

var menus = []Menu{
	{ID: 1, Name: "ไข่เจียวปู", Price: 800},
	{ID: 2, Name: "ราดหน้าทะเล", Price: 400},
	{ID: 3, Name: "ผัดขีเมาทะเล", Price: 400},
	{ID: 4, Name: "ต้มยำกุ้ง", Price: 600},
	{ID: 5, Name: "ต้มยำทะเล", Price: 600},
}

var orders []Order

func homeHandler(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to the Restaurant Jay Fai")
}

func getMenus(c *gin.Context) {
	c.JSON(http.StatusOK, menus)
}

func addMenu(c *gin.Context) {
	var newMenu Menu
	if err := c.ShouldBindJSON(&newMenu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	newMenu.ID = len(menus) + 1
	menus = append(menus, newMenu)
	c.JSON(http.StatusOK, newMenu)
}

func getOrders(c *gin.Context) {
	c.JSON(http.StatusOK, orders)
}

func createOrder(c *gin.Context) {
	var newOrder Order
	if err := c.ShouldBindJSON(&newOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	total := 0.0
	for _, itemID := range newOrder.Items {
		found := false
		for _, menu := range menus {
			if menu.ID == itemID {
				total += menu.Price
				found = true
				break
			}
		}
		if !found {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Menu ID %d not found", itemID)})
			return
		}
	}

	newOrder.ID = len(orders) + 1
	newOrder.TotalPrice = total
	orders = append(orders, newOrder)
	c.JSON(http.StatusOK, newOrder)
}

func main() {
	r := gin.Default()

	r.GET("/", homeHandler)

	api := r.Group("/api/v1")
	{
		api.GET("/menus", getMenus)
		api.POST("/menus", addMenu)

		api.GET("/orders", getOrders)
		api.POST("/orders", createOrder)
	}

	fmt.Println("🚀 Server is running at http://localhost:8080")
	r.Run(":8080")
}
