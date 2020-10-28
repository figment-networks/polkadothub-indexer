package store

const (
	eventSeqWithTxHashQuery = `
	SELECT
		e.height,
		e.method,
		e.section,
		e.data,
		t.hash
	FROM event_sequences AS e
	INNER JOIN transaction_sequences as t
		ON t.height = e.height AND t.index = e.extrinsic_index
	WHERE e.section = ? AND e.method = ?
`

	eventSeqWithTxHashForSrcQuery          = eventSeqWithTxHashQuery + " AND e.data->0->>'value' = ?"
	eventSeqWithTxHashForSrcAndTargetQuery = eventSeqWithTxHashQuery + " AND (e.data->0->>'value' = ? OR e.data->1->>'value' = ?)"
)
