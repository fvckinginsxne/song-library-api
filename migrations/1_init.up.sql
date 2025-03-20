CREATE TABLE IF NOT EXISTS songs
(
    id SERIAL PRIMARY KEY,
    performer VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_performer ON songs (performer);
CREATE INDEX IF NOT EXISTS idx_title ON songs (title);

CREATE TABLE IF NOT EXISTS song_datail
(
    id SERIAL PRIMARY KEY,
    song_id INT UNIQUE NOT NULL,
    release_date DATE NOT NULL,
    text TEXT NOT NULL,
    link VARCHAR(512) NOT NULL,
    FOREIGN KEY (song_id) REFERENCES songs(id) ON DELETE CASCADE
);