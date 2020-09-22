package store

type Store interface {
	Close() error
	SetDebugMode(enabled bool)

	Getters
}

type Getters interface {
	GetAccountEraSeq() AccountEraSeq
	GetBlockSeq() BlockSeq
	GetBlockSummary() BlockSummary
	GetDatabase() Database
	GetEventSeq() EventSeq
	GetReports() Reports
	GetSyncables() Syncables
	GetValidatorAgg() ValidatorAgg
	GetValidatorEraSeq() ValidatorEraSeq
	GetValidatorSessionSeq() ValidatorSessionSeq
	GetValidatorSummary() ValidatorSummary
}

type BaseStore interface {
	Create(record interface{}) error
	Update(record interface{}) error
	Save(record interface{}) error
}
