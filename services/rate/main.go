package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/harlow/go-micro-services/rate/data"
	"log"
	"net/http"
	"strings"
)

type Request struct {
	HotelIds []string `json:"hotelIds"`
	InDate   string   `json:"inDate"`
	OutDate  string   `json:"outDate"`
}

type Result struct {
	RatePlans []*RatePlan `json:"ratePlans"`
}

type RatePlan struct {
	HotelId         string   `json:"hotelId"`
	Code            string   `json:"code"`
	InDate          string   `json:"inDate"`
	OutDate         string   `json:"outDate"`
	RoomTypeDetails RoomType `json:"roomType"`
}

type RoomType struct {
	BookableRate       float64 `json:"bookableRate"`
	TotalRate          float64 `json:"totalRate"`
	TotalRateInclusive float64 `json:"totalRateInclusive"`
	Code               string  `json:"code"`
	Currency           string  `json:"currency"`
	RoomDescription    string  `json:"roomDescription"`
}

type stay struct {
	HotelID string
	InDate  string
	OutDate string
}

var rateTable map[stay]*RatePlan

func main() {
	port := flag.Int("port", 8080, "The service port")
	flag.Parse()

	rateTable = loadRateTable("data/inventory.json")

	http.HandleFunc("/rates", getRatesHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func getRatesHandler(w http.ResponseWriter, r *http.Request) {
	hotelIdsStr := r.URL.Query().Get("hotelIds")
	inDate := r.URL.Query().Get("inDate")
	outDate := r.URL.Query().Get("outDate")

	hotelIds := strings.Split(hotelIdsStr, ",")

	req := Request{
		HotelIds: hotelIds,
		InDate:   inDate,
		OutDate:  outDate,
	}

	var res Result
	for _, hotelID := range req.HotelIds {
		s := stay{
			HotelID: hotelID,
			InDate:  req.InDate,
			OutDate: req.OutDate,
		}
		if rate, ok := rateTable[s]; ok {
			res.RatePlans = append(res.RatePlans, rate)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func loadRateTable(path string) map[stay]*RatePlan {
	file := data.MustAsset(path)

	var rates []*RatePlan
	if err := json.Unmarshal(file, &rates); err != nil {
		log.Fatalf("Failed to load json: %v", err)
	}

	rateTable := make(map[stay]*RatePlan)
	for _, ratePlan := range rates {
		stay := stay{
			HotelID: ratePlan.HotelId,
			InDate:  ratePlan.InDate,
			OutDate: ratePlan.OutDate,
		}
		rateTable[stay] = ratePlan
	}

	return rateTable
}
