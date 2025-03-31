CREATE INDEX idx_songs_artist_lower ON songs (LOWER(artist));
CREATE INDEX idx_songs_artist_title ON songs (LOWER(title));
