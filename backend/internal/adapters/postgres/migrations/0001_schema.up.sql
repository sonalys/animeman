CREATE TABLE users (
    id UUID NOT NULL,
    username TEXT PRIMARY KEY,
    password_hash TEXT NOT NULL
);

CREATE TYPE auth_type AS ENUM ('userPassword', 'apiKey');

CREATE TABLE authentications (
    id          UUID PRIMARY KEY,
    type        auth_type NOT NULL,
    credentials JSONB NOT NULL,

    CONSTRAINT valid_auth_data CHECK (
        (type = 'userPassword' AND 
            credentials ? 'username' AND 
            credentials ? 'password') 
        OR
        (type = 'apiKey' AND 
            credentials ? 'apiKey')
    )
);

CREATE TYPE indexer_client_type AS ENUM ('prowlarr');

CREATE TABLE indexer_clients (
    id          UUID PRIMARY KEY,
    owner_id    UUID NOT NULL,
    address     TEXT NOT NULL,
    type        indexer_client_type NOT NULL,
    auth_id     UUID NOT NULL,

    CONSTRAINT fk_indexer_auth 
        FOREIGN KEY (auth_id) 
        REFERENCES authentications(id) 
        ON DELETE CASCADE
);

CREATE TYPE transfer_client_type AS ENUM ('qBittorrent');

CREATE TABLE transfer_clients (
    id              UUID PRIMARY KEY,
    owner_id        UUID NOT NULL,
    address         TEXT NOT NULL,
    type            transfer_client_type NOT NULL,
    auth_id         UUID NOT NULL,

    CONSTRAINT fk_client_auth 
        FOREIGN KEY (auth_id) 
        REFERENCES authentications(id) 
        ON DELETE CASCADE
);

CREATE TABLE collections (
    id          UUID PRIMARY KEY,
    owner_id    UUID NOT NULL,
    name        TEXT NOT NULL,
    base_path   TEXT NOT NULL,
    tags        TEXT[],
    
    monitored   BOOLEAN NOT NULL,
    created_at  TIMESTAMPTZ,


    CONSTRAINT valid_path CHECK (length(base_path) > 0)
);

CREATE INDEX idx_collections_tags ON collections USING GIN (tags);
CREATE INDEX idx_collections_owner ON collections(owner_id);

CREATE TYPE monitoring_status AS ENUM ('unknown', 'all', 'future', 'missing', 'existing', 'firstSeason', 'latestSeason', 'none');
CREATE TYPE resolution AS ENUM ('unknown', '480p', '720p', '1080p', '2160p');
CREATE TYPE video_codec AS ENUM ('unknown', 'x264', 'x265', 'av1');
CREATE TYPE audio_codec AS ENUM ('unknown', 'aac','opus','flac','mp3','ac3','dts','truehd');

CREATE TABLE quality_profiles (
    id                      UUID PRIMARY KEY,
    name                    TEXT NOT NULL,
    min_resolution          resolution NOT NULL,
    max_resolution          resolution NOT NULL,
    codec_preference        video_codec[] NOT NULL,
    release_group_preference TEXT[] NOT NULL
);

CREATE TABLE media (
    id                 UUID PRIMARY KEY,
    collection_id      UUID NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    quality_profile_id UUID NOT NULL REFERENCES quality_profiles(id),
    

    titles             JSONB NOT NULL DEFAULT '[]',
    
    monitoring_status  monitoring_status NOT NULL,
    monitored_since    TIMESTAMPTZ NOT NULL,
    

    genres             TEXT[],
    airing_started_at  TIMESTAMPTZ,
    airing_ended_at    TIMESTAMPTZ,
    
    created_at         TIMESTAMPTZ NOT NULL,

    CONSTRAINT valid_title CHECK (
        jsonb_typeof(titles) = 'array' AND
        jsonb_array_length(titles) > 0 AND
        (
            SELECT bool_and(
                obj ? 'value' AND 
                obj ? 'language' AND 
                obj ? 'type'
            )
            FROM jsonb_array_elements(titles) AS obj
        )
    )
);

CREATE INDEX idx_media_titles ON media USING GIN (titles);

