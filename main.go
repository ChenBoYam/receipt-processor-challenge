package main

import (
    "math"
    "net/http"
    "strconv"
    "strings"
    "sync"
    "time"
    "unicode"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// Receipt represents the structure of a receipt
type Receipt struct {
    Retailer     string
    PurchaseDate time.Time
    PurchaseTime time.Time
    Items        []Item
    Total        float64
}

// Item represents a single item on a receipt
type Item struct {
    ShortDescription string
    Price            float64
}

var (
    // receipts[id] = receipt
    receipts = make(map[string]Receipt)
    // lock for thread safe
    mu       sync.Mutex
)

// main initializes the server
// The application exposes two main endpoints:
// - POST /receipts/process: Processes new receipts
// - GET /receipts/:id/points: Retrieves points for a specific receipt
//                             id: [uuid-id]
// Input: none
// Output: starts HTTP server on port 8080

func main() {
    // Logger middleware
    router := gin.Default()
    /*
        
    */
    router.POST("/receipts/process", processReceipt)
    router.GET("/receipts/:id/points", getPoints)
    router.Run(":8080")
}

// processReceipt processes a new receipt
// Input: 
//   JSON receipt data in request body:
//   - retailer: string
//   - purchaseDate: string (YYYY-MM-DD)
//   - purchaseTime: string (HH:MM)
//   - items: array of {shortDescription: string, price: string}
//   - total: string
// Output: 
//   - Success: JSON with receipt ID {"id": "uuid-id"}
//   - Error: JSON with error message {"error": "message"}
func processReceipt(c *gin.Context) {
    // Input template
    var input struct {
        Retailer     string `json:"retailer"`
        PurchaseDate string `json:"purchaseDate"`
        PurchaseTime string `json:"purchaseTime"`
        Items        []struct {
            ShortDescription string `json:"shortDescription"`
            Price            string `json:"price"`
        } `json:"items"`
        Total string `json:"total"`
    }
    // c.ShouldBindJSON for parsing JSON
    if err := c.ShouldBindJSON(&input); err != nil {
        // c.JSON for responses
        // gin.H is a shorthand for map[string]interface{}
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
        return
    }

    // Validate and parse receipt data
    purchaseDate, err := time.Parse("2006-01-02", input.PurchaseDate)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchaseDate format"})
        return
    }
    
    // Validate and parse receipt time
    purchaseTime, err := time.Parse("15:04", input.PurchaseTime)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchaseTime format"})
        return
    }
    // Validate and parse receipt total price
    total, err := strconv.ParseFloat(input.Total, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid total"})
        return
    }
    // Validate receipt's purchase items > 0
    if len(input.Items) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "at least one item required"})
        return
    }
    // Validate and parse receipt purchase items
    items := make([]Item, len(input.Items))
    for i, item := range input.Items {
        price, err := strconv.ParseFloat(item.Price, 64)
        if err != nil || price < 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item price"})
            return
        }
        items[i] = Item{
            ShortDescription: item.ShortDescription,
            Price:            price,
        }
    }
    // Map parsed receipt items
    receipt := Receipt{
        Retailer:     input.Retailer,
        PurchaseDate: purchaseDate,
        PurchaseTime: purchaseTime,
        Items:        items,
        Total:        total,
    }
    // Generating new uuid-id
    id := uuid.New().String()
    // Lock for thread safe while modifying data
    mu.Lock()
    // map receipt with its unique uuid-id
    receipts[id] = receipt
    mu.Unlock()

    c.JSON(http.StatusOK, gin.H{"id": id})
}

// getPoints retrieves points for a receipt
// Input: 
//   - [uuid-id]: receipt ID in URL path parameter
// Output:
//   - Success: JSON with points {"points": number}
//   - Error: JSON with error {"error": "receipt not found"}
func getPoints(c *gin.Context) {
    // c.Param for URL parameters
    id := c.Param("id")
    // Lock for thread safe while Accessing data
    mu.Lock()
    receipt, exists := receipts[id]
    mu.Unlock()

    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "receipt not found"})
        return
    }

    points := calculatePoints(receipt)
    
    c.JSON(http.StatusOK, gin.H{"points": points})
}

// calculatePoints calculates total points for a receipt
// Input: Receipt struct containing receipt details
// Output: integer 
func calculatePoints(receipt Receipt) int {
    points := 0

    // Rule 1: Retailer name alphanumeric characters
    for _, r := range receipt.Retailer {
        if unicode.IsLetter(r) || unicode.IsDigit(r) {
            points++
        }
    }

    // Rule 2: Round dollar amount
    if receipt.Total == math.Trunc(receipt.Total) {
        points += 50
    }

    // Rule 3: Multiple of 0.25
    if math.Mod(receipt.Total, 0.25) == 0 {
        points += 25
    }

    // Rule 4: 5 points per two items
    points += (len(receipt.Items) / 2) * 5

    // Rule 5: Item description length multiple of 3
    for _, item := range receipt.Items {
        // TrimSpace removes leading and trailing white space
        trimmed := strings.TrimSpace(item.ShortDescription)
        if len(trimmed)%3 == 0 {
            points += int(math.Ceil(item.Price * 0.2))
        }
    }

    // Rule 6: Odd purchase day
    if receipt.PurchaseDate.Day()%2 != 0 {
        points += 6
    }

    // Rule 7: Purchase time between 2pm and 4pm
    hour := receipt.PurchaseTime.Hour()
    if hour >= 14 && hour < 16 {
        points += 10
    }

    return points
}