-- Drop indexes
DROP INDEX IF EXISTS idx_series_characters_series_id;
DROP INDEX IF EXISTS idx_series_tags_series_id;
DROP INDEX IF EXISTS idx_series_genres_series_id;
DROP INDEX IF EXISTS idx_covers_series_id;
DROP INDEX IF EXISTS idx_urls_series_id;
DROP INDEX IF EXISTS idx_series_updated_at;
DROP INDEX IF EXISTS idx_series_year;
DROP INDEX IF EXISTS idx_series_status;
DROP INDEX IF EXISTS idx_series_name;

-- Drop tables in reverse order to handle foreign key constraints
DROP TABLE IF EXISTS synonyms;
DROP TABLE IF EXISTS urls;
DROP TABLE IF EXISTS series_characters;
DROP TABLE IF EXISTS characters;
DROP TABLE IF EXISTS series_tags;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS series_genres;
DROP TABLE IF EXISTS genres;
DROP TABLE IF EXISTS covers;
DROP TABLE IF EXISTS series;

-- Drop enum types
DROP TYPE IF EXISTS manga_format;
DROP TYPE IF EXISTS manga_status;

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";
