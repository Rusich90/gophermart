CREATE TABLE users (
    id         UUID PRIMARY KEY,
    login      VARCHAR(255) UNIQUE NOT NULL,
    password   VARCHAR(255)        NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);