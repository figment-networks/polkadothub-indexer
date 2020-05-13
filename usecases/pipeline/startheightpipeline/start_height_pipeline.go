package startheightpipeline

import (
	"context"
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/models/report"
	"github.com/figment-networks/polkadothub-indexer/repos/blockseqrepo"
	"github.com/figment-networks/polkadothub-indexer/repos/extrinsicseqrepo"
	"github.com/figment-networks/polkadothub-indexer/repos/reportrepo"
	"github.com/figment-networks/polkadothub-indexer/repos/syncablerepo"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
	"github.com/figment-networks/polkadothub-indexer/utils/iterators"
	"github.com/figment-networks/polkadothub-indexer/utils/log"
	"github.com/figment-networks/polkadothub-indexer/utils/pipeline"
)

type UseCase interface {
	Execute(context.Context, int64) errors.ApplicationError
}

type useCase struct {
	syncableDbRepo    syncablerepo.DbRepo
	syncableProxyRepo syncablerepo.ProxyRepo

	blockSeqDbRepo       blockseqrepo.DbRepo
	extrinsicSeqDbRepo extrinsicseqrepo.DbRepo

	reportDbRepo reportrepo.DbRepo
}

func NewUseCase(
	syncableDbRepo syncablerepo.DbRepo,
	syncableProxyRepo syncablerepo.ProxyRepo,
	blockSeqDbRepo blockseqrepo.DbRepo,
	extrinsicSeqDbRepo extrinsicseqrepo.DbRepo,
	reportDbRepo reportrepo.DbRepo,
) UseCase {
	return &useCase{
		syncableDbRepo:    syncableDbRepo,
		syncableProxyRepo: syncableProxyRepo,

		blockSeqDbRepo:               blockSeqDbRepo,
		extrinsicSeqDbRepo:         extrinsicSeqDbRepo,

		reportDbRepo: reportDbRepo,
	}
}

func (uc *useCase) Execute(ctx context.Context, batchSize int64) errors.ApplicationError {
	i, err := uc.buildIterator(batchSize)
	if err != nil {
		return err
	}

	r, err := uc.createReport(types.SequenceTypeHeight, types.SequenceId(i.StartHeight()), types.SequenceId(i.EndHeight()))
	if err != nil {
		return err
	}

	p := NewPipeline(
		Config{
			SyncableDbRepo: uc.syncableDbRepo,
			Report:         *r,
		},

		pipeline.FIFO(NewSyncer(uc.syncableDbRepo, uc.syncableProxyRepo)),
		pipeline.FIFO(NewSequencer(uc.blockSeqDbRepo, uc.extrinsicSeqDbRepo)),
		pipeline.FIFO(NewAggregator()),
	)

	resp := p.Start(ctx, i)

	if resp.Error == nil {
		err = uc.completeReport(r, resp)
		if err != nil {
			return err
		}
	} else {
		return errors.NewErrorFromMessage(*resp.Error, errors.PipelineProcessingError)
	}

	return nil
}

/*************** Private ***************/

func (uc *useCase) buildIterator(batchSize int64) (*iterators.HeightIterator, errors.ApplicationError) {
	// Get start height
	h, err := uc.syncableDbRepo.GetMostRecentCommon(types.SequenceTypeHeight)
	var startH types.Height
	if err != nil {
		if err.Status() == errors.NotFoundError {
			startH = config.FirstBlockHeight()
		} else {
			return nil, err
		}
	} else {
		startH = types.Height(*h) + 1
	}

	// Get end height. It is most recent height in the node
	headInfo, err := uc.syncableProxyRepo.GetHeadInfo()
	if err != nil {
		return nil, err
	}
	endH := headInfo.Height

	// Validation of heights
	blocksToSyncCount := int64(endH - startH)
	if blocksToSyncCount == 0 {
		log.Warn("nothing to process", log.Field("type", "blockSync"))
		return nil, errors.NewErrorFromMessage("nothing to process", errors.PipelineProcessingError)
	}

	// Make sure that batch limit is respected
	if endH.Int64()-startH.Int64() > batchSize {
		endH = types.Height(startH.Int64()+batchSize) - 1
	}

	log.Info(fmt.Sprintf("iterator: %d - %d", startH, endH))

	i := iterators.NewHeightIterator(startH, endH)
	return i, nil
}

func (uc *useCase) createReport(st types.SequenceType, start types.SequenceId, end types.SequenceId) (*report.Model, errors.ApplicationError) {
	r := &report.Model{
		SequenceType: st,
		StartSequenceId: start,
		EndSequenceId:   end,
	}
	err := uc.reportDbRepo.Create(r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (uc *useCase) completeReport(report *report.Model, resp Results) errors.ApplicationError {
	report.Complete(resp.SuccessCount, resp.ErrorCount, resp.Error, resp.Details)

	return uc.reportDbRepo.Save(report)
}
