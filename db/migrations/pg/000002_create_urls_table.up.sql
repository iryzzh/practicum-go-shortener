CREATE TABLE IF NOT EXISTS urls
(
    url_id       serial PRIMARY KEY,
    user_id      int default 0,
    original_url VARCHAR(255) UNIQUE NOT NULL,
    short_url    VARCHAR(50) UNIQUE  NOT NULL
);