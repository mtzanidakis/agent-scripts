package main

import (
	"strings"
	"time"

	"github.com/mtzanidakis/agent-scripts/morning-tasks/namedays"
	miniflux "miniflux.app/v2/client"
)

// Result types for JSON output.

type weatherResult struct {
	Current *weatherCurrent `json:"current,omitempty"`
	Daily   *weatherDayInfo `json:"daily,omitempty"`
}

type weatherCurrent struct {
	Summary           string  `json:"summary"`
	Temperature       float64 `json:"temperature"`
	WindSpeed         float64 `json:"wind_speed"`
	WindDir           string  `json:"wind_dir"`
	CloudCover        int     `json:"cloud_cover"`
	PrecipitationMM   float64 `json:"precipitation_mm,omitempty"`
	PrecipitationType string  `json:"precipitation_type,omitempty"`
}

type weatherDayInfo struct {
	Date    string  `json:"date"`
	MinTemp float64 `json:"min_temp"`
	MaxTemp float64 `json:"max_temp"`
	Summary string  `json:"summary"`
}

type namedaysResult struct {
	Names []string `json:"names"`
}

type offerItem struct {
	Title string `json:"title"`
	URL   string `json:"url,omitempty"`
}

type offersResult struct {
	Items []offerItem `json:"items"`
}

type newsTopicItem struct {
	Title    string   `json:"title"`
	URL      string   `json:"url,omitempty"`
	Sources  []string `json:"sources"`
	Summary  string   `json:"summary,omitempty"`
	Category string   `json:"category"`
}

type newsResult struct {
	Topics []newsTopicItem `json:"topics"`
}

type taskResult struct {
	Weather  *weatherResult    `json:"weather,omitempty"`
	Namedays *namedaysResult   `json:"namedays,omitempty"`
	Offers   *offersResult     `json:"offers,omitempty"`
	News     *newsResult       `json:"news,omitempty"`
	Errors   map[string]string `json:"errors,omitempty"`
}

// Data extraction functions (separate from formatting).

func weatherData(data *weatherResponse, date time.Time) *weatherResult {
	res := &weatherResult{}
	dateStr := date.Format("2006-01-02")
	today := time.Now().Format("2006-01-02")
	isToday := dateStr == today

	if isToday {
		c := data.Current
		res.Current = &weatherCurrent{
			Summary:     c.Summary,
			Temperature: c.Temperature,
			WindSpeed:   c.Wind.Speed,
			WindDir:     c.Wind.Dir,
			CloudCover:  c.CloudCover,
		}
		if c.Precipitation.Total > 0 {
			res.Current.PrecipitationMM = c.Precipitation.Total
			res.Current.PrecipitationType = c.Precipitation.Type
		}
	}

	for _, d := range data.Daily.Data {
		if d.Day == dateStr {
			res.Daily = &weatherDayInfo{
				Date:    dateStr,
				MinTemp: d.AllDay.TemperatureMin,
				MaxTemp: d.AllDay.TemperatureMax,
				Summary: strings.ToLower(d.Summary),
			}
			break
		}
	}

	return res
}

func namedaysData(date time.Time) (*namedaysResult, error) {
	names, err := namedays.ForDate(date)
	if err != nil {
		return nil, err
	}
	if names == nil {
		names = []string{}
	}
	return &namedaysResult{Names: names}, nil
}

func offersData(entries []*miniflux.Entry) *offersResult {
	items := make([]offerItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, offerItem{Title: e.Title, URL: e.URL})
	}
	return &offersResult{Items: items}
}

func newsData(topics []hotTopic) *newsResult {
	items := make([]newsTopicItem, 0, len(topics))
	for _, t := range topics {
		items = append(items, newsTopicItem{
			Title:    t.title,
			URL:      t.url,
			Sources:  t.sources,
			Summary:  t.summary,
			Category: t.category,
		})
	}
	return &newsResult{Topics: items}
}
