CREATE TABLE IF NOT EXISTS songs
(
    id SERIAL PRIMARY KEY,
    artist VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_artist ON songs (artist);
CREATE INDEX IF NOT EXISTS idx_title ON songs (title);

CREATE TABLE IF NOT EXISTS song_details
(
    id SERIAL PRIMARY KEY,
    song_id INT UNIQUE NOT NULL,
    release_date DATE NOT NULL,
    lyrics TEXT NOT NULL,
    FOREIGN KEY (song_id) REFERENCES songs (id) ON DELETE CASCADE
);