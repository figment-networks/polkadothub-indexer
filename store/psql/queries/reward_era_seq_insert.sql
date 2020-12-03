INSERT INTO reward_era_sequences (
  era,
  start_height,
  end_height,
  time,
  stash_account,
  validator_stash_account,
  amount,
  kind,
  claimed
)
VALUES @values

ON CONFLICT (era, stash_account, validator_stash_account, kind) DO UPDATE
SET
  amount         = excluded.amount,
  claimed        = excluded.claimed
