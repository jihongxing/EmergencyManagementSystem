# Database migrations

Goose SQL migrations are embedded in the migration command. Run `go run ./cmd/migrate`
from `backend/` with `DATABASE_URL` set. The API never migrates automatically.
Version 1 creates only the `ems` schema; Goose stores migration metadata in
`public.goose_db_version`. No business tables or simulated authority are created.

Each new migration must follow a module contract and include an integration test.
Do not edit applied migrations. Update the readiness baseline when adding required
schema versions. No automated production downgrade is exposed.
