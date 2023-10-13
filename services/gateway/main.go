package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

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

type Gateway struct {
	searchAddr  string
	profileAddr string
}

func NewGateway(searchAddr, profileAddr string) *Gateway {
	return &Gateway{
		searchAddr:  searchAddr,
		profileAddr: profileAddr,
	}
}

func (g *Gateway) Run(port int) error {
	mux := http.NewServeMux()
	mux.Handle("/hotels", http.HandlerFunc(g.searchHandler))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func (g *Gateway) searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	latParam, lonParam := r.URL.Query().Get("lat"), r.URL.Query().Get("lon")
	if latParam == "" || lonParam == "" {
		http.Error(w, "Please specify lat/lon params", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(latParam), 64)
	if err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(strings.TrimSpace(lonParam), 64)
	if err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	searchResp, err := http.Get(fmt.Sprintf("%s/nearby?lat=%f&lon=%f&inDate=%s&outDate=%s", g.searchAddr, lat, lon, inDate, outDate))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer searchResp.Body.Close()

	var searchResult struct {
		HotelIds []string `json:"hotelIds"`
	}
	if err := json.NewDecoder(searchResp.Body).Decode(&searchResult); err != nil {
		http.Error(w, "Failed to decode search response", http.StatusInternalServerError)
		return
	}

	hotelIds := strings.Join(searchResult.HotelIds, ",")
	profileResp, err := http.Get(fmt.Sprintf("%s/profile?ids=%s", g.profileAddr, hotelIds))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer profileResp.Body.Close()

	var profileResponse struct {
		Hotels []*Hotel `json:"hotels"`
	}
	if err := json.NewDecoder(profileResp.Body).Decode(&profileResponse); err != nil {
		http.Error(w, "Failed to decode profile response", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(geoJSONResponse(profileResponse.Hotels))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	var (
		port           = flag.Int("port", 8080, "The service port")
		profileAddress = flag.String("profileAddress", "http://profile:8080", "Profile service address")
		searchAddress  = flag.String("searchAddress", "http://search:8080", "Search service address")
	)
	flag.Parse()

	gw := NewGateway(*searchAddress, *profileAddress)
	if err := gw.Run(*port); err != nil {
		log.Fatalf("run server error: %v", err)
	}
}

func geoJSONResponse(hotels []*Hotel) map[string]interface{} {
	var fs []interface{}

	for _, h := range hotels {
		fs = append(fs, map[string]interface{}{
			"type": "Feature",
			"id":   h.Id,
			"properties": map[string]string{
				"name":         h.Name,
				"phone_number": h.PhoneNumber,
			},
			"geometry": map[string]interface{}{
				"type": "Point",
				"coordinates": []float64{
					h.Address.Lon,
					h.Address.Lat,
				},
			},
		})
	}

	return map[string]interface{}{
		"type":     "FeatureCollection",
		"features": fs,
	}
}
