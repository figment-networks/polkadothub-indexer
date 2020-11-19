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
  b.controller_account = m.controller_account,
  b.session_accounts = m.session_accounts,
  b.index = m.index,
  b.total_stake = m.total_stake,
  b.own_stake = m.own_stake,
  b.stakers_stake = m.stakers_stake,
  b.reward_points = m.reward_points,
  b.commission = m.commission,
  b.stakers_count = m.stakers_count
