package chain

import (
	"context"
	"fmt"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type GetStatusCmdHandler struct {
	db     *store.Store
	client *client.Client

	useCase *getStatusUseCase
}

func NewGetStatusCmdHandler(db *store.Store, c *client.Client) *GetStatusCmdHandler {
	return &GetStatusCmdHandler{
		db:     db,
		client: c,
	}
}

func (h *GetStatusCmdHandler) Handle(ctx context.Context) {
	logger.Info("chain get status use case [handler=cmd]")

	details, err := h.getUseCase().Execute(true)
	if err != nil {
		logger.Error(err)
		return
	}

	fmt.Println("=== App ===")
	fmt.Println("Name:", details.AppName)
	fmt.Println("Version", details.AppVersion)
	fmt.Println("Go version", details.GoVersion)
	fmt.Println("")

	fmt.Println("=== Client ===")
	fmt.Println("Library info:", details.ClientInfo)

	fmt.Println("=== Chain ===")
	fmt.Println("Name:", details.ChainName)
	fmt.Println("Type:", details.ChainType)
	fmt.Println("")

	fmt.Println("=== Genesis ===")
	fmt.Println("Hash:", details.GenesisHash)
	fmt.Println("")

	fmt.Println("=== Node ===")
	fmt.Println("Name:", details.NodeName)
	fmt.Println("Version:", details.NodeVersion)
	fmt.Println("LocalPeerUID:", details.NodeLocalPeerUID)
	fmt.Println("Health:", details.NodeHealth)
	fmt.Println("Roles:", details.NodeRoles)
	fmt.Println("Properties:")
	for key, value := range details.NodeProperties {
		if key != "registry" {
			fmt.Println(fmt.Sprintf("  - %s:", key), value)
		}
	}

	if details.IndexingStarted {
		fmt.Println("=== Indexing ===")
		fmt.Println("Last index version:", details.LastIndexVersion)
		fmt.Println("Last indexed height:", details.LastIndexedHeight)
		fmt.Println("Last indexed time:", details.LastIndexedTime)
		fmt.Println("Last indexed at:", details.LastIndexedAt)
		fmt.Println("Last indexed session:", details.LastIndexedSession)
		fmt.Println("Last indexed era:", details.LastIndexedEra)
		fmt.Println("Last indexed spec version:", details.LastSpecVersion)
		fmt.Println("Lag behind head:", details.Lag)
	}
	fmt.Println("")
}

func (h *GetStatusCmdHandler) getUseCase() *getStatusUseCase {
	if h.useCase == nil {
		return NewGetStatusUseCase(h.db, h.client)
	}
	return h.useCase
}
