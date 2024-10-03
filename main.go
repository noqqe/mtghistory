package main

import (
	"bufio"
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

func loadAllCards(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %w", err)
	}

	return lines, nil
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

func historyPage(c *gin.Context) {
	var params Params
	log.Println("Params: ", params)
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

	// static for now
	totalCards := 40201

	c.HTML(http.StatusOK, "history.tmpl", gin.H{
		"title":      "Magic's History",
		"usercards":  userCards,
		"owned":      len(userCards),
		"total":      totalCards,
		"percentage": fmt.Sprintf("%.02f", float64(len(userCards))/float64(totalCards)*100),
		"hash":       params.Hash,
		"years": []int{2024,
			2023,
			2022,
			2021,
			2020,
			2019,
			2018,
			2017,
			2016,
			2015,
			2014,
			2013,
			2012,
			2011,
			2010,
			2009,
			2008,
			2007,
			2006,
			2005,
			2004,
			2003,
			2002,
			2001,
			2000,
			1999,
			1998,
			1997,
			1996,
			1995,
			1994,
			1993,
		},
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
