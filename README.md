# Tripl — Travel Assistant bot

A Telegram bot for planning trips. Keep your tickets, hotel bookings and events in one place, group them into trips, and get a reminder the day before. The interface is available in Russian and English.

## Features

- **Tickets, bookings and events.** Add them by typing the fields in a short fixed format, or by uploading a PDF (ticket, booking confirmation, event ticket) and letting the bot extract the details.
- **Trips.** Attach items to a trip and see them together, with expense totals per currency and export to CSV or XLSX.
- **Lists and cards.** Current and archived items with pagination. Open any item with its command (`/tt12`, `/bk3`, `/ev7`, `/tr2`), then edit any field, delete it or attach it to a trip.
- **Calendar export.** Download any item as an `.ics` file. For tickets you pick the departure and arrival time zones.
- **Daily reminders.** At 09:00 Moscow time the bot sends tomorrow's departures, check-ins and events.

Limits per user: 50 items of each type, 10 PDF uploads per day, PDFs up to 10 MB (only the first 10 pages are read).

### How PDF import works

1. The file is downloaded from Telegram and checked for type and size.
2. Text is extracted with `pdftotext` in a separate process with a timeout and an output limit.
3. A self-hosted LLM (Qwen3 on a llama.cpp server, OpenAI-compatible API in JSON mode) turns the text into structured records. A single ticket PDF can produce up to three journey segments.
4. Costs and currencies are normalized, and the user gets the saved items to review.

Uploads are processed in the background by a small worker queue, so the bot stays responsive.

## Tech stack

Go 1.26 · [go-telegram-bot-api v5](https://github.com/go-telegram-bot-api/telegram-bot-api) · PostgreSQL 17 ([pgx v5](https://github.com/jackc/pgx)) · [llama.cpp](https://github.com/ggml-org/llama.cpp) server · poppler-utils (`pdftotext`) · [excelize](https://github.com/xuri/excelize) · Docker Compose

## Project layout

| Path | Contents |
|---|---|
| `tripl.go` | Entry point: wiring, background workers, graceful shutdown |
| `bot/` | Update loop, message routing, parsing of typed fields |
| `handler/` | Lists, cards, editing, trips, calendar export, PDF queue |
| `service/` | Business logic, per-user state, daily upload quota |
| `internal/` | PostgreSQL repositories and schema |
| `document/` | PDF download and text extraction, ICS, CSV and XLSX generation |
| `ai/` | LLM client and prompts |
| `scheduler/` | Daily reminders |
| `lang/` | Russian and English texts |
| `common/` | Models, errors, limits, helpers |
