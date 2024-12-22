-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum types
CREATE TYPE manga_status AS ENUM ('Unknown', 'Ongoing', 'Completed', 'Cancelled', 'Hiatus');
CREATE TYPE manga_format AS ENUM ('manga', 'novel', 'one_shot');

-- Create series table
CREATE TABLE series (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description_formatted TEXT,
    description_text TEXT,
    publisher VARCHAR(255),
    status manga_status DEFAULT 'Unknown',
    year INTEGER,
    total_chapters INTEGER,
    total_issues INTEGER,
    book_type VARCHAR(50),
    comic_image TEXT,
    comic_id INTEGER,
    publication_run VARCHAR(100),
    volumes INTEGER DEFAULT 0,
    chapters INTEGER DEFAULT 0,
    average_score INTEGER,
    popularity INTEGER,
    mean_score INTEGER,
    is_licensed BOOLEAN DEFAULT false,
    updated_at BIGINT,
    banner_image TEXT,
    format manga_format,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_name UNIQUE (name)
);

-- Create cover table
CREATE TABLE covers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    series_id UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    extra_large TEXT,
    large TEXT,
    medium TEXT,
    color VARCHAR(7),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create genres table
CREATE TABLE genres (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    CONSTRAINT unique_genre UNIQUE (name)
);

-- Create series_genres junction table
CREATE TABLE series_genres (
    series_id UUID REFERENCES series(id) ON DELETE CASCADE,
    genre_id UUID REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (series_id, genre_id)
);

-- Create tags table
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    CONSTRAINT unique_tag UNIQUE (name)
);

-- Create series_tags junction table
CREATE TABLE series_tags (
    series_id UUID REFERENCES series(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (series_id, tag_id)
);

-- Create characters table
CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    CONSTRAINT unique_character UNIQUE (name)
);

-- Create series_characters junction table
CREATE TABLE series_characters (
    series_id UUID REFERENCES series(id) ON DELETE CASCADE,
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    PRIMARY KEY (series_id, character_id)
);

-- Create urls table
CREATE TABLE urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    series_id UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create synonyms table
CREATE TABLE synonyms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    series_id UUID NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_series_name ON series(name);
CREATE INDEX idx_series_status ON series(status);
CREATE INDEX idx_series_year ON series(year);
CREATE INDEX idx_series_updated_at ON series(updated_at);
CREATE INDEX idx_urls_series_id ON urls(series_id);
CREATE INDEX idx_covers_series_id ON covers(series_id);
CREATE INDEX idx_series_genres_series_id ON series_genres(series_id);
CREATE INDEX idx_series_tags_series_id ON series_tags(series_id);
CREATE INDEX idx_series_characters_series_id ON series_characters(series_id);
