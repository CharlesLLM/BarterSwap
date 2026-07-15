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
