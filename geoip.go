package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const freeGeoURL = "https://freegeoip.net/json/"

type freeGeoIP struct {
	IP          string  `json:"ip"`
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	RegionCode  string  `json:"region_code"`
	RegionName  string  `json:"region_name"`
	City        string  `json:"city"`
	ZipCode     string  `json:"zip_code"`
	TimeZone    string  `json:"time_zone"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	MetroCode   int     `json:"metro_code"`
}

func (f *freeGeoIP) Location() (*Location, error) {
	return &Location{IP: f.IP, Latitude: f.Latitude, Longitude: f.Longitude}, nil
}

var (
	cache = map[string]*freeGeoIP{}
)

type Location struct {
	IP        string  `json:"ip"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (l *Location) String() string {
	if l == nil {
		return ""
	}
	return fmt.Sprintf("%f,%f", l.Latitude, l.Longitude)
}

func LocationFrom(ip string) (*Location, error) {
	if g, ok := cache[ip]; ok {
		return g.Location()
	}
	g, err := getLocation(ip)
	if err != nil {
		return nil, err
	}
	cache[ip] = g
	return g.Location()
}

func getLocation(ip string) (*freeGeoIP, error) {
	u, err := url.Parse(freeGeoURL + ip)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	switch res.StatusCode {
	case http.StatusOK:
	default:
		return nil, fmt.Errorf("unexpected response code: %s", res.Status)
	}

	if res.Body == nil {
		return nil, errors.New("unexpectedly empty response")
	}
	defer res.Body.Close()

	g := &freeGeoIP{}
	if err = json.NewDecoder(res.Body).Decode(g); err != nil {
		return nil, err
	}
	return g, nil
}
