package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type weatherResponse struct {
	Current struct {
		Summary     string  `json:"summary"`
		Temperature float64 `json:"temperature"`
		CloudCover  int     `json:"cloud_cover"`
		Wind        struct {
			Speed float64 `json:"speed"`
			Dir   string  `json:"dir"`
		} `json:"wind"`
		Precipitation struct {
			Total float64 `json:"total"`
			Type  string  `json:"type"`
		} `json:"precipitation"`
	} `json:"current"`
	Daily weatherDaily `json:"daily"`
}

type weatherDaily struct {
	Data []dailyData `json:"data"`
}

type dailyData struct {
	Day     string `json:"day"`
	Summary string `json:"summary"`
	AllDay  struct {
		Temperature    float64 `json:"temperature"`
		TemperatureMin float64 `json:"temperature_min"`
		TemperatureMax float64 `json:"temperature_max"`
	} `json:"all_day"`
}

func fetchWeather(baseURL, apiKey, location string) (*weatherResponse, error) {
	url := fmt.Sprintf(
		"%s/api/v1/free/point?place_id=%s&sections=all&timezone=Europe/Athens&language=en&units=metric",
		baseURL, location,
	)

	client := &http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("weather: %w", err)
	}
	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("weather: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather: status %d", resp.StatusCode)
	}

	var data weatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("weather: %w", err)
	}
	return &data, nil
}

func formatWeather(w io.Writer, data *weatherResponse, date time.Time) {
	fmt.Fprintln(w, "=== Weather ===")

	dateStr := date.Format("2006-01-02")
	today := time.Now().Format("2006-01-02")
	isToday := dateStr == today

	if isToday {
		c := data.Current
		fmt.Fprintf(w, "Now: %s, %.0f°C\n", c.Summary, c.Temperature)
		fmt.Fprintf(w, "Wind: %.0f km/h %s | Clouds: %d%%\n", c.Wind.Speed, c.Wind.Dir, c.CloudCover)
		if c.Precipitation.Total > 0 {
			fmt.Fprintf(w, "Precipitation: %.1f mm (%s)\n", c.Precipitation.Total, c.Precipitation.Type)
		}
	}

	var found *dailyData
	for i := range data.Daily.Data {
		if data.Daily.Data[i].Day == dateStr {
			found = &data.Daily.Data[i]
			break
		}
	}

	if found != nil {
		label := dateStr
		if isToday {
			label = "Today"
		}
		fmt.Fprintf(w, "%s: %.0f–%.0f°C  %s\n",
			label, found.AllDay.TemperatureMin, found.AllDay.TemperatureMax, strings.ToLower(found.Summary))
	} else if !isToday {
		fmt.Fprintf(w, "%s: no data\n", dateStr)
	}
}
