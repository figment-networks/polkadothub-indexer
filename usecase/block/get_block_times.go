package block

import (
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getBlockTimesUseCase struct {
	blockSeqDb store.BlockSeq
}

func NewGetBlockTimesUseCase(blockSeqDb store.BlockSeq) *getBlockTimesUseCase {
	return &getBlockTimesUseCase{
		blockSeqDb: blockSeqDb,
	}
}

func (uc *getBlockTimesUseCase) Execute(limit int64) store.GetAvgRecentTimesResult {
	return uc.blockSeqDb.GetAvgRecentTimes(limit)
}
