CREATE TABLE IF NOT EXISTS users (
 id BIGSERIAL PRIMARY KEY, pseudo TEXT NOT NULL UNIQUE CHECK (btrim(pseudo) <> ''), bio TEXT NOT NULL DEFAULT '', ville TEXT NOT NULL DEFAULT '', created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS skills (
 user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, nom TEXT NOT NULL, niveau TEXT NOT NULL CHECK (niveau IN ('débutant','intermédiaire','expert')), PRIMARY KEY(user_id, nom)
);
CREATE TABLE IF NOT EXISTS services (
 id BIGSERIAL PRIMARY KEY, provider_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, titre TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', categorie TEXT NOT NULL, duree_minutes INT NOT NULL CHECK(duree_minutes > 0), credits INT NOT NULL CHECK(credits > 0), ville TEXT NOT NULL DEFAULT '', actif BOOLEAN NOT NULL DEFAULT true, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS exchanges (
 id BIGSERIAL PRIMARY KEY, service_id BIGINT NOT NULL REFERENCES services(id), requester_id BIGINT NOT NULL REFERENCES users(id), owner_id BIGINT NOT NULL REFERENCES users(id), status TEXT NOT NULL CHECK(status IN ('pending','accepted','rejected','cancelled','completed')), created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS one_open_exchange_per_service ON exchanges(service_id) WHERE status IN ('pending','accepted');
CREATE TABLE IF NOT EXISTS credit_transactions (
 id BIGSERIAL PRIMARY KEY, user_id BIGINT NOT NULL REFERENCES users(id), exchange_id BIGINT REFERENCES exchanges(id), montant INT NOT NULL, type TEXT NOT NULL CHECK(type IN ('welcome','earn','spend','refund')), created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS credit_transactions_user_idx ON credit_transactions(user_id);
CREATE TABLE IF NOT EXISTS reviews (
 id BIGSERIAL PRIMARY KEY, exchange_id BIGINT NOT NULL REFERENCES exchanges(id), author_id BIGINT NOT NULL REFERENCES users(id), target_id BIGINT NOT NULL REFERENCES users(id), note INT NOT NULL CHECK(note BETWEEN 1 AND 5), commentaire TEXT NOT NULL DEFAULT '', created_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE(exchange_id, author_id)
);