ALTER TABLE media ADD COLUMN titles_search_vector TEXT 
GENERATED ALWAYS AS (
    (SELECT string_agg(obj->>'value', ' ') FROM jsonb_array_elements(titles) AS obj)
) STORED;

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_media_titles_fuzzy ON media USING gin (titles_search_vector gin_trgm_ops);

CREATE TYPE media_type AS ENUM ('unknown', 'tv', 'movie', 'ova', 'special');
CREATE TYPE file_source AS ENUM ('unknown', 'tv', 'web', 'dvd', 'br');

CREATE TABLE seasons (
    id              UUID PRIMARY KEY,
    media_id        UUID NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    number          INTEGER NOT NULL,
    airing_status   TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}',
    UNIQUE(media_id, number)
);

CREATE TABLE episodes (
    id              UUID PRIMARY KEY,
    season_id       UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    media_id        UUID NOT NULL REFERENCES media(id) ON DELETE CASCADE,

    type            media_type NOT NULL,
    number          TEXT NOT NULL,
    titles          JSONB NOT NULL DEFAULT '[]',
    airing_date     TIMESTAMPTZ,

    UNIQUE(season_id, number)
);

CREATE TYPE subtitle_format AS ENUM ('unknown', 'srt', 'ass', 'ssa', 'pgs', 'vobsub');
CREATE TYPE hash_algorithm AS ENUM ('md5', 'sha1', 'sha256', 'crc32', 'ed2k');

CREATE TABLE collection_files (
    id              UUID PRIMARY KEY,
    episode_id      UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    season_id       UUID NOT NULL REFERENCES seasons(id),
    media_id        UUID NOT NULL REFERENCES media(id),

    relative_path   TEXT NOT NULL,
    size_bytes      BIGINT NOT NULL,
    release_group   TEXT,
    version         INTEGER NOT NULL DEFAULT 1,
    source          file_source NOT NULL,
    

    video_info       JSONB NOT NULL,
    audio_streams    JSONB NOT NULL DEFAULT '[]',
    subtitle_streams JSONB NOT NULL DEFAULT '[]',
    chapters         JSONB NOT NULL DEFAULT '[]',
    hashes           JSONB NOT NULL DEFAULT '[]',
    
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT enforce_video_metadata CHECK (
        jsonb_typeof(video_info) = 'object' AND
        (obj->>'codec')::video_codec IS NOT NULL AND
        video_info ? 'resolution' AND
        (video_info->>'bit_depth')::numeric >= 0 AND
        (video_info->>'width')::numeric > 0 AND
        (video_info->>'height')::numeric > 0
    ),

    CONSTRAINT enforce_audio_metadata CHECK (
        jsonb_typeof(audio_streams) = 'array' AND
        jsonb_array_length(audio_streams) > 0 AND
        (
            SELECT bool_and(
                obj ? 'language' AND
                (obj->>'codec')::audio_codec IS NOT NULL AND
                (obj->>'channels')::numeric >= 1.0
            )
            FROM jsonb_array_elements(audio_streams) AS obj
        )
    ),

    CONSTRAINT enforce_subtitle_metadata CHECK (
        jsonb_typeof(subtitle_streams) = 'array' AND
        (
            SELECT bool_and(
                obj ? 'language' AND 
                (obj->>'format')::subtitle_format IS NOT NULL
            )
            FROM jsonb_array_elements(subtitle_streams) AS obj
        )
    ),

    CONSTRAINT enforce_hashes_metadata CHECK (
        jsonb_typeof(hashes) = 'array' AND
        (
            SELECT bool_and(
                (obj->>'algorithm')::hash_algorithm IS NOT NULL AND
                obj ? 'value' AND
                length(obj->>'value') > 0
            )
            FROM jsonb_array_elements(hashes) AS obj
        )
    ),

    CONSTRAINT enforce_chapters_metadata CHECK (
        jsonb_typeof(chapters) = 'array' AND
        (
            SELECT bool_and(
                obj ? 'title' AND 
                (obj->>'startTime')::numeric >= 0 AND
                (obj->>'endTime')::numeric >= (obj->>'startTime')::numeric
            )
            FROM jsonb_array_elements(chapters) AS obj
        )
    )
);

CREATE INDEX idx_files_hashes ON collection_files USING GIN (hashes);