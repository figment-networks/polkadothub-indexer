SELECT * 
FROM validator_summary 
WHERE time_bucket >= (
	SELECT time_bucket 
	FROM validator_summary 
	WHERE time_interval = ?
	ORDER BY time_bucket DESC
	LIMIT 1
) - ?::INTERVAL
	AND stash_account = ? AND time_interval = ?
ORDER BY time_bucket