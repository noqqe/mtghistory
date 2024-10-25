package main

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/gin-gonic/gin"
)

type Params struct {
	Hash string `uri:"hash" binding:"required"`
}

type UserCards []string

type AllCards []struct {
	Year  string   `json:"year"`
	Cards []string `json:"cards"`
}

func loadAllCards(filename string) (AllCards, error) {
	file, err := os.Open(filename)

	// Checks for the error
	if err != nil {
		err := fmt.Errorf("failed to open file: %w", err)
		return nil, err
	}

	// Closes the file
	defer file.Close()
	byteValue, _ := io.ReadAll(file)

	var result AllCards
	json.Unmarshal(byteValue, &result)

	return result, nil
}

func loadUserCards(filename string) (UserCards, error) {

	var userCards UserCards
	file, err := os.Open("./uploads/" + filename)

	// Checks for the error
	if err != nil {
		err := fmt.Errorf("failed to open file: %w", err)
		return nil, err
	}

	// Closes the file
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	// Checks for the error
	if err != nil {
		return userCards, err
	}

	// Loop to iterate through
	// and print each of the string slice
	for _, row := range records {
		userCards = append(userCards, fmt.Sprintf("%s/%s", row[0], row[1]))
	}
	return userCards, nil
}

func checkIfOwned(card string, userCards []string) bool {
	for _, a := range userCards {
		if a == card {
			return true
		}
	}
	return false
}

// uploadPage handles the file upload
func uploadPage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "get form err: %s", err.Error())
		return
	}

	h := md5.New()
	src, _ := file.Open()
	defer src.Close()

	if _, err := io.Copy(h, src); err != nil {
		log.Fatal(err)
	}
	hashname := hex.EncodeToString(h.Sum(nil))
	if err := c.SaveUploadedFile(file, "./uploads/"+hashname); err != nil {
		c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
		return
	}

	c.Redirect(http.StatusFound, "/history/"+hashname)
}

// historyPage handles the history page
func historyPage(allCards AllCards) gin.HandlerFunc {

	fn := func(c *gin.Context) {

		// Get the hash from the URL
		var params Params
		if err := c.ShouldBindUri(&params); err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"title": "Magic's History",
			})
			return
		}

		// Load user cards
		userCards, err := loadUserCards(params.Hash)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"title": "Magic's History",
			})
			return
		}

		// TODO: static for now. Load from json in assets folder
		totalCards := 40201

		c.HTML(http.StatusOK, "history.tmpl", gin.H{
			"title":      "Magic's History",
			"usercards":  userCards,
			"allcards":   allCards,
			"owned":      len(userCards),
			"total":      totalCards,
			"percentage": fmt.Sprintf("%.02f", float64(len(userCards))/float64(totalCards)*100),
			"hash":       params.Hash,
		})
	}
	return gin.HandlerFunc(fn)
}

func landingPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Magic's History",
	})
}

func main() {

	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"checkIfOwned": checkIfOwned,
	})
	router.LoadHTMLGlob("templates/*.tmpl")
	router.Static("/assets", "./assets")

	allCards, err := loadAllCards("./assets/allcards.json")
	if err != nil {
		log.Fatal(err)
	}

	// Landing page
	router.GET("/", landingPage)

	// Upload page
	router.POST("/upload", uploadPage)

	// History page
	router.GET("/history/:hash", historyPage(allCards))

	router.Run("0.0.0.0" + ":" + strconv.FormatUint(8080, 10))
}
