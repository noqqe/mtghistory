package mtghistory

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type BulkIndex struct {
	Object  string `json:"object"`
	HasMore bool   `json:"has_more"`
	Data    []struct {
		Object          string    `json:"object"`
		ID              string    `json:"id"`
		Type            string    `json:"type"`
		UpdatedAt       time.Time `json:"updated_at"`
		URI             string    `json:"uri"`
		Name            string    `json:"name"`
		Description     string    `json:"description"`
		Size            int       `json:"size"`
		DownloadURI     string    `json:"download_uri"`
		ContentType     string    `json:"content_type"`
		ContentEncoding string    `json:"content_encoding"`
	} `json:"data"`
}

func FetchScryfallCards() ([]Card, error) {
	l := Logger()
	updatedCards := []Card{}

	// Fetch bulk file
	l.Infof("Fetching bulk data from scryfall...")
	downloadURL, err := fetchBulkDownloadURL()
	if err != nil {
		l.Error("Could not extract bulk download URL:", err)
		return updatedCards, err
	}

	l.Infof("Downloading bulk data file %s", downloadURL)
	bulkFilePath, err := downloadBulkData(downloadURL)
	if err != nil {
		l.Error("Could not fetch bulk json from scryfall", err)
		return updatedCards, err
	}

	l.Infof("Loading bulk data file %s", bulkFilePath)
	updatedCards, err = LoadBulkFile(bulkFilePath, true)
	if err != nil {
		l.Error("Could not load bulk file:", err)
		return updatedCards, err
	}

	if len(updatedCards) == 0 {
		l.Warnf("No cards were loaded from the bulk file %s", bulkFilePath)
		return updatedCards, nil
	}

	l.Infof("Successfully loaded %d cards", len(updatedCards))

	return updatedCards, nil
}

// http getter for scryfall api with custom headers
func queryScryfall(url string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json;q=0.9,*/*;q=0.8")
	req.Header.Set("User-Agent", fmt.Sprintf("mtghistory/%s", Version))
	return client.Do(req)
}

func fetchBulkDownloadURL() (string, error) {
	l := Logger()
	downloadURL := ""

	// Make an HTTP GET request
	resp, err := queryScryfall("https://api.scryfall.com/bulk-data")
	if err != nil {
		l.Fatalf("Error fetching data: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Fatalf("Error reading response body: %v", err)
	}

	// Unmarshal the JSON response
	var bulkData BulkIndex
	if err := json.Unmarshal(body, &bulkData); err != nil {
		l.Fatalf("Error unmarshaling JSON: %v", err)
	}

	// Find and print the unique cards URL
	for _, item := range bulkData.Data {
		if item.Type == "default_cards" {
			downloadURL = item.DownloadURI
		}
	}

	return downloadURL, nil
}

func downloadBulkData(downloadURL string) (string, error) {
	l := Logger()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "download")
	if err != nil {
		l.Fatalf("Error creating temporary directory: %v", err)
	}

	// Create a temporary file in the temporary directory
	tempFile, err := os.CreateTemp(tempDir, "downloaded-*.json") // Adjust the extension if necessary
	if err != nil {
		l.Fatalf("Error creating temporary file: %v", err)
	}
	defer tempFile.Close() // Ensure we close the file when we're done

	// Download the file
	resp, err := http.Get(downloadURL)
	if err != nil {
		l.Fatalf("Error downloading file: %v", err)
	}
	defer resp.Body.Close() // Make sure to close the response body

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		l.Fatalf("Error: received status code %d", resp.StatusCode)
	}

	// Copy the response body to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		l.Fatalf("Error saving file: %v", err)
	}

	return tempFile.Name(), nil
}

func LoadBulkFile(bulkFilePath string, cleanUp bool) ([]Card, error) {
	var cards []Card
	l := Logger()
	fileBytes, err := os.ReadFile(bulkFilePath)
	if err != nil {
		l.Fatalf("Error reading bulk file: %v", err)
	}

	if cleanUp {
		os.Remove(bulkFilePath)
	}

	err = json.Unmarshal(fileBytes, &cards)
	if err != nil {
		return cards, err
	}

	return cards, nil

}
