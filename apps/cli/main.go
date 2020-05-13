package main

import (
	"github.com/figment-networks/polkadothub-indexer/apps/shared"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/repos/blockseqrepo"
	"github.com/figment-networks/polkadothub-indexer/repos/extrinsicseqrepo"
	"github.com/figment-networks/polkadothub-indexer/repos/reportrepo"
	"github.com/figment-networks/polkadothub-indexer/repos/syncablerepo"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecases/pipeline/startheightpipeline"
	"github.com/figment-networks/polkadothub-indexer/utils/log"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	BatchSize = "batchSize"
)

var (
	rootCmd   *cobra.Command
	batchSize int64
)

func main() {
	//defer errors.RecoverError()

	// CLIENTS
	proxy := shared.NewProxyClient()
	defer proxy.Client().Close()

	db := shared.NewDbClient()
	defer db.Client().Close()

	// REPOSITORIES
	syncableDbRepo := syncablerepo.NewDbRepo(db.Client())
	syncableProxyRepo := syncablerepo.NewProxyRepo(
		blockpb.NewBlockServiceClient(proxy.Client()),
		transactionpb.NewTransactionServiceClient(proxy.Client()),
	)
	reportDbRepo := reportrepo.NewDbRepo(db.Client())

	blockSeqDbRepo := blockseqrepo.NewDbRepo(db.Client())
	extrinsicSeqDbRepo := extrinsicseqrepo.NewDbRepo(db.Client())

	//USE CASES
	startHeightPipelineUseCase := startheightpipeline.NewUseCase(
		syncableDbRepo,
		syncableProxyRepo,
		blockSeqDbRepo,
		extrinsicSeqDbRepo,
		reportDbRepo,
	)

	// HANDLERS
	startHeightPipelineCliHandler := startheightpipeline.NewCliHandler(startHeightPipelineUseCase)

	// CLI COMMANDS
	rootCmd = setupRootCmd()
	pipelineCmd := setupPipelineCmd(startHeightPipelineCliHandler)

	rootCmd.AddCommand(pipelineCmd)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

/*************** Private ***************/

func setupRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cli",
		Short: "Short description",
		Long: `Longer description.. 
            feel free to use a few lines here.
            `,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Usage(); err != nil {
				log.Error(err, log.Field("type", "cli"))
			}
		},
	}
}

func setupPipelineCmd(handler types.CliHandler) *cobra.Command {
	pipelineCmd := &cobra.Command{
		Use:   "pipeline [command]",
		Short: "Run pipeline commands",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Usage(); err != nil {
				log.Error(err, log.Field("type", "cli"))
			}
		},
	}

	startPipelineCmd := &cobra.Command{
		Use:   "start_height",
		Short: "Start one off height processing pipeline",
		Args:  cobra.MaximumNArgs(1),
		Run:   handler.Handle,
	}
	rootCmd.PersistentFlags().Int64Var(&batchSize, BatchSize, config.PipelineBatchSize(), "batch size")
	viper.BindPFlag(BatchSize, rootCmd.PersistentFlags().Lookup(BatchSize))
	pipelineCmd.AddCommand(startPipelineCmd)
	return pipelineCmd
}