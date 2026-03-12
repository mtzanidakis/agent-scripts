# YouTube Summarizer Skill

When the user sends a message matching `summarize <url>` where the URL is a YouTube link:

## Steps

1. Install yt-dlp if not already available:
   - Use the `nix_add` tool with package name `yt-dlp`

2. Fetch the transcript using yt-dlp:
   ```bash
   yt-dlp --write-auto-sub --write-sub --sub-lang en --sub-format vtt \
     --skip-download --no-warnings \
     -o /tmp/yt-summary "URL"
   ```
   If `en` is unavailable, retry without `--sub-lang` to get whatever language is available.

3. Find and clean the .vtt file:
   ```bash
   # Strip VTT timing headers and dedup repeated lines
   grep -v '^WEBVTT' /tmp/yt-summary*.vtt \
     | grep -v '^[0-9]' \
     | grep -v '^$' \
     | grep -v '\-\->' \
     | awk '!seen[$0]++' \
     > /tmp/yt-transcript.txt
   ```

4. Read `/tmp/yt-transcript.txt` and produce a concise summary:
   - Key points as bullet list
   - 1-2 paragraph overview
   - Keep it focused, skip filler/ads

5. Clean up temp files after summarizing:
   ```bash
   rm -f /tmp/yt-summary* /tmp/yt-transcript.txt
   ```

## Fallback

- If yt-dlp cannot find subtitles, retry with `--write-auto-sub` only (auto-generated captions).
- If no captions exist at all, inform the user that the video has no available transcript.

## Example trigger

```
summarize https://www.youtube.com/watch?v=dQw4w9WgXcQ
```
