package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Book struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Author     string    `json:"author"`
	ISBN       string    `json:"isbn"`
	Year       int       `json:"year"`
	Price      float64   `json:"price"`
	Created_At time.Time `json:"created_at"`
	Updated_At time.Time `json:"updated_at"`
}

// -------------------- Database Connection --------------------

func initDB() {
	var err error
	host := getEnv("DB_HOST", "localhost")
	name := getEnv("DB_NAME", "bookstore")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	port := getEnv("DB_PORT", "5432")

	conStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, name,
	)

	db, err = sql.Open("postgres", conStr)
	if err != nil {
		log.Fatalf("❌ Failed to open database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)

	for i := 1; i <= 10; i++ {
		err = db.Ping()
		if err == nil {
			log.Println("✅ Successfully connected to database!")
			return
		}
		log.Printf("⚠️  Failed to ping database (attempt %d/10): %v", i, err)
		time.Sleep(3 * time.Second)
	}

	log.Fatalf("❌ Failed to connect to database after retries: %v", err)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// -------------------- API Handlers --------------------

func getHealth(c *gin.Context) {
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "Unhealthy", "error": "Database not initialized"})
		return
	}
	if err := db.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "Unhealthy", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "healthy"})
}

func getAllBooks(c *gin.Context) {
	rows, err := db.Query(`SELECT id, title, author, isbn, year, price, created_at, updated_at FROM books`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price, &book.Created_At, &book.Updated_At); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books = append(books, book)
	}
	if books == nil {
		books = []Book{}
	}
	c.JSON(http.StatusOK, books)
}

func getBook(c *gin.Context) {
	id := c.Param("id")
	var book Book
	err := db.QueryRow(
		`SELECT id, title, author, isbn, year, price, created_at, updated_at FROM books WHERE id = $1`, id,
	).Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price, &book.Created_At, &book.Updated_At)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, book)
}

// ✅ เพิ่มฟังก์ชันใหม่: ดึงหนังสือใหม่ (4 เล่มล่าสุด)
func getNewBook(c *gin.Context) {
	rows, err := db.Query(`SELECT id, title, author, isbn, year, price, created_at, updated_at FROM books ORDER BY created_at DESC LIMIT 4`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Year, &b.Price, &b.Created_At, &b.Updated_At); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books = append(books, b)
	}

	c.JSON(http.StatusOK, books)
}

func createBook(c *gin.Context) {
	var newBook Book
	if err := c.ShouldBindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	var createdAt, updatedAt time.Time
	err := db.QueryRow(
		`INSERT INTO books (title, author, isbn, year, price)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, created_at, updated_at`,
		newBook.Title, newBook.Author, newBook.ISBN, newBook.Year, newBook.Price,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newBook.ID = id
	newBook.Created_At = createdAt
	newBook.Updated_At = updatedAt

	c.JSON(http.StatusCreated, newBook)
}

func updateBook(c *gin.Context) {
	id := c.Param("id")
	var updated Book
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updatedAt time.Time
	err := db.QueryRow(
		`UPDATE books
         SET title=$1, author=$2, isbn=$3, year=$4, price=$5, updated_at=NOW()
         WHERE id=$6 RETURNING updated_at`,
		updated.Title, updated.Author, updated.ISBN, updated.Year, updated.Price, id,
	).Scan(&updatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated.Updated_At = updatedAt
	c.JSON(http.StatusOK, updated)
}

func deleteBook(c *gin.Context) {
	id := c.Param("id")
	result, err := db.Exec(`DELETE FROM books WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "book deleted successfully"})
}

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()

	// ✅ เปิดใช้ CORS (สำหรับ React ที่ port 3001)
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true // อนุญาตทุก origin (ใช้เฉพาะตอนพัฒนา)
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Routes
	r.GET("/health", getHealth)

	api := r.Group("/api/v1")
	{
		api.GET("/books", getAllBooks)
		api.GET("/books/:id", getBook)
		api.GET("/books/new", getNewBook) 
		api.POST("/books", createBook)
		api.PUT("/books/:id", updateBook)
		api.DELETE("/books/:id", deleteBook)
	}

	log.Println("🚀 Server starting on :8080")
	r.Run(":8080")
}