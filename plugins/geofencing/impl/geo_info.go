package impl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type geoFencingResponse struct {
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
}

const endpoint = "http://ip-api.com/json"

var client *http.Client = &http.Client{
	Timeout: time.Second * 1,
	Transport: &http.Transport{
		MaxIdleConns:    50,
		IdleConnTimeout: 30 * time.Second,
		WriteBufferSize: 1024,
		ReadBufferSize:  1024,
	},
}

func getGeoFencingData(ip string) (*geoFencingResponse, error) {
	url := fmt.Sprintf("%s/%s", endpoint, ip)
	resp, err := client.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	gfResp := geoFencingResponse{}

	if err := decoder.Decode(&gfResp); err != nil {
		return nil, err
	}

	return &gfResp, nil
}

func (g *geoFencingResponse) headerString() string {
	return fmt.Sprintf("lat=%v;lon=%v;country=%v;region=%v;city=%v", g.Lat, g.Lon, g.CountryCode, g.Region, g.City)
}
