ALTER TABLE song_details DROP CONSTRAINT IF EXISTS song_detail_song_id_fkey;

DROP TABLE IF EXISTS song_details;

DROP TABLE IF EXISTS songs;
