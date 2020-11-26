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

const batchSize = 500

var (
	_ store.Accounts     = (*accounts)(nil)
	_ store.Blocks       = (*blocks)(nil)
	_ store.Database     = (*database)(nil)
	_ store.Events       = (*events)(nil)
	_ store.Reports      = (*reports)(nil)
	_ store.Rewards      = (*rewards)(nil)
	_ store.Validators   = (*validators)(nil)
	_ store.Syncables    = (*syncables)(nil)
	_ store.SystemEvents = (*systemEvents)(nil)
	_ store.Transactions = (*transactions)(nil)
)

type Store struct {
	db           *gorm.DB
	accounts     *accounts
	blocks       *blocks
	database     *database
	events       *events
	reports      *reports
	rewards      *rewards
	syncables    *syncables
	systemEvents *systemEvents
	transactions *transactions
	validators   *validators
}

type accounts struct {
	*AccountEraSeqStore
}

type blocks struct {
	*BlockSeqStore
	*BlockSummaryStore
}

type database struct {
	*DatabaseStore
}

type events struct {
	*EventSeqStore
}

type reports struct {
	*ReportsStore
}

type rewards struct {
	*RewardsStore
}
type syncables struct {
	*SyncablesStore
}

type systemEvents struct {
	*SystemEventStore
}
type transactions struct {
	*TransactionSeqStore
}

type validators struct {
	*ValidatorAggStore
	*ValidatorSeqStore
	*ValidatorEraSeqStore
	*ValidatorSessionSeqStore
	*ValidatorSummaryStore
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
	}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// GetAccounts gets accounts
func (s *Store) GetAccounts() *accounts {
	if s.accounts == nil {
		s.accounts = &accounts{
			NewAccountEraSeqStore(s.db),
		}
	}
	return s.accounts
}

// GetBlocks gets blocks
func (s *Store) GetBlocks() *blocks {
	if s.blocks == nil {
		s.blocks = &blocks{
			NewBlockSeqStore(s.db),
			NewBlockSummaryStore(s.db),
		}
	}
	return s.blocks
}

// GetDatabase gets database
func (s *Store) GetDatabase() *database {
	if s.database == nil {
		s.database = &database{
			NewDatabaseStore(s.db),
		}
	}
	return s.database
}

// GetEvents gets events
func (s *Store) GetEvents() *events {
	if s.events == nil {
		s.events = &events{
			NewEventSeqStore(s.db),
		}
	}
	return s.events
}

// GetReports gets reports
func (s *Store) GetReports() *reports {
	if s.reports == nil {
		s.reports = &reports{
			NewReportsStore(s.db),
		}
	}
	return s.reports
}

// GetRewards gets rewards
func (s *Store) GetRewards() *rewards {
	if s.rewards == nil {
		s.rewards = &rewards{
			NewRewardsStore(s.db),
		}
	}
	return s.rewards
}

// GetSyncables gets syncables
func (s *Store) GetSyncables() *syncables {
	if s.syncables == nil {
		s.syncables = &syncables{
			NewSyncablesStore(s.db),
		}
	}
	return s.syncables
}

// GetSystemEvents gets syncables
func (s *Store) GetSystemEvents() *systemEvents {
	if s.systemEvents == nil {
		s.systemEvents = &systemEvents{
			NewSystemEventsStore(s.db),
		}
	}
	return s.systemEvents
}

// GetTransactions gets transactions
func (s *Store) GetTransactions() *transactions {
	if s.transactions == nil {
		s.transactions = &transactions{
			NewTransactionSeqStore(s.db),
		}
	}
	return s.transactions
}

// GetValidators gets validators
func (s *Store) GetValidators() *validators {
	if s.validators == nil {
		s.validators = &validators{
			NewValidatorAggStore(s.db),
			NewValidatorSeqStore(s.db),
			NewValidatorEraSeqStore(s.db),
			NewValidatorSessionSeqStore(s.db),
			NewValidatorSummaryStore(s.db),
		}
	}
	return s.validators
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
