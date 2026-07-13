# BarterSwap — API d'échange de compétences

API REST en Go permettant d'échanger des services au moyen de crédits-temps.
Elle utilise uniquement la bibliothèque standard et le pilote PostgreSQL `lib/pq`.

## Installation

Prérequis : Go 1.18+, Docker avec Compose.

```sh
git clone git@github.com:CharlesLLM/BarterSwap.git
cd BarterSwap
cp .env.example .env
docker compose up -d
go mod tidy
go run .
```

La base est initialisée automatiquement.
Par défaut, l'API écoute sur `:8080` et utilise `postgres://barterswap:barterswap@localhost:5432/barterswap?sslmode=disable`. Les variables `DATABASE_URL` et `ADDR` permettent de modifier ces valeurs.

Toutes les routes privées attendent l'identifiant de l'utilisateur dans `X-User-ID`.

## Endpoints

| Méthode | Route | Auth |
| --- | --- | --- |
| POST | `/api/users` | Non |
| GET, PUT | `/api/users/{id}` | Non |
| GET, PUT | `/api/users/{id}/skills` | Non |
| GET | `/api/users/{id}/reviews` | Non |
| GET | `/api/users/{id}/stats` | Non |
| GET, POST | `/api/services` | POST |
| GET, PUT, DELETE | `/api/services/{id}` | Non |
| GET | `/api/services/{id}/reviews` | Non |
| GET, POST | `/api/exchanges` | Oui |
| GET | `/api/exchanges/{id}` | Oui |
| PUT | `/api/exchanges/{id}/{accept,reject,complete,cancel}` | Oui |
| POST | `/api/exchanges/{id}/review` | Oui |

La liste des services accepte `categorie`, `ville` et `search`. La liste des échanges accepte `status`.

## Exemples

```sh
curl -X POST localhost:8080/api/users -H 'Content-Type: application/json' \
  -d '{"pseudo":"alice","ville":"Paris"}'

curl -X PUT localhost:8080/api/users/1/skills -H 'Content-Type: application/json' \
  -H 'X-User-ID: 1' -d '[{"nom":"Jardinage","niveau":"expert"}]'

curl -X POST localhost:8080/api/services -H 'Content-Type: application/json' \
  -H 'X-User-ID: 1' \
  -d '{"titre":"Taille de haie","categorie":"Jardinage","duree_minutes":120,"credits":2,"ville":"Paris"}'

curl -X POST localhost:8080/api/exchanges -H 'Content-Type: application/json' \
  -H 'X-User-ID: 2' -d '{"service_id":1}'
```

## Architecture et garanties

- `api.go` : routage, JSON, codes HTTP et middlewares CORS/logging/recovery/timeout.
- `store.go` : accès SQL et transactions contenant les règles métier atomiques.
- `models.go` et `errors.go` : domaine et erreurs sentinelles.
- `schema.sql` : contraintes, index et journal de crédits.

L'index PostgreSQL partiel empêche deux réservations ouvertes sur le même service. L'acceptation, l'annulation et la complétion verrouillent l'échange et écrivent le journal de crédits dans une même transaction.

## Tests

```sh
go test -v -cover ./...
go vet ./...
```
