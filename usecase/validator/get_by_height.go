package validator

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-proxy/grpc/account/accountpb"
	"github.com/pkg/errors"
)

type getByHeightUseCase struct {
	cfg    *config.Config
	db     *store.Store
	client *client.Client
}

func NewGetByHeightUseCase(cfg *config.Config, db *store.Store, client *client.Client) *getByHeightUseCase {
	return &getByHeightUseCase{
		cfg:    cfg,
		db:     db,
		client: client,
	}
}

func (uc *getByHeightUseCase) Execute(height *int64) (SeqListView, error) {
	// Get last indexed height
	mostRecentSynced, err := uc.db.Syncables.FindMostRecent()
	if err != nil {
		return SeqListView{}, err
	}
	lastH := mostRecentSynced.Height

	// Show last synced height, if not provided
	if height == nil {
		height = &lastH
	}

	if *height > lastH {
		return SeqListView{}, errors.New("height is not indexed yet")
	}

	validatorSessionSequences, err := uc.db.ValidatorSessionSeq.FindByHeight(*height)
	if err != nil && err != store.ErrNotFound {
		return SeqListView{}, err
	}

	validatorEraSequences, err := uc.db.ValidatorEraSeq.FindByHeight(*height)
	if err != nil && err != store.ErrNotFound {
		return SeqListView{}, err
	}

	displayNameMap, err := uc.getDisplayNameMap(validatorSessionSequences)
	if err != nil {
		return SeqListView{}, err
	}

	return ToSeqListView(validatorSessionSequences, validatorEraSequences, displayNameMap), nil
}

func (uc *getByHeightUseCase) getDisplayNameMap(seqs []model.ValidatorSessionSeq) (displayNameMap, error) {
	type result struct {
		addr string
		resp *accountpb.GetIdentityResponse
		err  error
	}

	ch := make(chan result, len(seqs))
	defer close(ch)

	for _, seq := range seqs {
		go func(addr string) {
			resp, err := uc.client.Account.GetIdentity(addr)
			ch <- result{addr, resp, err}
		}(seq.StashAccount)
	}

	displayNameMap := make(displayNameMap)

	for result := range ch {
		if result.err != nil {
			return displayNameMap, nil
		}

		displayNameMap[result.addr] = result.resp.GetIdentity().GetDisplayName()

		if len(displayNameMap) == len(seqs) {
			break
		}
	}
	return displayNameMap, nil

}
