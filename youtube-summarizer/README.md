# youtube-summarizer

Claude Code skill that summarizes YouTube videos from their transcript. Triggered by sending `summarize <youtube-url>` in a conversation.

## How it works

1. Fetches subtitles/captions using [yt-dlp](https://github.com/yt-dlp/yt-dlp) (installed automatically via `nix_add` if missing)
2. Cleans the VTT file (strips timing headers, deduplicates lines)
3. Produces a concise summary: bullet-point key points + 1-2 paragraph overview
4. Cleans up temp files

Prefers English subtitles when available. Falls back to auto-generated captions or any available language. If no captions exist at all, it lets you know.

## Installation

Copy or symlink `SKILL.md` into your Claude Code custom skills directory so it gets picked up as a slash command.

## Example

```
summarize https://www.youtube.com/watch?v=dQw4w9WgXcQ
```
