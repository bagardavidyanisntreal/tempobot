CREATE TABLE IF NOT EXISTS users
(
    id               SERIAL PRIMARY KEY,
    telegram_user_id BIGINT UNIQUE,
    username         TEXT,
    first_name       TEXT,
    role             TEXT      DEFAULT 'member',
    created_at       TIMESTAMP DEFAULT now()
);


CREATE TABLE IF NOT EXISTS events
(
    id          SERIAL PRIMARY KEY,
    title       TEXT,
    description TEXT,
    chat_id     BIGINT,
    message_id  INT,
    created_at  TIMESTAMP DEFAULT now()
);


CREATE TABLE IF NOT EXISTS participants
(
    id         SERIAL PRIMARY KEY,
    event_id   BIGINT,
    user_id    BIGINT,
    status     TEXT,
    updated_at TIMESTAMP DEFAULT now(),
    UNIQUE (event_id, user_id)
);