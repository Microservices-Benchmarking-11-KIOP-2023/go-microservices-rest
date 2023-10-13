package main

import (
	"encoding/json"
	"github.com/harlow/go-micro-services/services/profile/data"
	"log"
	"net/http"
	"strings"
)

type Profile struct {
	Profiles map[string]*Hotel `json:"profiles"`
}

type Hotel struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	PhoneNumber string  `json:"phoneNumber"`
	Description string  `json:"description"`
	Address     Address `json:"address"`
	Images      []Image `json:"images"`
}

type Address struct {
	StreetNumber string  `json:"streetNumber"`
	StreetName   string  `json:"streetName"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	Country      string  `json:"country"`
	PostalCode   string  `json:"postalCode"`
	Lat          float64 `json:"lat"`
	Lon          float64 `json:"lon"`
}

type Image struct {
	URL     string `json:"url"`
	Default bool   `json:"default"`
}

func (p *Profile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idsParam := r.URL.Query().Get("ids")
	hotelIds := strings.Split(idsParam, ",")

	var responseHotels []*Hotel

	for _, hotelId := range hotelIds {
		if hotel, ok := p.Profiles[strings.TrimSpace(hotelId)]; ok {
			responseHotels = append(responseHotels, hotel)
		}
	}

	response := struct {
		Hotels []*Hotel `json:"hotels"`
	}{
		Hotels: responseHotels,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func loadProfiles(path string) map[string]*Hotel {
	var (
		file   = data.MustAsset(path)
		hotels []*Hotel
	)

	if err := json.Unmarshal(file, &hotels); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	profiles := make(map[string]*Hotel)
	for _, hotel := range hotels {
		profiles[hotel.Id] = hotel
	}
	return profiles
}

func main() {
	profiles := loadProfiles("data/hotels.json")
	profileServer := &Profile{Profiles: profiles}

	http.Handle("/profile", profileServer)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
