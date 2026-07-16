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

CREATE TABLE IF NOT EXISTS services (
    id BIGSERIAL PRIMARY KEY,
    provider_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    titre TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    categorie TEXT NOT NULL CHECK (categorie IN (
        'Informatique', 'Jardinage', 'Bricolage', 'Cuisine', 'Musique',
        'Langues', 'Sport', 'Tutorat', 'Déménagement', 'Photographie',
        'Animalier', 'Couture', 'Autre'
    )),
    duree_minutes INTEGER NOT NULL CHECK (duree_minutes > 0),
    credits INTEGER NOT NULL CHECK (credits > 0),
    ville TEXT NOT NULL DEFAULT '',
    actif BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS services_provider_id_idx ON services(provider_id);
CREATE INDEX IF NOT EXISTS services_categorie_idx ON services(categorie);
CREATE INDEX IF NOT EXISTS services_ville_idx ON services(ville);

CREATE TABLE IF NOT EXISTS exchanges (
    id BIGSERIAL PRIMARY KEY,
    service_id BIGINT NOT NULL REFERENCES services(id),
    requester_id BIGINT NOT NULL REFERENCES users(id),
    owner_id BIGINT NOT NULL REFERENCES users(id),
    credits INTEGER NOT NULL CHECK (credits > 0),
    status TEXT NOT NULL CHECK (status IN (
        'pending', 'accepted', 'rejected', 'cancelled', 'completed'
    )),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (requester_id <> owner_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS exchanges_active_service_idx
    ON exchanges(service_id)
    WHERE status IN ('pending', 'accepted');
CREATE INDEX IF NOT EXISTS exchanges_requester_id_idx ON exchanges(requester_id);
CREATE INDEX IF NOT EXISTS exchanges_owner_id_idx ON exchanges(owner_id);
CREATE INDEX IF NOT EXISTS exchanges_status_idx ON exchanges(status);

ALTER TABLE credit_transactions
    ADD COLUMN IF NOT EXISTS exchange_id BIGINT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'credit_transactions_exchange_id_fkey'
    ) THEN
        ALTER TABLE credit_transactions
            ADD CONSTRAINT credit_transactions_exchange_id_fkey
            FOREIGN KEY (exchange_id) REFERENCES exchanges(id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS credit_transactions_exchange_id_idx
    ON credit_transactions(exchange_id);

CREATE TABLE IF NOT EXISTS reviews (
    id BIGSERIAL PRIMARY KEY,
    exchange_id BIGINT NOT NULL REFERENCES exchanges(id) ON DELETE CASCADE,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    note INTEGER NOT NULL CHECK (note BETWEEN 1 AND 5),
    commentaire TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (exchange_id, author_id),
    CHECK (author_id <> target_id)
);

CREATE INDEX IF NOT EXISTS reviews_target_id_idx ON reviews(target_id);
CREATE INDEX IF NOT EXISTS reviews_exchange_id_idx ON reviews(exchange_id);
