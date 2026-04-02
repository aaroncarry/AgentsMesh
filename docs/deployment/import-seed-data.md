# Import Seed Data

Run following command to import seed data in a postgres docker container:

```bash
psql "postgresql://<database-user>:<database-password>@<database-host>:<database-port>/<database-name>" -f seed.sql
```