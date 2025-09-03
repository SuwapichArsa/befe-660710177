package main

import (
	"fmt"
	"net/http"

	"slices"

	"github.com/gin-gonic/gin"
)

type Student struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Year  int     `json:"year"`
	GPA   float64 `json:"gpa"`
}

var students = []Student{
	{ID: "1", Name: "John Doe", Email: "john@example.com", Year: 3, GPA: 3.50},
	{ID: "2", Name: "Jane Smith", Email: "jane@example.com", Year: 2, GPA: 3.75},
}

// ดึงนักเรียนทั้งหมด หรือ filter ด้วย year
func getStudents(c *gin.Context) {
	yearQuery := c.Query("year")

	if yearQuery != "" {
		filter := []Student{}
		for _, student := range students {
			if fmt.Sprint(student.Year) == yearQuery {
				filter = append(filter, student)
			}
		}
		c.JSON(http.StatusOK, filter)
		return
	}
	c.JSON(http.StatusOK, students)
}

// ดึงนักเรียนตาม id
func getStudent(c *gin.Context) {
	id := c.Param("id")

	for _, student := range students {
		if student.ID == id {
			c.JSON(http.StatusOK, student)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
}

func createStudent(c *gin.Context) {
	var newStudent Student
	
	if err := c.ShouldBindJSON(&newStudent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error" : err.Error()})
		return
	}

	if newStudent.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error" : "Name is Required"})
		return
	}

	if newStudent.Year < 1 || newStudent.Year > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error" : "Year must be 1-4"})
		return
	}
	
	newStudent.ID = fmt.Sprintf("%d", len(students)+1)

	students = append(students, newStudent)
	c.JSON(http.StatusOK, students)
}

func updateStudent(c *gin.Context) {
	id := c.Param("id")
	var updateStudent Student

	if err := c.ShouldBindJSON(&updateStudent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error" : err.Error()})
		return
	} 

	for i, student := range students {
		if student.ID == id {
			updateStudent.ID = id
			students[i] = updateStudent
			c.JSON(http.StatusOK, updateStudent)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error" : "Student not found"})
}

func deleteStudent(c *gin.Context) {
	id := c.Param("id")

	for i, student := range students {
		if student.ID == id {
			students = slices.Delete(students, i, i+1)
			c.JSON(http.StatusOK, gin.H{"message" : "Student delete successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error" : "Student not found"})
}

func main() {
	r := gin.Default()

	// health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "healthy"})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/students", getStudents)
		api.GET("/students/:id", getStudent)
		api.POST("/student", createStudent)
		api.PUT("/students/:id", updateStudent)
		api.DELETE("/students/:id", deleteStudent)
	}

	// run ที่ port 8080
	r.Run(":8080")
}