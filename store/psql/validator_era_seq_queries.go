package psql

const (
	summarizeValidatorsForEraQuerySelect = `
	stash_account,
	DATE_TRUNC(?, time) AS time_bucket,
   	AVG(total_stake) AS total_stake_avg,
   	MAX(total_stake) AS total_stake_max,
   	MIN(total_stake) AS total_stake_min,
	AVG(own_stake) AS own_stake_avg,
   	MAX(own_stake) AS own_stake_max,
   	MIN(own_stake) AS own_stake_min,
	AVG(stakers_stake) AS stakers_stake_avg,
   	MAX(stakers_stake) AS stakers_stake_max,
   	MIN(stakers_stake) AS stakers_stake_min,
	AVG(reward_points) AS reward_points_avg,
   	MAX(reward_points) AS reward_points_max,
   	MIN(reward_points) AS reward_points_min,
	AVG(commission) AS commission_avg,
   	MAX(commission) AS commission_max,
   	MIN(commission) AS commission_min,
	AVG(stakers_count) AS stakers_count_avg,
   	MAX(stakers_count) AS stakers_count_max,
   	MIN(stakers_count) AS stakers_count_min
`
)
