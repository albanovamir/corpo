package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve static files (HTML, JS, CSS)
	r.Static("/static", "./static")

	// API endpoint for file upload and counting
	r.POST("/count", handleCount)

	// Start server
	r.Run(":8080")
}

func handleCount(c *gin.Context) {
	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Open the uploaded file
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	// Count words, lines and symbols
	wordCount, lineCount, symbolCount, err := count(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the results
	c.JSON(http.StatusOK, gin.H{
		"words":   wordCount,
		"lines":   lineCount,
		"symbols": symbolCount,
	})
}

func count(reader io.Reader) (int, int, int, error) {
	var wordCount, lineCount, symbolCount int

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Split line into words
		words := strings.Fields(line)
		wordCount += len(words)

		// Count symbols in each word
		for _, word := range words {
			symbolCount += utf8.RuneCountInString(word)
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, 0, fmt.Errorf("error reading file: %v", err)
	}

	return wordCount, lineCount, symbolCount, nil
}

// http://localhost:8080/static/index.html
