package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

func main() {
	port := flag.Int("port", 8080, "The service port")
	geoaddr := flag.String("geoaddr", "http://geo:8080/nearby", "Geo server addr")
	rateaddr := flag.String("rateaddr", "http://rate:8080/rates", "Rate server addr")
	flag.Parse()

	http.HandleFunc("/nearby", func(w http.ResponseWriter, r *http.Request) {
		lat := r.URL.Query().Get("lat")
		lon := r.URL.Query().Get("lon")
		inDate := r.URL.Query().Get("inDate")
		outDate := r.URL.Query().Get("outDate")

		nearby, err := queryGeo(*geoaddr, lat, lon)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rates, err := queryRate(*rateaddr, nearby.HotelIds, inDate, outDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var result struct {
			HotelIds []string `json:"hotelIds"`
		}
		for _, ratePlan := range rates.RatePlans {
			result.HotelIds = append(result.HotelIds, ratePlan.HotelId)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func queryGeo(addr, lat, lon string) (*struct {
	HotelIds []string `json:"hotelIds"`
}, error) {
	resp, err := http.Get(fmt.Sprintf("%s?lat=%s&lon=%s", addr, url.QueryEscape(lat), url.QueryEscape(lon)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		HotelIds []string `json:"hotelIds"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func queryRate(addr string, hotelIds []string, inDate, outDate string) (*struct {
	RatePlans []*RatePlan `json:"ratePlans"`
}, error) {
	resp, err := http.Get(fmt.Sprintf("%s?hotelIds=%s&inDate=%s&outDate=%s", addr, url.QueryEscape(joinString(hotelIds, ",")), url.QueryEscape(inDate), url.QueryEscape(outDate)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		RatePlans []*RatePlan `json:"ratePlans"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func joinString(s []string, sep string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return s[0]
	}
	var b bytes.Buffer
	b.WriteString(s[0])
	for _, sn := range s[1:] {
		b.WriteString(sep)
		b.WriteString(sn)
	}
	return b.String()
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
