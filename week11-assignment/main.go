package main

import (
	"database/sql"
	"fmt"
	_ "week11-assignment/docs"

	"log"
	"os"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-contrib/cors"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type Book struct {
    ID            int       `json:"id"`
    Title         string    `json:"title"`
    Author        string    `json:"author"`
    ISBN          string    `json:"isbn"`
    Year          int       `json:"year"`
    Price         float64   `json:"price"`

    // ฟิลด์ใหม่
    Category      string    `json:"category"`
    OriginalPrice *float64  `json:"original_price,omitempty"`
    Discount      int       `json:"discount"`
    CoverImage    string    `json:"cover_image"`
    Rating        float64   `json:"rating"`
    ReviewsCount  int       `json:"reviews_count"`
    IsNew         bool      `json:"is_new"`
    Pages         *int      `json:"pages,omitempty"`
    Language      string    `json:"language"`
    Publisher     string    `json:"publisher"`
    Description   string    `json:"description"`

    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var db *sql.DB

func initDB() {
	var err error

	host := getEnv("DB_HOST", "")
	name := getEnv("DB_NAME", "")
	user := getEnv("DB_USER", "")
	password := getEnv("DB_PASSWORD", "")
	port := getEnv("DB_PORT", "")

	conSt := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
	//fmt.Println(conSt)
	db, err = sql.Open("postgres", conSt)
	if err != nil {
		log.Fatal("failed to open")
	}
	// กำหนดจำนวน Connection สูงสุด
	db.SetMaxOpenConns(25)

	// กำหนดจำนวน Idle connection สูงสุด
	db.SetMaxIdleConns(20)

	// กำหนดอายุของ Connection
	db.SetConnMaxLifetime(5 * time.Minute)
	err = db.Ping()
	if err != nil {
		log.Fatal("failed to connect to database")
	}
	log.Println("successfully connected to database")
}

// @Summary Get all books with optional category filter
// @Description Get a list of all books, optionally filtered by category.
// @Tags Books
// @Produce  json
// @Param    category query     string  false "Filter books by category"
// @Success  200      {array}   Book
// @Failure  500      {object}  ErrorResponse
// @Router   /books [get]
func getAllBooks(c *gin.Context) {
    category := c.Query("category") // ดึงค่า query param ที่ชื่อ 'category'

    var rows *sql.Rows
    var err error

    // สร้าง query ตามเงื่อนไข
    if category != "" {
        // ถ้ามีการส่ง category มา ให้กรองข้อมูล
        query := "SELECT * FROM books WHERE category = $1 ORDER BY id ASC"
        rows, err = db.Query(query, category)
    } else {
        // ถ้าไม่ส่งมา ให้ดึงข้อมูลทั้งหมด
        query := "SELECT * FROM books ORDER BY id ASC"
        rows, err = db.Query(query)
    }

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    books, err := scanBooks(rows)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan books: " + err.Error()})
        return
    }
    
    if books == nil {
        books = []Book{}
    }
    c.JSON(http.StatusOK, books)
}

// Helper function to scan rows into a slice of Book
func scanBooks(rows *sql.Rows) ([]Book, error) {
    var books []Book
    for rows.Next() {
        var book Book
        // ต้อง Scan ให้ครบทุก field ที่เรา SELECT มา
        err := rows.Scan(
            &book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price,
            &book.Category, &book.OriginalPrice, &book.Discount, &book.CoverImage,
            &book.Rating, &book.ReviewsCount, &book.IsNew, &book.Pages,
            &book.Language, &book.Publisher, &book.Description,
            &book.CreatedAt, &book.UpdatedAt,
        )
        if err != nil {
            return nil, err
        }
        books = append(books, book)
    }
    return books, nil
}

// GET /api/v1/categories - ค้นหา category ทั้งหมด
// @Summary Get all unique categories
// @Description Get a list of unique book categories.
// @Tags Books
// @Produce  json
// @Success  200      {array}   string
// @Failure  500      {object}  ErrorResponse
// @Router   /categories [get]
func getCategories(c *gin.Context) {
    rows, err := db.Query("SELECT DISTINCT category FROM books WHERE category IS NOT NULL ORDER BY category ASC")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var categories []string
    for rows.Next() {
        var category string
        if err := rows.Scan(&category); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        categories = append(categories, category)
    }
    
    if categories == nil {
        categories = []string{}
    }
    c.JSON(http.StatusOK, categories)
}

// GET /api/v1/books/search?q=keyword - ค้นหา
// @Summary Search for books
// @Description Search for books by keyword in title and author.
// @Tags Books
// @Produce  json
// @Param    q    query     string  true  "Search keyword"
// @Success  200  {array}   Book
// @Failure  400  {object}  ErrorResponse
// @Failure  500  {object}  ErrorResponse
// @Router   /books/search [get]
func searchBooks(c *gin.Context) {
    keyword := c.Query("q")
    if keyword == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Search keyword 'q' is required"})
        return
    }

    // ใช้ ILIKE สำหรับ case-insensitive search และ % สำหรับ wildcard
    query := "SELECT * FROM books WHERE title ILIKE $1 OR author ILIKE $1 ORDER BY id ASC"
    searchTerm := "%" + keyword + "%"

    rows, err := db.Query(query, searchTerm)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    books, err := scanBooks(rows)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan books: " + err.Error()})
        return
    }

    if books == nil {
        books = []Book{}
    }
    c.JSON(http.StatusOK, books)
}

