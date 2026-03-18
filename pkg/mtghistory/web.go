package mtghistory

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type TemplateData struct {
	Title         string
	Message       string
	UserCards     UserCards
	AllCards      ScryfallCards
	Owned         int
	EraCardsOwned int
	Total         int
	Percentage    string
	Hash          string
	Era           string
}

func newHTML() *template.Template {
	return template.Must(template.New("").Funcs(template.FuncMap{
		"checkIfOwned": checkIfOwned,
	},
	).ParseGlob("templates/*.gohtml"))
}

func newErrorHTML(w http.ResponseWriter, message string, statusCode int) {
	// Set the status code
	w.WriteHeader(statusCode)

	tmplData := TemplateData{
		Title:   "Magic's History",
		Message: message,
	}

	newHTML().ExecuteTemplate(w, "error.gohtml", tmplData)
}

func StartWebServer(scryfallCards ScryfallCards, totalCards int) error {
	l := Logger()

	l.Info("Starting web server on port 8080")
	// Init chi
	router := chi.NewRouter()

	// Use middlewares
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Define static assets directory
	fs := http.FileServer(http.Dir("./assets"))
	router.Handle("/assets/*", http.StripPrefix("/assets", fs))

	// Landing page route
	router.Get("/", indexPage)

	// Upload page
	router.Post("/upload", uploadPage)

	// History page
	router.Get("/history/{hash}/{era}", historyPage(scryfallCards, totalCards))

	// Start the server
	l.Info("Listening on 0.0.0.0:8080 now")
	err := http.ListenAndServe("0.0.0.0:8080", router)
	if err != nil {
		return err
	}
	return nil
}

// uploadPage handles the file upload
func uploadPage(w http.ResponseWriter, r *http.Request) {

	var userCards UserCards

	// Get the file from the form
	if err := r.ParseForm(); err != nil {
		newErrorHTML(w, "Could not parse http parameters", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		newErrorHTML(w, "No file uploaded", http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	src := io.TeeReader(file, &buf)

	// Calculate md5 hash of the converted file
	h := md5.New()
	if _, err := io.Copy(h, src); err != nil {
		newErrorHTML(w, "Problem with uploaded file", http.StatusBadRequest)
		return
	}
	hashname := hex.EncodeToString(h.Sum(nil))

	// Check if the format is Archidekt
	// If yes, convert it to our own format
	if format := r.FormValue("format"); format == "archidekt" {
		userCards, err = convertArchidektToCSV(&buf)
		if err != nil {
			newErrorHTML(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Check if the format is Deckbox
	// If yes, convert it to our own format
	if format := r.FormValue("format"); format == "deckbox" {
		userCards, err = convertDeckboxToCSV(&buf)
		if err != nil {
			newErrorHTML(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Check if the format is Moxfield
	// If yes, convert it to our own format
	if format := r.FormValue("format"); format == "moxfield" {
		userCards, err = convertMoxfieldToCSV(&buf)
		if err != nil {
			newErrorHTML(w, "Uploaded csv has the wrong format", http.StatusBadRequest)
			return
		}
	}

	// Check if the format is ManaBox
	// If yes, convert it to our own format
	if format := r.FormValue("format"); format == "manabox" {
		userCards, err = convertManaBoxToCSV(&buf)
		if err != nil {
			newErrorHTML(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Write the file to the uploads folder
	err = userCards.persist("./uploads/" + hashname)
	if err != nil {
		newErrorHTML(w, "Could not save the file", http.StatusBadRequest)
		return
	}

	// Redirect to the history page
	http.Redirect(w, r, "/history/"+hashname+"/2020s", http.StatusFound)
}

// historyPage handles the history page
func historyPage(scryfallCards ScryfallCards, noOfScryfallCards int) http.HandlerFunc {

	fn := func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		if err := r.ParseForm(); err != nil {
			newErrorHTML(w, "Could not parse http parameters", http.StatusBadRequest)
			return
		}

		era := chi.URLParam(r, "era")
		if era != "1990s" && era != "2000s" && era != "2010s" && era != "2020s" && era != "" {
			newErrorHTML(w, "Invalid era", http.StatusBadRequest)
			return
		}

		hash := chi.URLParam(r, "hash")
		if hash == "" {
			newErrorHTML(w, "Missing collection hash in url parameters", http.StatusBadRequest)
			return
		}

		if !isValidMD5(hash) {
			newErrorHTML(w, "Collection hash not valid", http.StatusBadRequest)
			return
		}

		// Load user cards
		userCards, err := loadUserCards(hash)
		if err != nil {
			newErrorHTML(w, fmt.Sprintf("Failed to load user cards: %s", err.Error()), http.StatusBadRequest)
			return
		}

		if len(userCards) == 0 {
			newErrorHTML(w, "Uploaded csv has the wrong format", http.StatusBadRequest)
			return
		}

		owned := noOfCardsOwned(userCards, scryfallCards)

		eraCards, err := filterCardsByEra(scryfallCards, era)
		if err != nil {
			newErrorHTML(w, fmt.Sprintf("Failed to filter cards by era: %s", err.Error()), http.StatusInternalServerError)
		}

		eraCardsOwned := noOfCardsOwned(userCards, eraCards)

		tmplData := TemplateData{
			Title:         "Magic's History",
			UserCards:     userCards,
			EraCardsOwned: eraCardsOwned,
			AllCards:      eraCards,
			Owned:         owned,
			Total:         noOfScryfallCards,
			Percentage:    fmt.Sprintf("%.2f", float64(owned)/float64(noOfScryfallCards)*100),
			Hash:          hash,
			Era:           era,
		}
		newHTML().ExecuteTemplate(w, "history.gohtml", tmplData)
	}
	return fn
}

// indexPage handles the landing page
func indexPage(w http.ResponseWriter, r *http.Request) {

	tmplData := TemplateData{
		Title: "Magic's History",
	}

	err := newHTML().ExecuteTemplate(w, "index.gohtml", tmplData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
