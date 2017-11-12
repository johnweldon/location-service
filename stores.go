package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const storesURL = "https://maps.googleapis.com/maps/api/place/nearbysearch/json"

type Pos struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

func (p *Pos) String() string { return fmt.Sprintf("%f,%f", p.Latitude, p.Longitude) }

func ParsePos(pos string) (*Pos, error) {
	var lat, lon float64
	if _, err := fmt.Sscanf(pos, "%f,%f", &lat, &lon); err != nil {
		return nil, err
	}
	return &Pos{Latitude: lat, Longitude: lon}, nil
}

type Store struct {
	Name     string `json:"name,omitempty"`
	Pos      *Pos   `json:"pos,omitempty"`
	Distance string `json:"distance,omitempty"`
}

var (
	stores    = map[string][]Store{}
	key       string
	storetype string
	allstores []string
)

func init() {
	key = os.Getenv("GOOGLE_LOCATION_API_KEY")
	storetype = os.Getenv("STORE_TYPE")
	allstores = strings.Split(os.Getenv("ALL_STORES"), ",")
}

func StoresNear(loc string) ([]Store, error) {
	if s, ok := stores[loc]; ok {
		return s, nil
	}
	s, err := getStores(allstores, loc)
	if err != nil {
		return nil, err
	}
	stores[loc] = s
	return s, nil
}

func getStores(stores []string, loc string) ([]Store, error) {
	pos, err := ParsePos(loc)
	if err != nil {
		return nil, err
	}
	var results []Store
	for _, store := range stores {
		gs, err := getStore(store, pos)
		if err != nil {
			return nil, err
		}
		results = append(results, gs.Stores()...)
	}
	return results, nil
}

func getStore(store string, pos *Pos) (*gStores, error) {
	u, err := url.Parse(storesURL)
	if err != nil {
		return nil, err
	}

	vals, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, err
	}
	vals.Set("key", key)
	vals.Set("location", pos.String())
	vals.Set("keyword", store)
	vals.Set("type", storetype)
	vals.Set("rankby", "distance")
	u.RawQuery = vals.Encode()

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

	g := &gStores{}
	if err = json.NewDecoder(res.Body).Decode(g); err != nil {
		return nil, err
	}
	return g, nil
}

type gStores struct {
	Results []gResult `json:"results,omitempty"`
	Status  string    `json:"status"`
}

func (g *gStores) Stores() []Store {
	var stores []Store
	for _, res := range g.Results {
		stores = append(stores, Store{
			Name: res.Name,
			Pos: &Pos{
				Latitude:  res.Geometry.Location.Latitude,
				Longitude: res.Geometry.Location.Longitude}})
	}
	return stores
}

type gResult struct {
	Name     string `json:"name"`
	Vicinity string `json:"vicinity"`
	Geometry struct {
		Location struct {
			Latitude  float64 `json:"lat"`
			Longitude float64 `json:"lng"`
		} `json:"location"`
	} `json:"geometry"`
}