// GET /api/v1/books/featured - หนังสือแนะนำ
// @Summary Get featured books
// @Description Get a list of featured books (e.g., high rating).
// @Tags Books
// @Produce  json
// @Success  200  {array}   Book
// @Failure  500  {object}  ErrorResponse
// @Router   /books/featured [get]
func getFeaturedBooks(c *gin.Context) {
    // ตัวอย่าง: หนังสือแนะนำคือหนังสือที่ rating >= 4.5 และเรียงตาม rating
    query := "SELECT * FROM books WHERE rating >= 4.5 ORDER BY rating DESC, reviews_count DESC LIMIT 10"
    rows, err := db.Query(query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    books, err := scanBooks(rows)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan books: " + err.Error()})
        return
    }

    if books == nil {
        books = []Book{}
    }
    c.JSON(http.StatusOK, books)
}

// GET /api/v1/books/new - หนังสือใหม่
// @Summary Get new arrival books
// @Description Get a list of the newest books.
// @Tags Books
// @Produce  json
// @Success  200  {array}   Book
// @Failure  500  {object}  ErrorResponse
// @Router   /books/new [get]
func getNewBooks(c *gin.Context) {
    // หนังสือใหม่คือหนังสือที่เพิ่มเข้ามาล่าสุด
    query := "SELECT * FROM books ORDER BY created_at DESC LIMIT 10"
    rows, err := db.Query(query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    books, err := scanBooks(rows)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan books: " + err.Error()})
        return
    }

    if books == nil {
        books = []Book{}
    }
    c.JSON(http.StatusOK, books)
}

// GET /api/v1/books/discounted - หนังสือลดราคา
// @Summary Get discounted books
// @Description Get a list of books that are on sale.
// @Tags Books
// @Produce  json
// @Success  200  {array}   Book
// @Failure  500  {object}  ErrorResponse
// @Router   /books/discounted [get]
func getDiscountedBooks(c *gin.Context) {
    // หนังสือลดราคาคือเล่มที่มีส่วนลด > 0
    query := "SELECT * FROM books WHERE discount > 0 ORDER BY discount DESC"
    rows, err := db.Query(query)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    books, err := scanBooks(rows)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan books: " + err.Error()})
        return
    }
    
    if books == nil {
        books = []Book{}
    }
    c.JSON(http.StatusOK, books)
}

func getBook(c *gin.Context) {
	id := c.Param("id")
	var book Book

	// QueryRow ใช้เมื่อคาดว่าจะได้ผลลัพธ์ 0 หรือ 1 แถว
	err := db.QueryRow("SELECT id, title, author FROM books WHERE id = $1", id).
		Scan(&book.ID, &book.Title, &book.Author)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}

func createBook(c *gin.Context) {
	var newBook Book

	if err := c.ShouldBindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ใช้ RETURNING เพื่อดึงค่าที่ database generate (id, timestamps)
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
	newBook.CreatedAt = createdAt
	newBook.UpdatedAt = updatedAt

	c.JSON(http.StatusCreated, newBook) // ใช้ 201 Created
}

func updateBook(c *gin.Context) {
	var ID int
	id := c.Param("id")
	var updateBook Book

	if err := c.ShouldBindJSON(&updateBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updatedAt time.Time
	err := db.QueryRow(
		`UPDATE books
         SET title = $1, author = $2, isbn = $3, year = $4, price = $5
         WHERE id = $6
         RETURNING id , updated_at`,
		updateBook.Title, updateBook.Author, updateBook.ISBN,
		updateBook.Year, updateBook.Price, id,
	).Scan(&updatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	updateBook.ID = ID
	updateBook.UpdatedAt = updatedAt
	c.JSON(http.StatusOK, updateBook)
}

func deleteBook(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM books WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "book deleted successfully"})
}

// @title           Bookstore API
// @version         1.1
// @description     This is an API for a simple bookstore.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
    initDB()
    defer db.Close()
    r := gin.Default()
    r.Use(cors.Default())

    // Swagger endpoint
    // URL for Swagger: http://localhost:8080/docs/index.html
    r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    r.GET("/health", func(c *gin.Context) {
        // ... (health check code)
    })

    api := r.Group("/api/v1")
    {
        // --- Endpoints สำหรับหนังสือ (Books) ---
        api.GET("/books", getAllBooks)           // ปรับปรุงแล้ว
        api.GET("/books/search", searchBooks)     // ใหม่
        api.GET("/books/featured", getFeaturedBooks) // ใหม่
        api.GET("/books/new", getNewBooks)        // ใหม่
        api.GET("/books/discounted", getDiscountedBooks) // ใหม่
        api.GET("/books/:id", getBook)            // เดิม
        api.POST("/books", createBook)            // เดิม
        api.PUT("/books/:id", updateBook)         // เดิม
        api.DELETE("/books/:id", deleteBook)      // เดิม

        // --- Endpoint สำหรับหมวดหมู่ (Categories) ---
        api.GET("/categories", getCategories) // ใหม่
    }

    r.Run(":8080")
}