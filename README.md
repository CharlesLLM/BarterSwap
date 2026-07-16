# BarterSwap

API Go de mise en relation pour des échanges de services.

## Architecture

```text
main.go                  Point d'entrée et assemblage des dépendances
internal/
  domain/                Entités, constantes et erreurs métier
  application/           Cas d'usage et validation métier
  postgres/              Accès aux données et schéma PostgreSQL embarqué
  httpapi/               Routes et handlers HTTP
```

Les dépendances vont de l'extérieur vers le domaine :

```text
HTTP -> Application -> Domain
PostgreSQL ---------> Domain
main.go assemble le tout
```

La couche `application` dépend d'interfaces de repository. Elle peut donc être
testée sans serveur HTTP ni base de données, et l'implémentation PostgreSQL peut
être remplacée sans modifier les règles métier.

## Lancement

Avec Docker Compose :

```bash
docker compose up --build
```

En local, après avoir défini `DATABASE_URL` :

```bash
go run .
```

## Échanges

Toutes les routes d'échange utilisent le header `X-User-ID`.

| Méthode | Route | Description |
| --- | --- | --- |
| `POST` | `/api/exchanges` | Créer une demande d'échange |
| `GET` | `/api/exchanges` | Lister les échanges demandés et reçus |
| `GET` | `/api/exchanges/{id}` | Consulter un échange |
| `PUT` | `/api/exchanges/{id}/accept` | Accepter une demande |
| `PUT` | `/api/exchanges/{id}/reject` | Refuser une demande |
| `PUT` | `/api/exchanges/{id}/complete` | Terminer un échange accepté |
| `PUT` | `/api/exchanges/{id}/cancel` | Annuler un échange en attente ou accepté |

La liste accepte le filtre optionnel `status` :

```bash
curl -H "X-User-ID: 1" \
  "http://localhost:8080/api/exchanges?status=pending"
```

Pour demander puis accepter un service :

```bash
curl -X POST -H "Content-Type: application/json" -H "X-User-ID: 1" \
  -d '{"service_id": 3}' http://localhost:8080/api/exchanges

curl -X PUT -H "X-User-ID: 2" \
  http://localhost:8080/api/exchanges/1/accept
```

## Tests

Depuis la racine du projet :

```bash
go test ./...
```

Pour afficher le nom et le résultat de chaque test :

```bash
go test -v ./...
```

Pour afficher la couverture :

```bash
go test -cover ./...
```
