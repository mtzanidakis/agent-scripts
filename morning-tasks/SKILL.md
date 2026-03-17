# Morning Routine

This skill runs as a scheduled daily task. It greets the user, runs `morning-tasks -json`, and sends the formatted output as a message.

## Steps

1. Greet the user with "Good morning!" and today's date.
2. Run `morning-tasks -json` from PATH. This outputs structured JSON for easy parsing.
3. Parse the JSON output and format it into a friendly message. Use the structured data to present each section (Weather, Namedays, Offers, News) clearly.
4. If the `errors` field is present, mention the errors but continue with available sections.
5. Send the formatted result as a single message to the user.

## JSON output format

```json
{
  "weather": {
    "current": {
      "summary": "Sunny",
      "temperature": 14,
      "wind_speed": 3,
      "wind_dir": "SW",
      "cloud_cover": 10
    },
    "daily": {
      "date": "2026-02-24",
      "min_temp": 5,
      "max_temp": 18,
      "summary": "sunny"
    }
  },
  "namedays": {
    "names": ["Name1", "Name2"]
  },
  "offers": {
    "items": [
      { "title": "50% off electronics", "url": "https://example.com/deal1" }
    ]
  },
  "news": {
    "topics": [
      {
        "title": "Breaking news headline",
        "url": "https://example.com/article",
        "sources": ["Source1", "Source2"],
        "summary": "Brief summary of the story.",
        "category": "news"
      }
    ]
  },
  "errors": {
    "weather": "METEOSOURCE_API_KEY not set"
  }
}
```

Sections that were not requested or returned no data may be omitted. The `errors` field is only present when there are errors.

## Notes

- If `morning-tasks` fails or a section errors, include the error in the message and continue with the rest.
- The binary expects `METEOSOURCE_API_KEY`, `MINIFLUX_API_URL`, and `MINIFLUX_API_KEY` environment variables to be set.
- Use `-task` to run specific tasks and `-date=YYYY-MM-DD` to override the date.
