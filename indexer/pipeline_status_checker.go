package indexer

import (
	"errors"
	"fmt"

	"github.com/figment-networks/polkadothub-indexer/store"
)

// pipelineStatusChecker checks if index version is up to date and what index versions are missing (are not up to date)
type pipelineStatusChecker struct {
	syncablesDb         store.Syncables
	currentIndexVersion int64
}

type pipelineStatus struct {
	// isPristine true when database is empty
	isPristine bool
	// isUpToDate true when all syncables index version is the same as current index version
	isUpToDate        bool
	missingVersionIds []int64
}

func (o *pipelineStatusChecker) getStatus() (*pipelineStatus, error) {
	var startIndexVersion int64
	var isUpToDate bool
	var isPristine bool

	smallestIndexVersion, err := o.syncablesDb.FindSmallestIndexVersion()
	if err != nil {
		if err == store.ErrNotFound {
			// When syncables not found in databases, set start version to first
			startIndexVersion = 1
			isUpToDate = false
			isPristine = true
		} else {
			return nil, err
		}
	} else {
		if *smallestIndexVersion < o.currentIndexVersion {
			// There are records with smaller index versions
			// Include tasks from missing versions
			startIndexVersion = *smallestIndexVersion + 1
			isUpToDate = false
		} else if *smallestIndexVersion == o.currentIndexVersion {
			// When everything is up to date, set start version to first
			startIndexVersion = 1
			isUpToDate = true
		} else {
			return nil, errors.New(fmt.Sprintf("current index version %d is too small", o.currentIndexVersion))
		}
	}

	return &pipelineStatus{
		isPristine:        isPristine,
		isUpToDate:        isUpToDate,
		missingVersionIds: o.getVersionIdsFrom(startIndexVersion),
	}, nil
}

func (o *pipelineStatusChecker) getVersionIdsFrom(startIndexVersion int64) []int64 {
	var ids []int64
	for i := startIndexVersion; i <= o.currentIndexVersion; i++ {
		ids = append(ids, i)
	}
	return ids
}
