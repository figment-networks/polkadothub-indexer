INSERT INTO rewards (
  created_at,
  updated_at,
  era,
  stash_account,
  validator_stash_account,
  amount,
  kind,
  claimed
)
VALUES @values

ON CONFLICT (era, stash_account, validator_stash_account, kind) DO UPDATE
SET
  updated_at     = excluded.updated_at,
  amount         = excluded.amount,
  claimed        = excluded.claimed
