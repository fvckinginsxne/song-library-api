ALTER TABLE song_detail DROP CONSTRAINT IF EXISTS song_detail_song_id_fkey;

DROP TABLE IF EXISTS song_detail;

DROP TABLE IF EXISTS songs;
