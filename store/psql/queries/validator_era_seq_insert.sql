INSERT INTO validator_era_sequences (
  era,
  start_height,
  end_height,
  time,
  stash_account,
  controller_account,
  session_accounts,
  index,
  total_stake,
  own_stake,
  stakers_stake,
  reward_points,
  commission,
  stakers_count
)
VALUES @values

ON CONFLICT (era, stash_account) DO UPDATE
SET
  controller_account = excluded.controller_account,
  session_accounts = excluded.session_accounts,
  index = excluded.index,
  total_stake = excluded.total_stake,
  own_stake = excluded.own_stake,
  stakers_stake = excluded.stakers_stake,
  reward_points = excluded.reward_points,
  commission = excluded.commission,
  stakers_count = excluded.stakers_count
