package job

import (
	"github.com/figment-networks/polkadothub-indexer/apps/shared"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
	"github.com/figment-networks/polkadothub-indexer/utils/log"
	"github.com/robfig/cron/v3"
)

var (
	cronJob *cron.Cron
	cronLog = log.NewCronLogger()
	job     cron.Job

)

func main() {
	defer errors.RecoverError()

	// CLIENTS
	node := shared.NewNodeClient()
	db := shared.NewDbClient()

	// CRON
	cronJob = cron.New(
		cron.WithLogger(cron.VerbosePrintfLogger(log.GetLogger())),
		cron.WithChain(
			cron.Recover(cronLog),
		),
	)

	log.Info("starting cron job")

	cronJob.Start()

	//Run forever
	select {}
}
