package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/manolis/morning-tasks/namedays"
	miniflux "miniflux.app/v2/client"
)

func weather(date time.Time) error {
	apiKey := os.Getenv("METEOSOURCE_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("METEOSOURCE_API_KEY not set")
	}

	location := os.Getenv("WEATHER_LOCATION")
	if location == "" {
		location = "cholargos"
	}

	data, err := fetchWeather("https://www.meteosource.com", apiKey, location)
	if err != nil {
		return err
	}

	formatWeather(os.Stdout, data, date)
	return nil
}

func news(date time.Time) error {
	apiURL := os.Getenv("MINIFLUX_API_URL")
	if apiURL == "" {
		apiURL = "https://feeder.mtzanidakis.com"
	}
	apiKey := os.Getenv("MINIFLUX_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("MINIFLUX_API_KEY not set")
	}

	entries, err := fetchNews(apiURL, apiKey)
	if err != nil {
		return err
	}

	topics := processEntries(entries, date)
	formatNews(os.Stdout, topics)
	return nil
}

func offers(date time.Time) error {
	apiURL := os.Getenv("MINIFLUX_API_URL")
	if apiURL == "" {
		apiURL = "https://feeder.mtzanidakis.com"
	}
	apiKey := os.Getenv("MINIFLUX_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("MINIFLUX_API_KEY not set")
	}

	c := miniflux.NewClient(apiURL, apiKey)

	feed, err := findFeedByTitle(c, offersFeedTitle)
	if err != nil {
		return err
	}

	since := date.Add(-24 * time.Hour)
	entries, err := fetchOffers(c, feed.ID, since)
	if err != nil {
		return err
	}

	formatOffers(os.Stdout, entries)

	if err := markAsRead(c, entries); err != nil {
		return err
	}
	return nil
}

func nameday(date time.Time) error {
	out, err := namedays.Format(date)
	if err != nil {
		return err
	}
	if out != "" {
		fmt.Println(out)
	}
	return nil
}

func main() {
	taskFlag := flag.String("task", "", "comma-separated task names to run (default: all)")
	dateFlag := flag.String("date", "", "date to use in YYYY-MM-DD format (default: today)")
	listFlag := flag.Bool("list", false, "list available tasks and exit")
	flag.Parse()

	tasks := []struct {
		name string
		fn   func(time.Time) error
	}{
		{"weather", weather},
		{"namedays", nameday},
		{"offers", offers},
		{"news", news},
	}

	if *listFlag {
		for _, t := range tasks {
			fmt.Println(t.name)
		}
		return
	}

	date := time.Now()
	if *dateFlag != "" {
		var err error
		date, err = time.Parse("2006-01-02", *dateFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid date %q: %v\n", *dateFlag, err)
			os.Exit(1)
		}
	}

	selected := make(map[string]bool)
	if *taskFlag != "" {
		for _, name := range strings.Split(*taskFlag, ",") {
			selected[strings.TrimSpace(name)] = true
		}
	}

	for _, t := range tasks {
		if len(selected) > 0 && !selected[t.name] {
			continue
		}
		if err := t.fn(date); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", t.name, err)
		}
	}
}
