CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    pseudo TEXT NOT NULL UNIQUE,
    bio TEXT NOT NULL DEFAULT '',
    ville TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS credit_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    montant INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('welcome', 'earn', 'spend', 'refund')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS skills (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nom TEXT NOT NULL,
    niveau TEXT NOT NULL CHECK (niveau IN ('débutant', 'intermédiaire', 'expert')),
    PRIMARY KEY (user_id, nom)
);
