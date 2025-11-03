-- =====================================
-- Postgres Init SQL for Space Astrophysics
-- Tables: planets, worlds, world_planets, users
-- =====================================

-- ==========================
-- PLANETS
-- ==========================
CREATE TABLE IF NOT EXISTS planets (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image_url TEXT,
    status VARCHAR(20) DEFAULT 'active',
    a DOUBLE PRECISION,
    e DOUBLE PRECISION,
    period DOUBLE PRECISION,
    t0 DOUBLE PRECISION
);

-- ==========================
-- USERS
-- ==========================
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    is_moderator BOOLEAN DEFAULT FALSE
);

-- ==========================
-- WORLDS
-- ==========================
CREATE TABLE IF NOT EXISTS worlds (
    id SERIAL PRIMARY KEY,
    theme VARCHAR(255),
    description TEXT,
    world_status VARCHAR(20) DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT now(),
    creator_id INT REFERENCES users(id),
    formation_date TIMESTAMP,
    completion_date TIMESTAMP,
    moderator_id INT REFERENCES users(id),
    total_cost DOUBLE PRECISION
);

-- ==========================
-- WORLD_PLANETS (m-m)
-- ==========================
CREATE TABLE IF NOT EXISTS world_planets (
    world_id INT NOT NULL REFERENCES worlds(id) ON DELETE RESTRICT,
    planet_id INT NOT NULL REFERENCES planets(id) ON DELETE RESTRICT,
    quantity INT DEFAULT 1,
    angle DOUBLE PRECISION,
    distance DOUBLE PRECISION,
    comment TEXT,
    PRIMARY KEY (world_id, planet_id)
);
