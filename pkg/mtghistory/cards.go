// Package mtghistory is a package that provides functions to load and process Magic:
// The Gathering card data. It includes functions to load user cards from a CSV
// file, load all cards from a JSON file, and calculate the number of cards
// owned by a user. The package also includes a function to check if a specific
// card is owned by the user and a method to persist user cards to a file.
package mtghistory

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"slices"
)

var Version = "unknown"

type Card struct {
	Set              string   `json:"set"`
	CollectionNumber string   `json:"collector_number"`
	Games            []string `json:"games"`
	SetType          string   `json:"set_type"`
	ReleaseDate      string   `json:"released_at"`
	Year             int64
}

type UserCards []string

type ScryfallCards []ScryfallYearCards

type ScryfallYearCards struct {
	Year  int64    `json:"year"`
	Cards []string `json:"cards"`
}

func isValidMD5(md5sum string) bool {
	// Regular expression to match a valid MD5 hash
	var md5Regex = regexp.MustCompile("^[a-f0-9]{32}$")
	return md5Regex.MatchString(md5sum)
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

func noOfCardsOwned(userCards UserCards, scryfallCards ScryfallCards) int {
	owned := 0
	for _, year := range scryfallCards {
		for _, card := range year.Cards {
			if slices.Contains(userCards, card) {
				owned = owned + 1
			}
		}
	}
	return owned
}

func checkIfOwned(card string, userCards []string) bool {
	return slices.Contains(userCards, card)
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
