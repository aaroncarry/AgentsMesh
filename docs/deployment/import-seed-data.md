# Import Seed Data

Copy `deploy/selfhost/seed/seed.sql` to `seed.sql` to a postgres docker container and then run following command to import seed data:

```bash
psql "postgresql://<database-user>:<database-password>@<database-host>:<database-port>/<database-name>" -f seed.sql
```