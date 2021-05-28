package apr

import (
	"math/big"

	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

var (
	daysInYear   = big.NewFloat(365)
	decPrecision = 4
)

type DailyApr struct {
	TimeBucket   types.Time     `json:"time_bucket"`
	Era          int64          `json:"era"`
	Bonded       types.Quantity `json:"bonded"`
	TotalRewards types.Quantity `json:"total_rewards"`
	APR          string         `json:"apr"`
	Validator    string         `json:"validator"`
}

func dailyAPR(rewardSeq model.RewardEraSeq, stake types.Quantity) (DailyApr, error) {
	reward, err := types.NewQuantityFromString(rewardSeq.Amount)
	if err != nil {
		return DailyApr{}, err
	}

	r := reward.Clone()
	rValue := new(big.Float).SetInt(&r.Int)
	b := stake.Clone()
	bValue := new(big.Float).SetInt(&b.Int)
	apr := new(big.Float).Quo(rValue, bValue)
	apr = apr.Mul(apr, daysInYear)

	return DailyApr{
		TimeBucket:   rewardSeq.Time,
		Era:          rewardSeq.Era,
		Bonded:       stake,
		TotalRewards: reward,
		APR:          apr.Text('f', decPrecision),
		Validator:    rewardSeq.ValidatorStashAccount,
	}, nil
}

func toAPRView(rewardSeqs []model.RewardEraSeq, accountSeqs []model.AccountEraSeq) (res []DailyApr, err error) {
	stakeLookup := make(map[int64]types.Quantity)
	for _, seq := range accountSeqs {
		stakeLookup[seq.Era] = seq.Stake
	}

	var stake types.Quantity
	for _, rewardSeq := range rewardSeqs {
		if rewardSeq.Kind != model.RewardReward {
			continue
		}
		stake = stakeLookup[rewardSeq.Era]

		apr, err := dailyAPR(rewardSeq, stake)
		if err != nil {
			return res, err
		}
		res = append(res, apr)
	}
	return res, nil
}
