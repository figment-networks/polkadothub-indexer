package blockseqrepo

import (
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/models/blockseq"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
	"github.com/jinzhu/gorm"
)

type DbRepo interface {
	// Queries
	Exists(types.Height) bool
	Count() (*int64, errors.ApplicationError)
	GetByHeight(types.Height) (*blockseq.Model, errors.ApplicationError)
	GetMostRecent(BlockDbQuery) (*blockseq.Model, errors.ApplicationError)
	GetRecentProcessed(int64) ([]blockseq.Model, errors.ApplicationError)
	GetAvgBlockTimesForRecentBlocks(int64) Result
	GetAvgBlockTimesForInterval(string, string) ([]Row, errors.ApplicationError)

	// Commands
	Save(*blockseq.Model) errors.ApplicationError
	Create(*blockseq.Model) errors.ApplicationError
}

type dbRepo struct {
	client *gorm.DB
}

func NewDbRepo(c *gorm.DB) DbRepo {
	return &dbRepo{
		client: c,
	}
}

func (r *dbRepo) Exists(h types.Height) bool {
	q := heightQuery(h)
	m := blockseq.Model{}

	if err := r.client.Where(&q).First(&m).Error; err != nil {
		return false
	}
	return true
}

func (r *dbRepo) Count() (*int64, errors.ApplicationError) {
	var count int64
	if err := r.client.Table(blockseq.Model{}.TableName()).Count(&count).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.NewError("block sequence not found", errors.NotFoundError, err)
		}
		return nil, errors.NewError("error getting count of block sequences", errors.QueryError, err)
	}
	return &count, nil
}

func (r *dbRepo) GetByHeight(h types.Height) (*blockseq.Model, errors.ApplicationError) {
	q := heightQuery(h)
	m := blockseq.Model{}

	if err := r.client.Where(&q).First(&m).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.NewError("block sequence not found", errors.NotFoundError, err)
		}
		return nil, errors.NewError(fmt.Sprintf("could not find block sequence with height %d", h), errors.QueryError, err)
	}
	return &m, nil
}

func (r *dbRepo) GetMostRecent(q BlockDbQuery) (*blockseq.Model, errors.ApplicationError) {
	m := blockseq.Model{}
	if err := r.client.Where(q.String()).Order("height desc").First(&m).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.NewError("most recent block sequence not found", errors.NotFoundError, err)
		}
		return nil, errors.NewError("could not find most recent block sequence", errors.QueryError, err)
	}
	return &m, nil
}

func (r *dbRepo) GetRecentProcessed(limit int64) ([]blockseq.Model, errors.ApplicationError) {
	var ms []blockseq.Model
	if err := r.client.Where("processed_at IS NOT NULL").Order("height desc").Limit(limit).Find(&ms).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.NewError("most recent block sequence not found", errors.NotFoundError, err)
		}
		return nil, errors.NewError("could not find most recent block sequence", errors.QueryError, err)
	}
	return ms, nil
}

type Result struct {
	StartHeight int64   `json:"start_height"`
	EndHeight   int64   `json:"end_height"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	Count       int64   `json:"count"`
	Diff        float64 `json:"diff"`
	Avg         float64 `json:"avg"`
}

func (r *dbRepo) GetAvgBlockTimesForRecentBlocks(limit int64) Result {
	var res Result
	r.client.Raw(blockTimesForRecentBlocksQuery, limit).Scan(&res)

	return res
}

type Row struct {
	TimeInterval string  `json:"time_interval"`
	Count        int64   `json:"count"`
	Avg          float64 `json:"avg"`
}

func (r *dbRepo) GetAvgBlockTimesForInterval(interval string, period string) ([]Row, errors.ApplicationError) {
	rows, err := r.client.Debug().Raw(blockTimesForIntervalQuery, interval, period).Rows()
	if err != nil {
		return nil, errors.NewError("could not query block times for interval", errors.QueryError, err)
	}
	defer rows.Close()

	var res []Row
	for rows.Next() {
		var row Row
		if err := r.client.ScanRows(rows, &row); err != nil {
			return nil, errors.NewError("could not scan rows", errors.QueryError, err)
		}

		res = append(res, row)
	}
	return res, nil
}

func (r *dbRepo) Save(m *blockseq.Model) errors.ApplicationError {
	if err := r.client.Save(m).Error; err != nil {
		return errors.NewError("could not save block sequence", errors.SaveError, err)
	}
	return nil
}

func (r *dbRepo) Create(m *blockseq.Model) errors.ApplicationError {
	if err := r.client.Create(m).Error; err != nil {
		return errors.NewError("could not create block sequence", errors.CreateError, err)
	}
	return nil
}

type BlockDbQuery struct {
	Processed bool
}

func (bq *BlockDbQuery) String() string {
	q := ""
	if bq.Processed {
		q += "processed_at IS NOT NULL"
	}
	return q
}

/*************** Private ***************/

func heightQuery(h types.Height) blockseq.Model {
	return blockseq.Model{
		HeightSequence: &shared.HeightSequence{
			Height: h,
		},
	}
}
