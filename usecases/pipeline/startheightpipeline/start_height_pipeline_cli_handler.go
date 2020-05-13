package startheightpipeline

import (
	"context"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cliHandler struct {
	useCase UseCase
}

func NewCliHandler(useCase UseCase) types.CliHandler {
	return &cliHandler{
		useCase: useCase,
	}
}

func (h *cliHandler) Handle(*cobra.Command, []string) {
	batchSize := viper.GetInt64("batchSize")
	ctx := context.Background()

	err := h.useCase.Execute(ctx, batchSize)
	if err != nil {
		log.Error(err)
		return
	}
}

