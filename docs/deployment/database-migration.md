# Database Migration

Run following command to migrate database in a postgres docker container:

```bash
docker compose exec -T backend migrate -path /app/migrations \
  -database "postgres://<database-user>:<database-password>@<database-host>:<database-port>/<database-name>?sslmode=disable" up
```