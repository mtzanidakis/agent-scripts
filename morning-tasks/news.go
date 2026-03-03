package main

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	miniflux "miniflux.app/v2/client"
)

var gossipKeywords = []string{
	"γάμος", "γάμο", "διαζύγιο", "χώρισε", "χωρισμός", "ζευγάρι",
	"παντρεύτηκε", "σχέση", "ερωτευμέν", "κουτσομπολ", "σκάνδαλο ερωτικ",
	"celebrity", "wedding", "divorce",
}

var commonWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
	"over": true, "with": true, "by": true, "from": true, "as": true, "is": true, "was": true,
	"are": true, "be": true, "been": true, "have": true, "has": true, "had": true,
	"do": true, "does": true, "did": true, "will": true, "would": true, "could": true,
	"should": true, "may": true, "might": true, "can": true, "must": true,
	"shall": true, "being": true, "this": true, "that": true, "these": true,
	"those": true, "i": true, "you": true, "he": true, "she": true, "it": true,
	"we": true, "they": true, "what": true, "which": true, "who": true,
	"when": true, "where": true, "why": true, "how": true,
	"και": true, "το": true, "τη": true, "την": true, "τον": true,
	"της": true, "του": true, "των": true, "ο": true, "η": true,
	"στο": true, "στη": true, "στην": true, "από": true, "για": true,
	"με": true, "σε": true, "ή": true, "ότι": true, "που": true,
	"είναι": true, "να": true, "δεν": true, "θα": true, "ένα": true,
	"μια": true, "ένας": true, "μία": true,
}

var targetCategories = map[string]bool{
	"news": true, "news nerdy": true, "finance": true, "crypto": true,
}

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)
var spacesRe = regexp.MustCompile(`\s+`)
var wordRe = regexp.MustCompile(`[\p{L}\p{N}_]+`)
var sentenceRe = regexp.MustCompile(`[.!?]+\s+`)

type hotTopic struct {
	title    string
	summary  string
	sources  []string
	url      string
	category string
}

func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = spacesRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func isGossip(entry *miniflux.Entry) bool {
	title := strings.ToLower(entry.Title)
	content := strings.ToLower(entry.Content)
	text := title + " " + content

	count := 0
	for _, kw := range gossipKeywords {
		if strings.Contains(text, kw) {
			count++
		}
	}

	if entry.FeedID == 96 && count > 0 {
		return true
	}
	return count >= 2
}

func extractKeywords(title string) map[string]bool {
	words := wordRe.FindAllString(strings.ToLower(title), -1)
	kw := make(map[string]bool)
	for _, w := range words {
		if utf8.RuneCountInString(w) >= 4 && !commonWords[w] {
			kw[w] = true
		}
	}
	return kw
}

func areSimilar(title1, title2 string) bool {
	// Keyword Jaccard similarity
	kw1 := extractKeywords(title1)
	kw2 := extractKeywords(title2)
	if len(kw1) == 0 || len(kw2) == 0 {
		return false
	}

	intersection := 0
	for w := range kw1 {
		if kw2[w] {
			intersection++
		}
	}

	union := len(kw1)
	for w := range kw2 {
		if !kw1[w] {
			union++
		}
	}

	if union == 0 {
		return false
	}

	return float64(intersection)/float64(union) >= 0.4
}

func clusterEntries(entries []*miniflux.Entry) [][]*miniflux.Entry {
	clustered := make(map[int]bool)
	var clusters [][]*miniflux.Entry

	for i, e1 := range entries {
		if clustered[i] {
			continue
		}
		cluster := []*miniflux.Entry{e1}
		clustered[i] = true

		for j := i + 1; j < len(entries); j++ {
			if clustered[j] {
				continue
			}
			e2 := entries[j]
			if e1.FeedID == e2.FeedID {
				continue
			}
			if areSimilar(e1.Title, e2.Title) {
				cluster = append(cluster, e2)
				clustered[j] = true
			}
		}

		if len(cluster) >= 2 {
			clusters = append(clusters, cluster)
		}
	}

	return clusters
}

