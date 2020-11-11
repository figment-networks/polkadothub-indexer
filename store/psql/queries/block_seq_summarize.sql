DATE_TRUNC(?, time) AS time_bucket,
COUNT(*) AS count,
EXTRACT(EPOCH FROM (MAX(time) - MIN(time)) / COUNT(*)) AS block_time_avg