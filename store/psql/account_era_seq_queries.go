package psql

const (
	findLastAccountEraSeqByStashQuery = `
SELECT * FROM account_era_sequences
  WHERE stash_account = ?
  AND era = (
	SELECT era 
		FROM account_era_sequences 
		WHERE stash_account = ? 
		GROUP BY era 
		ORDER BY era LIMIT 1);
`
	findLastAccountEraSeqByValidatorStashQuery = `
SELECT * FROM account_era_sequences
  WHERE validator_stash_account = ?
  AND era = (
	SELECT era 
		FROM account_era_sequences 
		WHERE validator_stash_account = ? 
		GROUP BY era 
		ORDER BY era LIMIT 1);
`
)
