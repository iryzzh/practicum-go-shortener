CREATE TABLE IF NOT EXISTS users
(
    user_id serial PRIMARY KEY,
    uuid    VARCHAR(50) UNIQUE NOT NULL
);