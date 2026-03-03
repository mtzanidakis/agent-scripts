package namedays

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

//go:embed recurring.json relative_to_easter.json
var dataFS embed.FS

type recurringFile struct {
	Data []struct {
		Names []string `json:"names"`
		Date  string   `json:"date"`
	} `json:"data"`
}

type easterFile struct {
	Special []struct {
		ToEaster   int      `json:"toEaster"`
		Main       string   `json:"main"`
		Variations []string `json:"variations"`
	} `json:"special"`
}

// OrthodoxEaster calculates the date of Orthodox Easter for the given year.
// This is a Go port of alexstyl/Greek-namedays OrthodoxEasterCalculator.java.
func OrthodoxEaster(year int) time.Time {
	a := year % 4
	b := year % 7
	c := year % 19
	d := (19*c + 15) % 30
	e := (2*a + 4*b - d + 34) % 7

	// Java Calendar months are 0-indexed, so month here is 1-indexed
	// (the formula yields 3 for March or 4 for April in 0-indexed Java,
	// which maps to 4 or 5 in Go's 1-indexed time.Month).
	month := (d + e + 114) / 31
	day := ((d+e+144)%31 + 1) + 1

	// the formula computes Julian calendar Easter; add 13 days for Gregorian
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	t = t.AddDate(0, 0, 13)
	return t
}

// ForDate returns all nameday names for the given date.
func ForDate(date time.Time) ([]string, error) {
	recurring, err := loadRecurring()
	if err != nil {
		return nil, err
	}

	easter, err := loadEaster()
	if err != nil {
		return nil, err
	}

	dateKey := fmt.Sprintf("%02d/%02d", date.Day(), date.Month())
	var names []string

	for _, entry := range recurring.Data {
		if entry.Date == dateKey {
			names = append(names, entry.Names...)
		}
	}

	easterDate := OrthodoxEaster(date.Year())
	offset := int(date.Sub(easterDate).Hours() / 24)

	for _, entry := range easter.Special {
		if entry.ToEaster == offset && len(entry.Variations) > 0 {
			names = append(names, entry.Variations...)
		}
	}

	return names, nil
}

// Format returns a formatted nameday string for the given date.
func Format(date time.Time) (string, error) {
	names, err := ForDate(date)
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", nil
	}
	return "=== Namedays ===\nγιορτάζουν σήμερα: " + strings.Join(names, ", "), nil
}

func loadRecurring() (*recurringFile, error) {
	data, err := dataFS.ReadFile("recurring.json")
	if err != nil {
		return nil, fmt.Errorf("namedays: %w", err)
	}
	var f recurringFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("namedays: %w", err)
	}
	return &f, nil
}

func loadEaster() (*easterFile, error) {
	data, err := dataFS.ReadFile("relative_to_easter.json")
	if err != nil {
		return nil, fmt.Errorf("namedays: %w", err)
	}
	var f easterFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("namedays: %w", err)
	}
	return &f, nil
}
