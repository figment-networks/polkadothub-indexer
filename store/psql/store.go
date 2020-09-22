package psql

import (
	"reflect"
	"time"

	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	_ store.Store = (*Store)(nil)
)

// Store handles all database operations
type Store struct {
	db *gorm.DB

	AccountEraSeq       store.AccountEraSeq
	BlockSeq            store.BlockSeq
	BlockSummary        store.BlockSummary
	Database            store.Database
	EventSeq            store.EventSeq
	Reports             store.Reports
	Syncables           store.Syncables
	ValidatorAgg        store.ValidatorAgg
	ValidatorEraSeq     store.ValidatorEraSeq
	ValidatorSessionSeq store.ValidatorSessionSeq
	ValidatorSummary    store.ValidatorSummary
}

// New returns a new postgres store from the connection string
func New(connStr string) (*Store, error) {
	conn, err := gorm.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	registerPlugins(conn)

	return &Store{
		db: conn,

		Database:  NewDatabaseStore(conn),
		Syncables: NewSyncablesStore(conn),
		Reports:   NewReportsStore(conn),

		BlockSeq:            NewBlockSeqStore(conn),
		ValidatorSessionSeq: NewValidatorSessionSeqStore(conn),
		ValidatorEraSeq:     NewValidatorEraSeqStore(conn),
		AccountEraSeq:       NewAccountEraSeqStore(conn),
		EventSeq:            NewEventSeqStore(conn),

		ValidatorAgg: NewValidatorAggStore(conn),

		BlockSummary:     NewBlockSummaryStore(conn),
		ValidatorSummary: NewValidatorSummaryStore(conn),
	}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// GetAccountEraSeq gets AccountEraSeq
func (s *Store) GetAccountEraSeq() store.AccountEraSeq {
	return s.AccountEraSeq
}

// GetBlockSeq gets BlockSeq
func (s *Store) GetBlockSeq() store.BlockSeq {
	return s.BlockSeq
}

// GetBlockSummary gets BlockSummary
func (s *Store) GetBlockSummary() store.BlockSummary {
	return s.BlockSummary
}

// GetDatabase gets Database
func (s *Store) GetDatabase() store.Database {
	return s.Database
}

// GetEventSeq gets EventSeq
func (s *Store) GetEventSeq() store.EventSeq {
	return s.EventSeq
}

// GetReports gets Reports
func (s *Store) GetReports() store.Reports {
	return s.Reports
}

// GetSyncables gets Syncables
func (s *Store) GetSyncables() store.Syncables {
	return s.Syncables
}

// GetValidatorAgg gets ValidatorAgg
func (s *Store) GetValidatorAgg() store.ValidatorAgg {
	return s.ValidatorAgg
}

// GetValidatorEraSeq gets ValidatorEraSeq
func (s *Store) GetValidatorEraSeq() store.ValidatorEraSeq {
	return s.ValidatorEraSeq
}

// GetValidatorSessionSeq gets ValidatorSessionSeq
func (s *Store) GetValidatorSessionSeq() store.ValidatorSessionSeq {
	return s.ValidatorSessionSeq
}

// GetValidatorSummary gets ValidatorSummary
func (s *Store) GetValidatorSummary() store.ValidatorSummary {
	return s.ValidatorSummary
}

// Test checks the connection status
func (s *Store) Test() error {
	return s.db.DB().Ping()
}

// SetDebugMode enabled detailed query logging
func (s *Store) SetDebugMode(enabled bool) {
	s.db.LogMode(enabled)
}

// registerPlugins registers gorm plugins
func registerPlugins(c *gorm.DB) {
	c.Callback().Create().Before("gorm:create").Register("db_plugin:before_create", castQuantity)
	c.Callback().Update().Before("gorm:update").Register("db_plugin:before_update", castQuantity)
}

// castQuantity casts decimal to quantity type
func castQuantity(scope *gorm.Scope) {
	for _, f := range scope.Fields() {
		v := f.Field.Type().String()
		if v == "types.Quantity" {
			f.IsNormal = true
			t := f.Field.Interface().(types.Quantity)
			f.Field = reflect.ValueOf(gorm.Expr("cast(? AS DECIMAL(65,0))", t.String()))
		}
	}
}

func logQueryDuration(start time.Time, queryName string) {
	elapsed := time.Since(start)
	metric.DatabaseQueryDuration.WithLabelValues(queryName).Set(elapsed.Seconds())
}
