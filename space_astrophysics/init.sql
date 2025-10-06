-- =====================================
-- Postgres Init SQL for Space Astrophysics
-- Tables: planets, worlds, world_planets, users
-- =====================================

-- =========================
-- USERS
-- =========================
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    is_moderator BOOLEAN NOT NULL DEFAULT FALSE
);

-- =========================
-- PLANETS
-- =========================
CREATE TABLE planets (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    image_url VARCHAR(255),
    status VARCHAR(10) NOT NULL DEFAULT 'active', -- 'active' / 'deleted'
    a DOUBLE PRECISION, -- semi-major axis
    e DOUBLE PRECISION, -- eccentricity
    period DOUBLE PRECISION, -- orbital period
    t0 DOUBLE PRECISION -- time of pericenter
);

-- =========================
-- WORLDS (Заявки)
-- =========================
CREATE TABLE worlds (
    id SERIAL PRIMARY KEY,
    world_status VARCHAR(20) NOT NULL DEFAULT 'draft', -- draft, deleted, formed, completed, rejected
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    creator_id INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    formation_date TIMESTAMP,
    completion_date TIMESTAMP,
    moderator_id INT REFERENCES users(id) ON DELETE RESTRICT
    -- динамические поля (угол/расстояние) не храним, считаются в Go
);

-- =========================
-- WORLD_PLANETS (m-m связь)
-- =========================
CREATE TABLE world_planets (
    world_id INT NOT NULL REFERENCES worlds(id) ON DELETE RESTRICT,
    planet_id INT NOT NULL REFERENCES planets(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 1,
    is_main BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (world_id, planet_id)
);
