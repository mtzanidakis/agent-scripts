# Morning Routine

This skill runs as a scheduled daily task. It greets the user, runs `morning-tasks`, and sends the formatted output as a message.

## Steps

1. Greet the user with "Good morning!" and today's date.
2. Run `morning-tasks` from PATH.
3. Format the output: each section (Weather, Namedays, Offers, News) is already separated by `===` headers. Present them clearly, preserving the structure.
4. Send the formatted result as a single message to the user.

## Example output

```
Good morning! Today is Tuesday, February 24, 2026.

=== Weather ===
Now: Sunny, 14°C
Wind: 3 km/h SW | Clouds: 10%
Today: 5–18°C  sunny

=== Namedays ===
γιορτάζουν σήμερα: ...

=== Offers ===
- 50% off electronics
  https://example.com/deal1

=== News ===
- Breaking news headline
  https://example.com/article
```

## Notes

- If `morning-tasks` fails or a section errors, include the error in the message and continue with the rest.
- The binary expects `METEOSOURCE_API_KEY`, `MINIFLUX_API_URL`, and `MINIFLUX_API_KEY` environment variables to be set.
- Use `-task` to run specific tasks and `-date=YYYY-MM-DD` to override the date.
