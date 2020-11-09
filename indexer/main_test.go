package indexer

import (
	"os"
	"testing"

	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

func TestMain(m *testing.M) {
	setup()
	exitVal := m.Run()
	os.Exit(exitVal)
}

func setup() {
	logger.InitTest()
}
