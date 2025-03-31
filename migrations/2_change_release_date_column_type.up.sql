ALTER TABLE songs
ALTER COLUMN release_date TYPE varchar(20)
USING to_char(release_date, 'YYYY-MM-DD');