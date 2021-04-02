package indexer

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
)

var (
	_ pipeline.Source = (*reindexSource)(nil)
)

func NewReindexSource(cfg *config.Config, syncablesDb store.Syncables, client *client.Client, indexVersion int64, isLastInSession, isLastInEra bool, transactionDb store.Transactions, trxFilter []model.TransactionKind) (*reindexSource, error) {
	src := &reindexSource{
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

type reindexSource struct {
	cfg                 *config.Config
	syncablesDb         store.Syncables
	transactionDb       store.Transactions
	client              *client.Client
	heightsWhitelist    map[int64]struct{}
	sortedWhiteListKeys []int64
	sortedIndex         int64
	whiteListStages     []pipeline.StageName
	currentIndexVersion int64
	currentHeight       int64
	startHeight         int64
	endHeight           int64
	err                 error
}

func (s *reindexSource) Next(context.Context, pipeline.Payload) bool {
	s.sortedIndex = s.sortedIndex + 1
	if s.currentHeight < s.endHeight && s.sortedIndex+1 <= int64(len(s.sortedWhiteListKeys)) {
		s.currentHeight = s.sortedWhiteListKeys[s.sortedIndex]
		return true
	}
	return false
}

func (s *reindexSource) Current() int64 {
	return s.currentHeight
}

func (s *reindexSource) Err() error {
	return s.err
}

func (s *reindexSource) Skip(stageName pipeline.StageName) bool {
	return false
}

func (s *reindexSource) Len() int64 {
	return s.endHeight - s.startHeight + 1
}

func (s *reindexSource) init(isLastInSession, isLastInEra bool, kinds []model.TransactionKind) error {
	heightsWhitelist := make(map[int64]struct{})
	if isLastInSession || isLastInEra {
		if err := s.setHeightsWhitelistForEraOrSession(heightsWhitelist, isLastInSession, isLastInEra); err != nil {
			return err
		}
	}

	useWhiteListForTrxFilter := len(kinds) != 0
	if useWhiteListForTrxFilter {
		if err := s.setHeightsWhitelistForTrxFilter(heightsWhitelist, kinds); err != nil {
			return err
		}
	}

	if len(heightsWhitelist) == 0 {
		return errors.New("no heights for whitelist to reindex")
	}

	logger.Info(fmt.Sprintf("[reindex] set %d heights to reindex", len(heightsWhitelist)))

	keysWhitelist := make([]int64, 0, len(heightsWhitelist))
	for k := range heightsWhitelist {
		keysWhitelist = append(keysWhitelist, k)
	}
	sort.Slice(keysWhitelist, func(i, j int) bool { return keysWhitelist[i] < keysWhitelist[j] })
	s.sortedWhiteListKeys = keysWhitelist

	if err := s.setStartHeight(); err != nil {
		return err
	}
	if err := s.setEndHeight(); err != nil {
		return err
	}

	s.currentHeight = s.sortedWhiteListKeys[0]

	return nil
}

func (s *reindexSource) setStartHeight() error {
	syncable, err := s.syncablesDb.FindFirstByDifferentIndexVersion(s.currentIndexVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New(fmt.Sprintf("nothing to reindex [currentIndexVersion=%d]", s.currentIndexVersion))
		}
		return err
	}

	s.currentHeight = syncable.Height
	s.startHeight = syncable.Height
	return nil
}

func (s *reindexSource) setEndHeight() error {
	syncable, err := s.syncablesDb.FindMostRecentByDifferentIndexVersion(s.currentIndexVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New(fmt.Sprintf("nothing to reindex [currentIndexVersion=%d]", s.currentIndexVersion))
		}
		return err
	}

	s.endHeight = syncable.Height
	return nil
}

func (s *reindexSource) setHeightsWhitelistForEraOrSession(whitelist map[int64]struct{}, isLastInSession, isLastInEra bool) error {
	syncables, err := s.syncablesDb.FindAllByLastInSessionOrEra(-1, isLastInSession, isLastInEra)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("[setHeightsWhitelistForEraOrSession] set %d heights in whitelist", len(syncables)))

	for i := 0; i < len(syncables); i++ {
		whitelist[syncables[i].Height] = struct{}{}
	}

	return nil
}

func (s *reindexSource) setHeightsWhitelistForTrxFilter(whitelist map[int64]struct{}, kinds []model.TransactionKind) error {
	beforelen := len(whitelist)

	for _, kind := range kinds {
		transactions, err := s.transactionDb.GetTransactionsByTransactionKind(kind)
		if err != nil {
			return err
		}

		for _, trx := range transactions {
			whitelist[trx.Height] = struct{}{}
		}
	}

	logger.Info(fmt.Sprintf("[setHeightsWhitelistForTrxFilter] set %d heights in whitelist", len(whitelist)-beforelen))
	return nil
}
