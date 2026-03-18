package main

import (
	"fmt"
	"os"

	"github.com/noqqe/mtghistory/pkg/mtghistory"
)

func main() {

	l := mtghistory.Logger()
	var fetchedCards []mtghistory.Card
	var err error

	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("mtghistory version %s\n", mtghistory.Version)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "--bulk-file" {
		l.Infof("Loading bulk data file %s", os.Args[2])
		fetchedCards, err = mtghistory.LoadBulkFile(os.Args[2], false)
		if err != nil {
			l.Fatal(err)
		}
	} else {
		fetchedCards, err = mtghistory.FetchScryfallCards()
		if err != nil {
			l.Fatal(err)
		}
	}

	scryfallCards, noOfScryfallCards, err := mtghistory.ConvertCardstoCardIDs(fetchedCards)
	if err != nil {
		l.Fatal(err)
	}

	err = mtghistory.StartWebServer(scryfallCards, noOfScryfallCards)
	if err != nil {
		l.Fatal(err)
	}

}
