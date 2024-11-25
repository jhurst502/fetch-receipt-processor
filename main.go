package main

import (
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var receiptPoints = make(map[string]int)

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

func main() {
	router := gin.Default()
	router.POST("/receipts/process", processReceipts)
	router.GET("/receipts/:id/points", getReceiptPointsByID)
	router.Run("localhost:8080")
}

func processReceipts(c *gin.Context) {
	var points = 0
	var receipt Receipt
	if err := c.BindJSON(&receipt); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid receipt",
			"error":   err.Error(),
		})
		return
	}

	// One point for every alphanumeric character in the retailer name.
	for _, char := range receipt.Retailer {
		if unicode.IsLetter(char) || unicode.IsNumber(char) {
			points++
		}
	}

	// 50 points if the total is a round dollar amount with no cents.
	total, _ := strconv.ParseFloat(receipt.Total, 64)
	if isWholeNumber(total) {
		points += 50
	}

	// 25 points if the total is a multiple of 0.25.
	remainder := math.Mod(total, 0.25)
	if remainder == 0 {
		points += 25
	}

	// 5 points for every two items on the receipt.
	var size = len(receipt.Items)
	points += int(math.Floor(float64(size)/2)) * 5

	// If the trimmed length of the item description is a multiple of 3, multiply the price by 0.2 and round up to the nearest integer. The result is the number of points earned.
	for _, item := range receipt.Items {
		str := item.ShortDescription
		trimmedStr := strings.TrimSpace(str)
		trimmedLength := len(trimmedStr)
		if trimmedLength%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			var pointsToAdd = math.Ceil(price * 0.2)
			points += int(pointsToAdd)
		}
	}

	// 6 points if the day in the purchase date is odd.
	split := strings.Split(receipt.PurchaseDate, "-")
	var purchaseDate = split[len(split)-1]
	var purchaseDateInt, _ = strconv.Atoi(purchaseDate)

	if purchaseDateInt%2 != 0 {
		points += 6
	}

	// 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	var purchaseTime = strings.Replace(receipt.PurchaseTime, ":", "", 1)
	var purchaseTimeInt, _ = strconv.Atoi(purchaseTime)
	if purchaseTimeInt > 1400 && purchaseTimeInt < 1600 {
		points += 10
	}

	// Process the receipt and generate a unique ID
	uuid := uuid.New()

	// Add id and value to map
	receiptPoints[uuid.String()] = points

	c.JSON(200, gin.H{
		"id": uuid.String(),
	})
}

func getReceiptPointsByID(c *gin.Context) {
	id := c.Param("id")

	result, ok := receiptPoints[id]
	if ok {
		c.JSON(200, gin.H{
			"points": result,
		})
	} else {
		c.JSON(404, gin.H{
			"message": "Receipt not found",
		})
	}
}

func isWholeNumber(x float64) bool {
	return math.Ceil(x) == x
}
