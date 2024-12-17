package main

import (
	"bytes"
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
	"strings"
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

func (uc UserCards) persist(filename string) error {
	f, err := os.Create(filename)

	if err != nil {
		return err
	}

	defer f.Close()

	for _, card := range uc {
		f.WriteString(card + "\n")
	}

	return nil
}

func convertArchidektToCSV(file io.Reader) (UserCards, error) {

	var userCards UserCards
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	// Checks for the error
	if err != nil {
		return userCards, err
	}

	// Checks for the error
	// Moxfield CSV format
	// Quantity,Name,Finish,Condition,Date Added,Language,Purchase Price,Tags,Edition Name,Edition Code,Multiverse Id,Scryfall ID,Collector Number
	//  0         1   2        3        4          5         6            7     8             9             10           11         12
	for _, row := range records {
		if len(row) < 10 {
			return userCards, err
		}
		userCards = append(userCards, fmt.Sprintf("%s,%s", strings.ToLower(row[9]), row[12]))
	}

	return userCards, nil
}

func convertMoxfieldToCSV(file io.Reader) (UserCards, error) {

	var userCards UserCards
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	// Checks for the error
	if err != nil {
		return userCards, err
	}

	// Checks for the error
	// Moxfield CSV format
	// Count,Tradelist Count,Name,Edition,Condition,Language,Foil,Tags,Last Modified,Collector Number,Alter,Proxy,Purchase Price
	//   0         1          2     3       4           5      6   7     8              9              10     11    12
	for _, row := range records {
		if len(row) < 10 {
			return userCards, err
		}
		userCards = append(userCards, fmt.Sprintf("%s,%s", strings.ToLower(row[3]), row[9]))
	}

	return userCards, nil

}

func convertManaBoxToCSV(file io.Reader) (UserCards, error) {

	var userCards UserCards
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	// Checks for the error
	if err != nil {
		return userCards, err
	}

	if records[0][0] == "Name" {
		return nil, fmt.Errorf("Please use manabox \"Collection\" export, not \"Binder\" export!")
	}

	// ManaBox CSV format
	// BinderName,BinName,Setcode,Setname,Collectornumber,Foil,Rarity,Quantity,ManaBoxID,ScryfallID,[..]
	//   0          1        2        3      4       5             6     7        8        9       10
	for _, row := range records {
		if len(row) < 4 {
			return userCards, err
		}
		userCards = append(userCards, fmt.Sprintf("%s,%s", strings.ToLower(row[3]), row[5]))
	}

	return userCards, nil

}

func convertDeckboxToCSV(file io.Reader) (UserCards, error) {

	var userCards UserCards
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()

	// Checks for the error
	if err != nil {
		return userCards, err
	}

	// ManaBox CSV format
	// Count,Tradelist Count,Name,Edition,Edition Code,Card Number,Condition,Language,Foil,[...]
	//   0      1             2        3      4          5             6          7        8        9
	for _, row := range records {
		if len(row) < 6 {
			return userCards, err
		}
		userCards = append(userCards, fmt.Sprintf("%s,%s", strings.ToLower(row[4]), row[5]))
	}

	return userCards, nil
}

// uploadPage handles the file upload
func uploadPage(c *gin.Context) {

	var userCards UserCards

	// Get the file from the form
	file, err := c.FormFile("file")
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"title":   "Magic's History",
			"message": "No file uploaded",
		})
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"title":   "Magic's History",
			"message": "Problem with uploaded file",
		})
	}

	// Calculate md5 hash of the converted file
	h := md5.New()
	var buf bytes.Buffer
	tr := io.TeeReader(src, &buf)
	if _, err := io.Copy(h, tr); err != nil {
		log.Fatal(err)
	}
	hashname := hex.EncodeToString(h.Sum(nil))

	// Check if the format is Archidekt
	// If yes, convert it to our own format
	if format := c.PostForm("format"); format == "archidekt" {
		userCards, err = convertArchidektToCSV(&buf)

		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"title":   "Magic's History",
				"message": err.Error(),
			})
		}
	}

	// Check if the format is Deckbox
	// If yes, convert it to our own format
	if format := c.PostForm("format"); format == "deckbox" {
		userCards, err = convertDeckboxToCSV(&buf)

		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"title":   "Magic's History",
				"message": err.Error(),
			})
		}
	}

	// Check if the format is Moxfield
	// If yes, convert it to our own format
	if format := c.PostForm("format"); format == "moxfield" {
		userCards, err = convertMoxfieldToCSV(&buf)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"title":   "Magic's History",
				"message": "Uploaded csv has the wrong format",
			})
		}
	}

	// Check if the format is ManaBox
	// If yes, convert it to our own format
	if format := c.PostForm("format"); format == "manabox" {
		userCards, err = convertManaBoxToCSV(&buf)

		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"title":   "Magic's History",
				"message": err.Error(),
			})
		}
	}

	// Write the file to the uploads folder
	err = userCards.persist("./uploads/" + hashname)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
			"title":   "Magic's History",
			"message": "Cloud not save the file",
		})
	}

	// Redirect to the history page
	c.Redirect(http.StatusFound, "/history/"+hashname)
}

// historyPage handles the history page
func historyPage(allCards AllCards) gin.HandlerFunc {

	fn := func(c *gin.Context) {

		// Get the hash from the URL
		var params Params
		if err := c.ShouldBindUri(&params); err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"title":   "Magic's History",
				"message": fmt.Sprintf("Failed to bind uri: %s", err.Error()),
			})
			return
		}

		// Load user cards
		userCards, err := loadUserCards(params.Hash)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"title":   "Magic's History",
				"message": fmt.Sprintf("Failed to load user cards: %s", params.Hash),
			})
			return
		}

		if len(userCards) == 0 {
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{
				"title":   "Magic's History",
				"message": "Uploaded csv has the wrong format",
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

	// Load all from pre constructed json
	allCards, err := loadAllCards("./assets/allcards.json")
	if err != nil {
		log.Fatal(err)
	}

	// Setup new web server
	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"checkIfOwned": checkIfOwned,
	})

	// Load the templates
	router.LoadHTMLGlob("templates/*.tmpl")
	router.Static("/assets", "./assets")

	// Landing page
	router.GET("/", landingPage)

	// Upload page
	router.POST("/upload", uploadPage)

	// History page
	router.GET("/history/:hash", historyPage(allCards))

	// Run the server
	router.Run("0.0.0.0" + ":" + strconv.FormatUint(8080, 10))
}
