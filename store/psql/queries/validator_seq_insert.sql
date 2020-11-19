INSERT INTO validator_sequences (
  height,
  time,
  stash_account,
  active_balance
)
VALUES @values

ON CONFLICT (height, stash_account) DO UPDATE
SET
  active_balance   = excluded.active_balance,
