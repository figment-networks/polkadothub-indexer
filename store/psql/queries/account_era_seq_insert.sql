INSERT INTO account_era_sequences (
  era,
  start_height,
  end_height,
  time,
  stash_account,
  controller_account,
  validator_stash_account,
  validator_controller_account,
  stake
)
VALUES @values

ON CONFLICT (era, stash_account, validator_stash_account) DO UPDATE
SET
  controller_account                 = excluded.controller_account,
  validator_controller_account       = excluded.validator_controller_account,
  stake                              = excluded.stake
