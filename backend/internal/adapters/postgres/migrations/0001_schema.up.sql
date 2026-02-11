CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,

    CONSTRAINT unique_username UNIQUE (username)
);