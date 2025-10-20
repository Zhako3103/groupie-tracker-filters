package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Relations    string   `json:"relations"`
}

type Locations struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
}

type Dates struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

type Relation struct {
	Index []struct {
		ID             int                 `json:"id"`
		DatesLocations map[string][]string `json:"datesLocations"`
	} `json:"index"`
}

type ArtistFull struct {
	Artist
	Locations    []string
	Dates        []string
	DatesByPlace map[string][]string
}

type PageData struct {
	Artists     []ArtistFull
	SearchQuery string
}

var artistsFull []ArtistFull

func main() {
	log.Println("Загрузка данных c API...")
	loadData()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/artist", artistHandler)
	http.HandleFunc("/api/artists", apiArtistsHandler)

	log.Println("Сервер запущен на http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func loadData() {
	var artists []Artist
	if err := fetchJSON("https://groupietrackers.herokuapp.com/api/artists", &artists); err != nil {
		log.Fatal("Ошибка запроса artists:", err)
	}

	var locations struct {
		Index []Locations `json:"index"`
	}
	if err := fetchJSON("https://groupietrackers.herokuapp.com/api/locations", &locations); err != nil {
		log.Fatal("Ошибка запроса locations:", err)
	}

	var dates struct {
		Index []Dates `json:"index"`
	}
	if err := fetchJSON("https://groupietrackers.herokuapp.com/api/dates", &dates); err != nil {
		log.Fatal("Ошибка запроса dates:", err)
	}

	var rel Relation
	if err := fetchJSON("https://groupietrackers.herokuapp.com/api/relation", &rel); err != nil {
		log.Fatal("Ошибка запроса relation:", err)
	}

	locByID := make(map[int][]string)
	for _, l := range locations.Index {
		locByID[l.ID] = l.Locations
	}

	datesByID := make(map[int][]string)
	for _, d := range dates.Index {
		datesByID[d.ID] = d.Dates
	}

	for _, a := range artists {
		af := ArtistFull{
			Artist:       a,
			Locations:    locByID[a.ID],
			Dates:        datesByID[a.ID],
			DatesByPlace: map[string][]string{},
		}

		for _, r := range rel.Index {
			if r.ID == a.ID {
				af.DatesByPlace = r.DatesLocations
				break
			}
		}

		artistsFull = append(artistsFull, af)
	}

	log.Printf("Данные загружены. Артистов: %d\n", len(artistsFull))
}

func fetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query := r.URL.Query().Get("search")
	creationMin := r.URL.Query().Get("creation_min")
	creationMax := r.URL.Query().Get("creation_max")
	albumMin := r.URL.Query().Get("album_min")
	albumMax := r.URL.Query().Get("album_max")
	members := r.URL.Query()["members"]
	location := r.URL.Query().Get("location")

	var artists []ArtistFull

	for _, artist := range artistsFull {
		// Filter by search query
		if query != "" && !strings.Contains(strings.ToLower(artist.Name), strings.ToLower(query)) {
			continue
		}
		// Filter by creation date
		if creationMin != "" {
			minYear := atoi(creationMin)
			if artist.CreationDate < minYear {
				continue
			}
		}
		if creationMax != "" {
			maxYear := atoi(creationMax)
			if artist.CreationDate > maxYear {
				continue
			}
		}
		// Filter by first album date
		if albumMin != "" {
			minYear := atoi(albumMin)
			albumYear := parseYear(artist.FirstAlbum)
			if albumYear < minYear {
				continue
			}
		}
		if albumMax != "" {
			maxYear := atoi(albumMax)
			albumYear := parseYear(artist.FirstAlbum)
			if albumYear > maxYear {
				continue
			}
		}
		// Filter by number of members
		if len(members) > 0 {
			match := false
			for _, m := range members {
				if len(artist.Members) == atoi(m) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		// Filter by location (partial match)
		if location != "" {
			found := false
			locLower := strings.ToLower(location)
			for _, loc := range artist.Locations {
				if strings.Contains(strings.ToLower(loc), locLower) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		artists = append(artists, artist)
	}

	data := PageData{
		Artists:     artists,
		SearchQuery: query,
	}

	tmpl.Execute(w, data)
}

func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func parseYear(s string) int {
	if len(s) >= 4 {
		year := 0
		fmt.Sscanf(s[:4], "%d", &year)
		return year
	}
	return 0
}

func apiArtistsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artistsFull)
}

func artistHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.NotFound(w, r)
		return
	}

	var artist ArtistFull
	found := false
	for _, a := range artistsFull {
		if idStr == itoa(a.ID) {
			artist = a
			found = true
			break
		}
	}
	if !found {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("templates/artist.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, artist)
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
