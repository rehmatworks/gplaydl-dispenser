-- The dispenser no longer runs a shared community pool: every account is now
-- private to its owner. Fold any account that was public back to private so it
-- keeps working, owned by the same person, just no longer handed to strangers.
UPDATE accounts SET visibility = 'private', shared_at = NULL WHERE visibility = 'public';

-- New accounts are private from here on. The visibility, shared_at and
-- consent_version columns are left in place: the code stops using them, but a
-- rolled-back binary during deploy still expects to read and write them. A
-- later migration can drop them once this one has settled.
ALTER TABLE accounts ALTER COLUMN visibility SET DEFAULT 'private';
