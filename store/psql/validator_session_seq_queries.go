package psql

const (
	summarizeValidatorsForSessionQuerySelect = `
	stash_account,
	DATE_TRUNC(?, time) AS time_bucket,
   	AVG(online::INT) AS uptime_avg,
   	MAX(online::INT) AS uptime_max,
   	MIN(online::INT) AS uptime_min
`
)
