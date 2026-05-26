--reader/database Schema
CREATE TABLE IF NOT EXISTS libraries (
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS series (
    id INTEGER NOT NULL PRIMARY KEY,
    library_id INTEGER NOT NULL,
    name TEXT NOT NULL UNIQUE,
    book_count INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(library_id, name),
    FOREIGN KEY (library_id) REFERENCES libraries(id)
);

CREATE TABLE IF NOT EXISTS books (
    id INTEGER NOT NULL PRIMARY KEY,
    series_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    number INTEGER DEFAULT 0,
    publication TEXT NOT NULL,
    size_bytes INTEGER DEFAULT 0,
    file_hash TEXT,
    status TEXT DEFAULT 'READY',
    page_count INTEGER DEFAULT 0,
    first_issue INTEGER DEFAULT 0,
    last_issue INTEGER DEFAULT 0,
    release_date TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(name, series_id, publication),
    FOREIGN KEY (series_id) REFERENCES series(id)
);

CREATE TABLE IF NOT EXISTS episodes (
    id INTEGER NOT NULL PRIMARY KEY,
    book_id INTEGER NOT NULL,
    filename TEXT NOT NULL,
    issue_number INTEGER DEFAULT 0,
    title TEXT NOT NULL,
    part INTEGER DEFAULT 0,
    page_from INTEGER NOT NULL,
    page_to INTEGER NOT NULL,
    release_date TEXT,
    UNIQUE(book_id, issue_number, part),
    FOREIGN KEY (book_id) REFERENCES books(id)
);

CREATE TABLE IF NOT EXISTS creators (
    id INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS episode_creators (
    id INTEGER NOT NULL PRIMARY KEY,
    episode_id INTEGER NOT NULL,
    creator_id INTEGER NOT NULL,
    role TEXT NOT NULL,
    FOREIGN KEY (episode_id) REFERENCES episodes(id),
    FOREIGN KEY (creator_id) REFERENCES creators(id)
);

CREATE TABLE IF NOT EXISTS covers (
    id INTEGER NOT NULL PRIMARY KEY,
    issue_number INTEGER DEFAULT 0,
    publication TEXT,
    text TEXT,
    series_id INTEGER NOT NULL,
    artist TEXT,
    filename TEXT,
    created_at TEXT NOT NULL,
    UNIQUE (issue_number, publication, series_id),
    FOREIGN KEY (series_id) REFERENCES series(id)
);