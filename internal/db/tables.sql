CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    language VARCHAR(2) NOT NULL DEFAULT 'ru',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS trips (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    name VARCHAR(256),
    start_date DATE,
    end_date DATE,
    comment VARCHAR(256),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
);

CREATE TABLE IF NOT EXISTS tickets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trip_id BIGINT REFERENCES trips(id) ON DELETE SET NULL,

    vehicle VARCHAR(256),
    departure_point VARCHAR(256),
    arrival_point VARCHAR(256),
    departure_time TIMESTAMP,
    arrival_time TIMESTAMP,

    cost VARCHAR(256),
    currency VARCHAR(16),
    departure_address VARCHAR(256),
    arrival_address VARCHAR(256),
    seat VARCHAR(256),
    number VARCHAR(256),
    comment VARCHAR(256),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
);

CREATE TABLE IF NOT EXISTS bookings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trip_id BIGINT REFERENCES trips(id) ON DELETE SET NULL,

    city VARCHAR(256),
    place VARCHAR(256),
    check_in_date DATE,
    check_out_date DATE,

    cost VARCHAR(256),
    currency VARCHAR(16),
    address VARCHAR(256),
    check_in_time TIME,
    check_out_time TIME,
    number VARCHAR(256),
    comment VARCHAR(256),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
);

CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trip_id BIGINT REFERENCES trips(id) ON DELETE SET NULL,

    name VARCHAR(256),
    location VARCHAR(256),
    start_time TIMESTAMP,

    end_time TIMESTAMP,
    cost VARCHAR(256),
    currency VARCHAR(256),
    comment VARCHAR(256),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
);

CREATE TABLE IF NOT EXISTS user_daily_usage (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    usage_date DATE NOT NULL,
    upload_count INTEGER NOT NULL DEFAULT 0 CHECK (upload_count >= 0),
    PRIMARY KEY (user_id, usage_date)
);