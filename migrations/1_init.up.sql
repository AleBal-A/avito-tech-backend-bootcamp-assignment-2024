CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     email VARCHAR(255) UNIQUE NOT NULL,
     password_hash TEXT NOT NULL,
     role VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS houses (
    id SERIAL PRIMARY KEY,
    address VARCHAR(255) NOT NULL,
    year_built INT NOT NULL,
    builder VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_flat_added TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS flats (
    id SERIAL PRIMARY KEY,
    house_id INT NOT NULL REFERENCES houses(id),
    flat_number INT,
    price INT NOT NULL,
    rooms INT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'created',
    moderator_id UUID,
    UNIQUE (flat_number, house_id)
);
