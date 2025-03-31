ALTER TABLE songs
ALTER COLUMN release_date TYPE date
USING to_date(release_date, 'YYYY-MM-DD');