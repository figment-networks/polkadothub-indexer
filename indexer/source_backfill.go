package indexer

import (
	"context"
	"errors"
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/model"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
)

var (
	_ pipeline.Source = (*backfillSource)(nil)
)

func NewBackfillSource(cfg *config.Config, syncablesDb store.Syncables, client *client.Client, indexVersion int64, isLastInSession, isLastInEra bool, transactionDb store.Transactions, trxFilter []model.TransactionKind) (*backfillSource, error) {
	src := &backfillSource{
		cfg:           cfg,
		syncablesDb:   syncablesDb,
		transactionDb: transactionDb,
		client:        client,

		currentIndexVersion: indexVersion,
	}

	if err := src.init(isLastInSession, isLastInEra, trxFilter); err != nil {
		return nil, err
	}

	return src, nil
}

type backfillSource struct {
	cfg                 *config.Config
	syncablesDb         store.Syncables
	transactionDb       store.Transactions
	client              *client.Client
	useWhiteList        bool
	heightsWhitelist    map[int64]struct{}
	whiteListStages     []pipeline.StageName
	currentIndexVersion int64
	currentHeight       int64
	startHeight         int64
	endHeight           int64
	err                 error
}

func (s *backfillSource) Next(context.Context, pipeline.Payload) bool {
	if s.err == nil && s.currentHeight < s.endHeight {
		s.currentHeight = s.currentHeight + 1
		return true
	}
	return false
}

func (s *backfillSource) isCurrentHeightInWhitelist() bool {
	_, found := s.heightsWhitelist[s.currentHeight]
	return found
}

func (s *backfillSource) UseWhiteList() bool {
	return s.useWhiteList
}

func (s *backfillSource) Current() int64 {
	return s.currentHeight
}

func (s *backfillSource) Err() error {
	return s.err
}

func (s *backfillSource) Skip(stageName pipeline.StageName) bool {
	if s.UseWhiteList() {
		if !s.isCurrentHeightInWhitelist() {
			return !s.isStageInWhiteList(stageName)
		} else {
			return false
		}
	}

	return false
}

func (s *backfillSource) Len() int64 {
	return s.endHeight - s.startHeight + 1
}

func (s *backfillSource) init(isLastInSession, isLastInEra bool, kinds []model.TransactionKind) error {
	useWhiteListForEra := isLastInSession || isLastInEra
	s.heightsWhitelist = make(map[int64]struct{})
	if useWhiteListForEra {
		if err := s.setHeightsWhitelistForEra(isLastInSession, isLastInEra); err != nil {
			return err
		}
	}

	useWhiteListForTrxFilter := len(kinds) != 0
	if useWhiteListForTrxFilter {
		if err := s.setHeightsWhitelistForTrxFilter(kinds); err != nil {
			return err
		}
	}

	s.useWhiteList = useWhiteListForEra || useWhiteListForTrxFilter

	if err := s.setStartHeight(); err != nil {
		return err
	}
	if err := s.setEndHeight(); err != nil {
		return err
	}
	return nil
}

func (s *backfillSource) setStartHeight() error {
	syncable, err := s.syncablesDb.FindFirstByDifferentIndexVersion(s.currentIndexVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New(fmt.Sprintf("nothing to backfill [currentIndexVersion=%d]", s.currentIndexVersion))
		}
		return err
	}

	s.currentHeight = syncable.Height
	s.startHeight = syncable.Height
	return nil
}

func (s *backfillSource) setEndHeight() error {
	syncable, err := s.syncablesDb.FindMostRecentByDifferentIndexVersion(s.currentIndexVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New(fmt.Sprintf("nothing to backfill [currentIndexVersion=%d]", s.currentIndexVersion))
		}
		return err
	}

	s.endHeight = syncable.Height
	return nil
}

func (s *backfillSource) setHeightsWhitelistForEra(isLastInSession, isLastInEra bool) error {
	syncables, err := s.syncablesDb.FindAllByLastInSessionOrEra(s.currentIndexVersion, isLastInSession, isLastInEra)
	if err != nil {
		return err
	}

	if len(syncables) == 0 {
		return errors.New(fmt.Sprintf("no heights for whitelist to backfill [currentIndexVersion=%d]", s.currentIndexVersion))
	}

	for i := 0; i < len(syncables); i++ {
		s.heightsWhitelist[syncables[i].Height] = struct{}{}
	}

	return nil
}

func (s *backfillSource) setHeightsWhitelistForTrxFilter(kinds []model.TransactionKind) error {
	for _, kind := range kinds {
		transactions, err := s.transactionDb.GetTransactionsByTransactionKind(kind)
		if err != nil {
			return err
		}

		for _, trx := range transactions {
			s.heightsWhitelist[trx.Height] = struct{}{}
		}
	}

	return nil
}

func (s *backfillSource) isStageInWhiteList(stageName pipeline.StageName) bool {
	if pipeline.StageSyncer == stageName {
		return true
	}

	return false
}
