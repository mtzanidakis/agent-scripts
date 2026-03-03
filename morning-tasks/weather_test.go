package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var testResponse = weatherResponse{
	Daily: weatherDaily{
		Data: []dailyData{
			{
				Day:     "2026-02-24",
				Summary: "Sunny. Temperature 5/18 °C.",
			},
			{
				Day:     "2026-02-25",
				Summary: "Cloudy. Temperature 8/14 °C.",
			},
		},
	},
}

func init() {
	testResponse.Current.Summary = "Sunny"
	testResponse.Current.Temperature = 14
	testResponse.Current.CloudCover = 10
	testResponse.Current.Wind.Speed = 3
	testResponse.Current.Wind.Dir = "SW"
	testResponse.Current.Precipitation.Total = 0
	testResponse.Daily.Data[0].AllDay.TemperatureMin = 5
	testResponse.Daily.Data[0].AllDay.TemperatureMax = 18
	testResponse.Daily.Data[1].AllDay.TemperatureMin = 8
	testResponse.Daily.Data[1].AllDay.TemperatureMax = 14
}

func TestFormatWeather(t *testing.T) {
	var buf bytes.Buffer
	date := time.Now()
	formatWeather(&buf, &testResponse, date)
	out := buf.String()

	checks := []string{
		"=== Weather ===",
		"Now: Sunny, 14°C",
		"Wind: 3 km/h SW | Clouds: 10%",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}

	if strings.Contains(out, "Precipitation") {
		t.Error("should not show precipitation when total is 0")
	}
}

func TestFormatWeatherWithPrecipitation(t *testing.T) {
	data := testResponse
	data.Current.Precipitation.Total = 2.5
	data.Current.Precipitation.Type = "rain"

	var buf bytes.Buffer
	formatWeather(&buf, &data, time.Now())
	out := buf.String()

	if !strings.Contains(out, "Precipitation: 2.5 mm (rain)") {
		t.Errorf("expected precipitation line, got:\n%s", out)
	}
}

func TestFormatWeatherNonToday(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2026-02-25")

	var buf bytes.Buffer
	formatWeather(&buf, &testResponse, date)
	out := buf.String()

	if strings.Contains(out, "Now:") {
		t.Error("should not show current conditions for non-today date")
	}
	if !strings.Contains(out, "2026-02-25: 8–14°C") {
		t.Errorf("expected daily entry for 2026-02-25, got:\n%s", out)
	}
}

func TestFormatWeatherNoData(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2026-11-08")

	var buf bytes.Buffer
	formatWeather(&buf, &testResponse, date)
	out := buf.String()

	if !strings.Contains(out, "2026-11-08: no data") {
		t.Errorf("expected 'no data' for missing date, got:\n%s", out)
	}
}

func TestFetchWeather(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Query().Get("place_id") != "cholargos" {
			http.Error(w, "bad place_id", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testResponse)
	}))
	defer srv.Close()

	data, err := fetchWeather(srv.URL, "test-key", "cholargos")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Current.Summary != "Sunny" {
		t.Errorf("got summary %q, want Sunny", data.Current.Summary)
	}
	if data.Current.Temperature != 14 {
		t.Errorf("got temperature %v, want 14", data.Current.Temperature)
	}
}

func TestFetchWeatherBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := fetchWeather(srv.URL, "key", "nowhere")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "status 404") {
		t.Errorf("expected status 404 in error, got: %v", err)
	}
}
