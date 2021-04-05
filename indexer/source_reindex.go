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

func NewReindexSource(cfg *config.Config, syncablesDb store.Syncables, transactionDb store.Transactions, client *client.Client, indexVersion int64, isLastInSession, isLastInEra bool, trxFilter []model.TransactionKind,
	startHeight, endHeight int64) (*reindexSource, error) {
	src := &reindexSource{
		cfg:           cfg,
		syncablesDb:   syncablesDb,
		transactionDb: transactionDb,
		client:        client,

		currentIndexVersion: indexVersion,
	}

	if err := src.init(isLastInSession, isLastInEra, trxFilter, startHeight, endHeight); err != nil {
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

func (s *reindexSource) init(isLastInSession, isLastInEra bool, kinds []model.TransactionKind, startHeight, endHeight int64) error {
	heightsWhitelist := make(map[int64]struct{})
	if isLastInSession || isLastInEra {
		if err := s.setHeightsWhitelistForEraOrSession(heightsWhitelist, isLastInSession, isLastInEra, startHeight, endHeight); err != nil {
			return err
		}
	}

	useWhiteListForTrxFilter := len(kinds) != 0
	if useWhiteListForTrxFilter {
		if err := s.setHeightsWhitelistForTrxFilter(heightsWhitelist, kinds, startHeight, endHeight); err != nil {
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

	fmt.Println("keysWhitelist", keysWhitelist[0])

	if err := s.setStartHeight(startHeight); err != nil {
		return err
	}
	if err := s.setEndHeight(endHeight); err != nil {
		return err
	}

	s.currentHeight = s.sortedWhiteListKeys[0]

	return nil
}

func (s *reindexSource) setStartHeight(startHeight int64) error {
	startH := startHeight
	if startH < 1 {
		startH = 1
	}
	s.currentHeight = startH
	s.startHeight = startH
	return nil
}

func (s *reindexSource) setEndHeight(endHeight int64) error {
	if endHeight > 0 {
		s.endHeight = endHeight
		return nil
	}

	syncable, err := s.syncablesDb.FindMostRecent()
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New("nothing to reindex")
		}
		return err
	}

	s.endHeight = syncable.Height
	return nil
}

func (s *reindexSource) setHeightsWhitelistForEraOrSession(whitelist map[int64]struct{}, isLastInSession, isLastInEra bool, startHeight, endHeight int64) error {
	syncables, err := s.syncablesDb.FindAllByLastInSessionOrEra(-1, isLastInSession, isLastInEra, startHeight, endHeight)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("[setHeightsWhitelistForEraOrSession] set %d heights in whitelist", len(syncables)))

	for i := 0; i < len(syncables); i++ {
		whitelist[syncables[i].Height] = struct{}{}
	}

	return nil
}

func (s *reindexSource) setHeightsWhitelistForTrxFilter(whitelist map[int64]struct{}, kinds []model.TransactionKind, startHeight, endHeight int64) error {
	beforelen := len(whitelist)

	for _, kind := range kinds {
		transactions, err := s.transactionDb.GetTransactionsByTransactionKind(kind, startHeight, endHeight)
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
