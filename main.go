package main

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Params struct {
	Hash string `uri:"hash" binding:"required"`
}

func loadUserCards(filename string) ([]string, error) {

	var userCards []string
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
	print(userCards)
	return userCards, nil
}

func checkIfCardExists(card string, my_cards []string) bool {
	for _, a := range my_cards {
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
func historyPage(c *gin.Context) {
	var params Params
	if err := c.ShouldBindUri(&params); err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"title": "Magic's History",
		})
		return
	}

	userCards, err := loadUserCards(params.Hash)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"title": "Magic's History",
		})
		return
	}

	// TODO: static for now. Load from json in assets folder
	totalCards := 40201

	// TODO: static for now. Load from json in assets folder
	years := []int{}
	for i := 1993; i <= 2024; i++ {
		years = append(years, i)
	}

	c.HTML(http.StatusOK, "history.tmpl", gin.H{
		"title":      "Magic's History",
		"usercards":  userCards,
		"owned":      len(userCards),
		"total":      totalCards,
		"percentage": fmt.Sprintf("%.02f", float64(len(userCards))/float64(totalCards)*100),
		"hash":       params.Hash,
		"years":      years,
	})
}

func landingPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Magic's History",
	})
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*.tmpl")
	router.Static("/assets", "./assets")

	// Landing page
	router.GET("/", landingPage)

	// Upload page
	router.POST("/upload", uploadPage)

	// History page
	router.GET("/history/:hash", historyPage)

	router.Run("0.0.0.0" + ":" + strconv.FormatUint(8080, 10))
}
