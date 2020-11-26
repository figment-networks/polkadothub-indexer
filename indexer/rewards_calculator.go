package indexer

import (
	"fmt"
	"math/big"

	"github.com/figment-networks/polkadothub-indexer/types/perbill"
)

type RewardsCalculator interface {
	commissionPayout(validatorRewardPoints, validatorCommission int64) (big.Int, big.Int)
	nominatorPayout(validatorLeftoverPayout, nominatorStake, validatorStake big.Int)
}

type calculator struct {
	totalEraRewardPoints big.Int
	totalEraRewardPayout big.Int
}

func newRewardsCalulator(points int64, payout string) (*calculator, error) {
	totalEraRewardPayout := new(big.Int)
	totalEraRewardPayout, ok := totalEraRewardPayout.SetString(payout, 10)
	if !ok {
		return nil, fmt.Errorf("totalEraRewardPayout '%v' is an invalid quantity", payout)
	}
	totalEraRewardPoints := *big.NewInt(points)
	return &calculator{totalEraRewardPoints, *totalEraRewardPayout}, nil
}

func (c *calculator) nominatorPayout(validatorLeftoverPayout, nominatorStake, validatorStake big.Int) big.Int {
	exposurePart := perbill.FromRationalApproximation(nominatorStake, validatorStake)
	payout := exposurePart.Mul(validatorLeftoverPayout)
	return payout
}

func (c *calculator) commissionPayout(validatorRewardPoints, validatorCommission int64) (big.Int, big.Int) {
	rewardPoints := *big.NewInt((validatorRewardPoints))
	totalRewardPart := perbill.FromRationalApproximation(rewardPoints, c.totalEraRewardPoints)

	totalPayout := totalRewardPart.Mul(c.totalEraRewardPayout)
	commission := perbill.FromParts(validatorCommission)
	commissionPayout := commission.Mul(totalPayout)

	// leftoverPayouts is divided between stakers
	leftoverPayout := big.Int{}
	leftoverPayout.Sub(&totalPayout, &commissionPayout)

	return commissionPayout, leftoverPayout
}
