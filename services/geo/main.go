package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hailocab/go-geoindex"
	"github.com/harlow/go-micro-services/geo/data"
	"log"
	"net/http"
	"strconv"
)

const (
	maxSearchRadius  = 10
	maxSearchResults = 1000000000
)

type point struct {
	Pid  string  `json:"hotelId"`
	Plat float64 `json:"lat"`
	Plon float64 `json:"lon"`
}

func (p *point) Lat() float64 { return p.Plat }
func (p *point) Lon() float64 { return p.Plon }
func (p *point) Id() string   { return p.Pid }

type Geo struct {
	geoIndex *geoindex.ClusteringIndex
}

func New() *Geo {
	return &Geo{
		geoIndex: newGeoIndex("data/geo.json"), // Adjust path accordingly if needed
	}
}

func (s *Geo) Run(port int) {
	http.HandleFunc("/nearby", s.Nearby)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
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

	points := s.getNearbyPoints(lat, lon)

	res := struct {
		HotelIds []string `json:"hotelIds"`
	}{}

	for _, p := range points {
		res.HotelIds = append(res.HotelIds, p.Id())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (s *Geo) getNearbyPoints(lat, lon float64) []geoindex.Point {
	center := &geoindex.GeoPoint{
		Pid:  "",
		Plat: lat,
		Plon: lon,
	}

	return s.geoIndex.KNearest(
		center,
		maxSearchResults,
		geoindex.Km(maxSearchRadius),
		func(p geoindex.Point) bool {
			return true
		},
	)
}

func newGeoIndex(path string) *geoindex.ClusteringIndex {
	var (
		file   = data.MustAsset(path)
		points []*point
	)

	// load geo points from json file
	if err := json.Unmarshal(file, &points); err != nil {
		log.Fatalf("Failed to load hotels: %v", err)
	}

	// add points to index
	index := geoindex.NewClusteringIndex()
	for _, point := range points {
		index.Add(point)
	}

	return index
}

func main() {
	var (
		port = flag.Int("port", 8082, "The service port")
	)
	flag.Parse()

	srv := New()
	srv.Run(*port)
}
