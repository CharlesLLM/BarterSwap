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

## Swagger

Une fois l'application démarrée, l'interface de test est disponible sur :

```text
http://localhost:8080/swagger/
```

Utilisez le bouton **Authorize** pour renseigner le header `X-User-ID`, puis
exécutez les requêtes directement depuis la documentation. Le schéma OpenAPI
brut est également disponible sur `http://localhost:8080/openapi.yaml`.

Swagger UI est chargé depuis un CDN ; une connexion Internet est donc nécessaire
pour l'interface graphique, mais pas pour consulter le schéma OpenAPI brut.

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

## Parcours nominal complet

Ce parcours teste tout le cycle principal, de la création des utilisateurs au transfert définitif des crédits.

### 1. Créer l'offreur

Chaque nouvel utilisateur reçoit automatiquement 10 crédits.

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{
    "pseudo": "alice",
    "bio": "Passionnée de jardinage",
    "ville": "Paris"
  }' \
  http://localhost:8080/api/users
```

### 2. Créer le demandeur

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{
    "pseudo": "bob",
    "bio": "Je souhaite apprendre à jardiner",
    "ville": "Paris"
  }' \
  http://localhost:8080/api/users
```

### 3. Ajouter la compétence de l'offreur

```bash
curl -X PUT -H "Content-Type: application/json" -H "X-User-ID: 1" \
  -d '[
    {
      "nom": "Jardinage",
      "niveau": "expert"
    }
  ]' \
  http://localhost:8080/api/users/1/skills
```

### 4. Publier un service à 3 crédits

```bash
curl -X POST -H "Content-Type: application/json" -H "X-User-ID: 1" \
  -d '{
    "titre": "Initiation au jardinage",
    "description": "Apprendre à préparer et entretenir un potager",
    "categorie": "Jardinage",
    "duree_minutes": 60,
    "credits": 3,
    "ville": "Paris"
  }' \
  http://localhost:8080/api/services
```

### 5. Demander le service

Bob, l'utilisateur `2`, demande le service `1` d'Alice.

```bash
curl -X POST -H "Content-Type: application/json" -H "X-User-ID: 2" \
  -d '{"service_id": 1}' \
  http://localhost:8080/api/exchanges
```

L'échange doit être créé avec le statut `pending`. Aucun crédit n'est encore
débité.

### 6. Consulter les demandes reçues

```bash
curl -H "X-User-ID: 1" \
  "http://localhost:8080/api/exchanges?status=pending"
```

### 7. Accepter l'échange

Alice accepte l'échange. Les 3 crédits sont alors débités du solde de Bob, mais
ils ne sont pas encore crédités à Alice.

```bash
curl -X PUT -H "X-User-ID: 1" \
  http://localhost:8080/api/exchanges/1/accept
```

Le solde de Bob doit maintenant être égal à 7 :

```bash
curl http://localhost:8080/api/users/2
```

### 8. Terminer l'échange

Après la réalisation du service, un des deux participants marque l'échange
comme terminé.

```bash
curl -X PUT -H "X-User-ID: 2" \
  http://localhost:8080/api/exchanges/1/complete
```

### 9. Vérifier le résultat final

```bash
curl -H "X-User-ID: 1" http://localhost:8080/api/exchanges/1
curl http://localhost:8080/api/users/1
curl http://localhost:8080/api/users/2
```

Le résultat attendu est :

- échange au statut `completed` ;
- Alice possède 13 crédits : 10 crédits de bienvenue et 3 crédits gagnés ;
- Bob possède 7 crédits : 10 crédits de bienvenue et 3 crédits dépensés.

### 10. Évaluer l'échange

Une fois l'échange terminé, Bob peut évaluer Alice. L'auteur et le destinataire
sont déduits automatiquement de l'échange.

```bash
curl -X POST -H "Content-Type: application/json" -H "X-User-ID: 2" \
  -d '{
    "note": 5,
    "commentaire": "Service excellent et très pédagogique"
  }' \
  http://localhost:8080/api/exchanges/1/review
```

L'avis est ensuite visible sur le profil d'Alice et sur le service :

```bash
curl http://localhost:8080/api/users/1/reviews
curl http://localhost:8080/api/services/1/reviews
```

## Tests

Depuis la racine du projet :

```bash
go test -v ./...
```
