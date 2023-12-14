package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/harlow/go-micro-services/geo/data"
	"log"
	"math"
	"net/http"
	"strconv"
)

const (
	maxSearchRadius = 10
	earthRadiusKm   = 6371
)

// point represents a hotel's geographic location on map
type point struct {
	Pid  string  `json:"hotelId"`
	Plat float64 `json:"lat"`
	Plon float64 `json:"lon"`
}

// Implement Point interface
func (p *point) Lat() float64 { return p.Plat }
func (p *point) Lon() float64 { return p.Plon }
func (p *point) Id() string   { return p.Pid }

// New returns a new server
func New() *Geo {
	return &Geo{
		points: loadPoints("data/geo.json"),
	}
}

// Server implements the geo service
type Geo struct {
	points []*point
}

func (s *Geo) Run(port int) {
	http.HandleFunc("/nearby", s.Nearby)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// Haversine formula to calculate the distance between two points on the earth
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convert latitude and longitude from degrees to radians
	lat1Rad, lon1Rad := lat1*math.Pi/180, lon1*math.Pi/180
	lat2Rad, lon2Rad := lat2*math.Pi/180, lon2*math.Pi/180

	// Haversine formula
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

func (s *Geo) Nearby(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, "Invalid latitude", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, "Invalid longitude", http.StatusBadRequest)
		return
	}

	var nearbyHotels []string
	for _, point := range s.points {
		distance := haversineDistance(lat, lon, point.Plat, point.Plon)
		if distance <= maxSearchRadius {
			nearbyHotels = append(nearbyHotels, point.Pid)
		}
	}

	res := struct {
		HotelIds []string `json:"hotelIds"`
	}{
		HotelIds: nearbyHotels,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func loadPoints(path string) []*point {
	var (
		file   = data.MustAsset(path)
		points []*point
	)

	if err := json.Unmarshal(file, &points); err != nil {
		log.Fatalf("Failed to load hotels: %v", err)
	}

	return points
}

func main() {
	var (
		port = flag.Int("port", 8080, "The service port")
	)
	flag.Parse()

	srv := New()
	srv.Run(*port)
}