func bestSummary(cluster []*miniflux.Entry) string {
	var best string
	maxLen := 0
	for _, e := range cluster {
		c := stripHTML(e.Content)
		if len(c) > maxLen {
			maxLen = len(c)
			best = c
		}
	}
	if best == "" {
		return ""
	}

	parts := sentenceRe.Split(best, 4)
	if len(parts) > 3 {
		parts = parts[:3]
	}
	summary := strings.Join(parts, ". ")

	if utf8.RuneCountInString(summary) > 300 {
		runes := []rune(summary)
		summary = string(runes[:297]) + "..."
	}
	return summary
}

func buildTopics(clusters [][]*miniflux.Entry) []hotTopic {
	var topics []hotTopic
	for _, cluster := range clusters {
		// Best title: longest
		bestTitle := cluster[0].Title
		for _, e := range cluster[1:] {
			if utf8.RuneCountInString(e.Title) > utf8.RuneCountInString(bestTitle) {
				bestTitle = e.Title
			}
		}

		// Unique sources
		srcSet := make(map[string]bool)
		for _, e := range cluster {
			name := "Unknown"
			if e.Feed != nil {
				name = e.Feed.Title
			}
			srcSet[name] = true
		}
		sources := make([]string, 0, len(srcSet))
		for s := range srcSet {
			sources = append(sources, s)
		}
		sort.Strings(sources)

		// Best URL: entry with most content
		bestURL := cluster[0].URL
		maxLen := 0
		for _, e := range cluster {
			c := stripHTML(e.Content)
			if len(c) > maxLen {
				maxLen = len(c)
				bestURL = e.URL
			}
		}

		// Category
		cat := "unknown"
		if cluster[0].Feed != nil && cluster[0].Feed.Category != nil {
			cat = cluster[0].Feed.Category.Title
		}

		topics = append(topics, hotTopic{
			title:    bestTitle,
			summary:  bestSummary(cluster),
			sources:  sources,
			url:      bestURL,
			category: cat,
		})
	}

	// Sort by source count descending
	sort.Slice(topics, func(i, j int) bool {
		return len(topics[i].sources) > len(topics[j].sources)
	})

	return topics
}

func processEntries(entries []*miniflux.Entry, date time.Time) []hotTopic {
	dateStr := date.Format("2006-01-02")

	// Filter by date
	var recent []*miniflux.Entry
	for _, e := range entries {
		if e.Date.Format("2006-01-02") == dateStr {
			recent = append(recent, e)
		}
	}

	// Filter by target categories
	var categorized []*miniflux.Entry
	for _, e := range recent {
		if e.Feed != nil && e.Feed.Category != nil {
			if targetCategories[strings.ToLower(e.Feed.Category.Title)] {
				categorized = append(categorized, e)
			}
		}
	}

	// Filter gossip
	var filtered []*miniflux.Entry
	for _, e := range categorized {
		if !isGossip(e) {
			filtered = append(filtered, e)
		}
	}

	clusters := clusterEntries(filtered)
	return buildTopics(clusters)
}

func formatNews(w io.Writer, topics []hotTopic) {
	fmt.Fprintln(w, "=== News ===")
	if len(topics) == 0 {
		fmt.Fprintln(w, "no hot topics")
		return
	}
	for _, t := range topics {
		fmt.Fprintf(w, "- %s\n", t.title)
		if t.url != "" {
			fmt.Fprintf(w, "  %s\n", t.url)
		}
	}
}

func fetchNews(apiURL, apiKey string) ([]*miniflux.Entry, error) {
	c := miniflux.NewClient(apiURL, apiKey)
	result, err := c.Entries(&miniflux.Filter{
		Status: miniflux.EntryStatusUnread,
		Limit:  1000,
	})
	if err != nil {
		return nil, fmt.Errorf("news: %w", err)
	}
	return result.Entries, nil
}
