CREATE TABLE users (
    id UUID NOT NULL,
    username TEXT PRIMARY KEY,
    password_hash TEXT NOT NULL
);

CREATE TYPE anime_list_source AS ENUM ('mal', 'anilist');

CREATE TABLE anime_lists (
    id UUID NOT NULL,
    owner_id UUID NOT NULL,
    remote_username TEXT NOT NULL,
    source anime_list_source NOT NULL,

    PRIMARY KEY (source, remote_username),

    CONSTRAINT fk_anime_list_owner
        FOREIGN KEY (owner_id)
        REFERENCES users(id)
);

CREATE TYPE torrent_client_source AS ENUM ('qbittorrent');

CREATE TABLE torrent_clients (
    id UUID NOT NULL,
    owner_id UUID NOT NULL,
    source torrent_client_source NOT NULL,
    host TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    password TEXT NOT NULL,

    CONSTRAINT fk_torrent_client_owner
        FOREIGN KEY (owner_id)
        REFERENCES users(id)
);

CREATE TABLE import_configurations (
    id UUID NOT NULL,
    owner_id UUID NOT NULL,
    anime_list_id UUID NOT NULL,
    torrent_client_id UUID NOT NULL,
    prowlarr_configuration_id UUID NOT NULL,

    PRIMARY KEY (anime_list_id, torrent_client_id),

    CONSTRAINT fk_import_configuration_owner
        FOREIGN KEY (owner_id)
        REFERENCES users(id),

    CONSTRAINT fk_import_configurations_anime_list
        FOREIGN KEY (anime_list_id)
        REFERENCES anime_lists(id),

    CONSTRAINT fk_import_configurations_torrent_client
        FOREIGN KEY (torrent_client_id)
        REFERENCES torrent_clients(id),

    CONSTRAINT fk_import_configurations_prowlarr_configuration
        FOREIGN KEY (prowlarr_configuration_id)
        REFERENCES prowlarr_configurations(id)
);

CREATE TABLE prowlarr_configurations (
    id UUID NOT NULL,
    owner_id UUID NOT NULL,
    host TEXT PRIMARY KEY,
    api_key TEXT NOT NULL,

    CONSTRAINT fk_prowlarr_configuration_owner
        FOREIGN KEY (owner_id)
        REFERENCES users(id)
);