package mtghistory

import (
	"encoding/csv"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

func ConvertCardstoCardIDs(fetchedCards []Card) (ScryfallCards, int, error) {

	var scryfallCards ScryfallCards
	totalCards := 0

	// generate all years from current year to 1993
	thisYear := time.Now().Year()
	for i := thisYear; 1993 <= i; i-- {
		scryfallCards = append(scryfallCards, ScryfallYearCards{Year: int64(i), Cards: []string{}})
	}

	for _, card := range fetchedCards {
		if !slices.Contains(card.Games, "paper") {
			continue
		}

		if card.SetType != "core" && card.SetType != "expansion" && card.SetType != "masters" {
			continue
		}

		// set year
		y, _ := strconv.ParseInt(card.ReleaseDate[:4], 10, 32)
		card.Year = y

		for i := range scryfallCards {
			if scryfallCards[i].Year == card.Year {
				scryfallCards[i].Cards = append(scryfallCards[i].Cards, fmt.Sprintf("%s/%s", strings.ToLower(card.Set), card.CollectionNumber))
				totalCards = totalCards + 1
				continue
			}
		}
	}

	// order by set and collection number
	for i := range scryfallCards {
		slices.Sort(scryfallCards[i].Cards)
	}

	return scryfallCards, totalCards, nil
}

func filterCardsByEra(scryfallCards ScryfallCards, era string) (ScryfallCards, error) {
	var eraCards ScryfallCards
	switch era {
	case "1990s":
		for _, cards := range scryfallCards {
			if cards.Year >= 1993 && cards.Year <= 1999 {
				eraCards = append(eraCards, cards)
			}
		}
	case "2000s":
		for _, cards := range scryfallCards {
			if cards.Year >= 2000 && cards.Year <= 2009 {
				eraCards = append(eraCards, cards)
			}
		}
	case "2010s":
		for _, cards := range scryfallCards {
			if cards.Year >= 2010 && cards.Year <= 2019 {
				eraCards = append(eraCards, cards)
			}
		}
	case "2020s":
		for _, cards := range scryfallCards {
			if cards.Year >= 2020 && cards.Year <= 2029 {
				eraCards = append(eraCards, cards)
			}
		}

	default:
		return eraCards, fmt.Errorf("invalid era")
	}

	return eraCards, nil
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
		return nil, fmt.Errorf("please use manabox \"Collection\" export, not \"Binder\" export")
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
