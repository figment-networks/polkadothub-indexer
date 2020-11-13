package system_event

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
)

type getForAddressUseCase struct {
	systemEventDb store.SystemEvents
}

func NewGetForAddressUseCase(systemEventDb store.SystemEvents) *getForAddressUseCase {
	return &getForAddressUseCase{
		systemEventDb: systemEventDb,
	}
}

func (uc *getForAddressUseCase) Execute(address string, minHeight *int64, kind *model.SystemEventKind) (*ListView, error) {
	systemEvents, err := uc.systemEventDb.FindByActor(address, kind, minHeight)
	if err != nil {
		return nil, err
	}

	return ToListView(systemEvents), nil
}
