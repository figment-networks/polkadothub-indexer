package block

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type getBlockSummaryUseCase struct {
	blockSummaryDb store.BlockSummary
}

func NewGetBlockSummaryUseCase(blockSummaryDb store.BlockSummary) *getBlockSummaryUseCase {
	return &getBlockSummaryUseCase{
		blockSummaryDb: blockSummaryDb,
	}
}

func (uc *getBlockSummaryUseCase) Execute(interval types.SummaryInterval, period string) ([]model.BlockSummary, error) {
	return uc.blockSummaryDb.FindSummaries(interval, period)
}
