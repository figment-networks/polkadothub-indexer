SELECT
  time_bucket,
  time_interval,

  AVG(total_stake_avg) AS total_stake_avg,
  MAX(total_stake_max) AS total_stake_max,
  MIN(total_stake_min) AS total_stake_min,
  AVG(own_stake_avg) AS own_stake_avg,
  MAX(own_stake_max) AS own_stake_max,
  MIN(own_stake_min) AS own_stake_min,
  AVG(stakers_stake_avg) AS stakers_stake_avg,
  MAX(stakers_stake_max) AS stakers_stake_max,
  MIN(stakers_stake_min) AS stakers_stake_min,
  AVG(reward_points_avg) AS reward_points_avg,
  MAX(reward_points_max) AS reward_points_max,
  MIN(reward_points_min) AS reward_points_min,
  AVG(commission_avg) AS commission_avg,
  MIN(commission_min) AS commission_min,
  MAX(commission_max) AS commission_max,
  AVG(stakers_count_avg) AS stakers_count_avg,
  MIN(stakers_count_min) AS stakers_count_min,
  MAX(stakers_count_max) AS stakers_count_max,
  AVG(uptime_avg) AS uptime_avg
FROM validator_summary
WHERE time_bucket >= (
	SELECT time_bucket 
	FROM validator_summary 
	WHERE time_interval = ?
	ORDER BY time_bucket DESC 
	LIMIT 1
) - ?::INTERVAL
	AND time_interval = ?
GROUP BY time_bucket, time_interval
ORDER BY time_bucket