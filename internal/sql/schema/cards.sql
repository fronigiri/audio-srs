CREATE TABLE IF NOT EXISTS cards (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    audio_path  TEXT NOT NULL,
    interval    INTEGER NOT NULL DEFAULT 1,  -- days until next review
    ease_factor REAL NOT NULL DEFAULT 2.5,   -- SM-2 algorithm factor
    due_date    TEXT NOT NULL,               -- ISO 8601 date string
    created_at  TEXT NOT NULL
    FOREIGN KEY (deck_id) REFERENCES decks(id)
);
