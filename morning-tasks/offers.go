package main

import (
	"fmt"
	"io"
	"time"

	miniflux "miniflux.app/v2/client"
)

const offersFeedTitle = "Lagonika.gr"

func findFeedByTitle(c *miniflux.Client, title string) (*miniflux.Feed, error) {
	feeds, err := c.Feeds()
	if err != nil {
		return nil, fmt.Errorf("offers: %w", err)
	}
	for _, f := range feeds {
		if f.Title == title {
			return f, nil
		}
	}
	return nil, fmt.Errorf("offers: feed %q not found", title)
}

func fetchOffers(c *miniflux.Client, feedID int64, since time.Time) ([]*miniflux.Entry, error) {
	result, err := c.Entries(&miniflux.Filter{
		FeedID:         feedID,
		Status:         miniflux.EntryStatusUnread,
		PublishedAfter: since.Unix(),
		Limit:          1000,
	})
	if err != nil {
		return nil, fmt.Errorf("offers: %w", err)
	}
	return result.Entries, nil
}

func markAsRead(c *miniflux.Client, entries []*miniflux.Entry) error {
	if len(entries) == 0 {
		return nil
	}
	ids := make([]int64, len(entries))
	for i, e := range entries {
		ids[i] = e.ID
	}
	if err := c.UpdateEntries(ids, miniflux.EntryStatusRead); err != nil {
		return fmt.Errorf("offers: mark read: %w", err)
	}
	return nil
}

func formatOffers(w io.Writer, entries []*miniflux.Entry) {
	fmt.Fprintln(w, "=== Offers ===")
	if len(entries) == 0 {
		fmt.Fprintln(w, "no new offers")
		return
	}
	for _, e := range entries {
		fmt.Fprintf(w, "- %s\n", e.Title)
		if e.URL != "" {
			fmt.Fprintf(w, "  %s\n", e.URL)
		}
	}
}
