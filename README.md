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
