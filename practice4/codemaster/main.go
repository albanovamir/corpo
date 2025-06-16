// http://localhost:8080/static/index.html

package main

import (
	"encoding/xml"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Game struct {
	SecretCode  string
	Players     []*Player
	CurrentTurn int
	Attempts    map[string]int
	StartTime   time.Time
	EndTime     time.Time
	Winner      string
	MaxAttempts int
	Mutex       sync.Mutex
}

type Player struct {
	ID     string
	Name   string
	Active bool
}

type GameResult struct {
	XMLName     xml.Name     `xml:"gameResult"`
	StartTime   string       `xml:"startTime"`
	EndTime     string       `xml:"endTime"`
	SecretCode  string       `xml:"secretCode"`
	Winner      string       `xml:"winner"`
	PlayerStats []PlayerStat `xml:"playerStats>playerStat"`
}

type PlayerStat struct {
	Name     string `xml:"name"`
	Attempts int    `xml:"attempts"`
}

var game *Game

func main() {
	rand.Seed(time.Now().UnixNano())
	initGame()

	r := gin.Default()
	r.Static("/static", "./static")

	r.POST("/join", handleJoin)
	r.POST("/guess", handleGuess)
	r.GET("/status", handleStatus)

	go func() {
		for {
			time.Sleep(5 * time.Second)
			checkGameState()
		}
	}()

	r.Run(":8080")
}

func initGame() {
	game = &Game{
		SecretCode:  generateCode(4),
		Players:     make([]*Player, 0),
		Attempts:    make(map[string]int),
		StartTime:   time.Now(),
		MaxAttempts: 10,
	}
	fmt.Println("New game started. Secret code:", game.SecretCode)
}

func generateCode(length int) string {
	code := ""
	for i := 0; i < length; i++ {
		code += fmt.Sprintf("%d", rand.Intn(10))
	}
	fmt.Println(code)
	return code
}

func handleJoin(c *gin.Context) {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	if len(game.Players) >= 4 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Maximum players reached"})
		return
	}

	var player struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&player); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newPlayer := &Player{
		ID:     fmt.Sprintf("player%d", len(game.Players)+1),
		Name:   player.Name,
		Active: true,
	}

	game.Players = append(game.Players, newPlayer)
	game.Attempts[newPlayer.ID] = 0

	c.JSON(http.StatusOK, gin.H{
		"message":  "Joined successfully",
		"playerId": newPlayer.ID,
		"codeLen":  len(game.SecretCode),
	})
}

func handleGuess(c *gin.Context) {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	var request struct {
		PlayerID string `json:"playerId"`
		Guess    string `json:"guess"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate it's the player's turn
	currentPlayer := game.Players[game.CurrentTurn]
	if currentPlayer.ID != request.PlayerID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your turn"})
		return
	}

	// Validate guess length
	if len(request.Guess) != len(game.SecretCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Guess must be %d digits", len(game.SecretCode))})
		return
	}

	game.Attempts[request.PlayerID]++

	// Check guess
	black, white := evaluateGuess(game.SecretCode, request.Guess)

	// Check for win
	if black == len(game.SecretCode) {
		game.Winner = currentPlayer.Name
		game.EndTime = time.Now()
		saveGameResult()
		c.JSON(http.StatusOK, gin.H{
			"result":   "win",
			"black":    black,
			"white":    white,
			"attempts": game.Attempts[request.PlayerID],
		})
		return
	}

	// Switch to next player
	game.CurrentTurn = (game.CurrentTurn + 1) % len(game.Players)

	c.JSON(http.StatusOK, gin.H{
		"result":   "continue",
		"black":    black,
		"white":    white,
		"attempts": game.Attempts[request.PlayerID],
		"nextTurn": game.Players[game.CurrentTurn].Name,
	})
}

func evaluateGuess(secret, guess string) (black, white int) {
	secretRunes := []rune(secret)
	guessRunes := []rune(guess)

	// Check for black markers (correct position)
	for i := 0; i < len(secretRunes); i++ {
		if secretRunes[i] == guessRunes[i] {
			black++
			secretRunes[i] = 'x' // Mark as counted
			guessRunes[i] = 'y'  // Mark as counted
		}
	}

	// Check for white markers (correct number but wrong position)
	for i := 0; i < len(secretRunes); i++ {
		for j := 0; j < len(guessRunes); j++ {
			if i != j && secretRunes[i] == guessRunes[j] {
				white++
				secretRunes[i] = 'x' // Mark as counted
				guessRunes[j] = 'y'  // Mark as counted
				break
			}
		}
	}

	return black, white
}

func handleStatus(c *gin.Context) {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	playerID := c.Query("playerId")

	var playerNames []string
	for _, p := range game.Players {
		playerNames = append(playerNames, p.Name)
	}

	status := gin.H{
		"players":      playerNames,
		"currentTurn":  game.Players[game.CurrentTurn].Name,
		"attempts":     game.Attempts[playerID],
		"gameFinished": game.Winner != "",
	}

	if game.Winner != "" {
		status["winner"] = game.Winner
	}

	c.JSON(http.StatusOK, status)
}

func checkGameState() {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	if game.Winner != "" {
		// Start new game
		initGame()
		return
	}

	// Check if any player reached max attempts
	for _, player := range game.Players {
		if game.Attempts[player.ID] >= game.MaxAttempts {
			game.EndTime = time.Now()
			saveGameResult()
			initGame()
			return
		}
	}
}

func saveGameResult() {
	result := GameResult{
		StartTime:  game.StartTime.Format(time.RFC3339),
		EndTime:    game.EndTime.Format(time.RFC3339),
		SecretCode: game.SecretCode,
		Winner:     game.Winner,
	}

	for _, player := range game.Players {
		result.PlayerStats = append(result.PlayerStats, PlayerStat{
			Name:     player.Name,
			Attempts: game.Attempts[player.ID],
		})
	}

	xmlData, err := xml.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling XML:", err)
		return
	}

	filename := fmt.Sprintf("game_results_%s.xml", game.StartTime.Format("20060102_150405"))
	err = os.WriteFile(filename, xmlData, 0644)
	if err != nil {
		fmt.Println("Error writing XML file:", err)
		return
	}

	fmt.Println("Game result saved to", filename)
}
