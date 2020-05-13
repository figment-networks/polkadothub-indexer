package validator

import (
	"github.com/jinzhu/gorm"
)

type SessionValidator struct {
	gorm.Model

	SessionId           int64  `json:"session_id" gorm:"null"`
	RankValidator       int64  `json:"rank_validator" gorm:"null; index:ix_data_session_validator_rank_validator"`
	ValidatorStash      string `json:"validator_stash" gorm:"null; type:varchar(64); index:ix_data_session_validator_validator_stash"`
	ValidatorController string `json:"validator_controller" gorm:"null; type:varchar(64); index:ix_data_session_validator_validator_controller"`
	ValidatorSession    string `json:"validator_session" gorm:"null; type:varchar(64); index:ix_data_session_validator_validator_session"`
	BondedTotal         int64  `json:"bonded_total" gorm:"type:numeric(65, 0)"`
	BondedActive        int64  `json:"bonded_active" gorm:"type:numeric(65, 0)"`
	BondedNominators    int64  `json:"bonded_nominators" gorm:"type:numeric(65, 0)"`
	BondedOwn           int64  `json:"bonded_own" gorm:"type:numeric(65, 0)"`
	Commission          int64  `json:"commission" gorm:"type:numeric(65, 0)"`
	Unlocking           string `json:"unlocking" gorm:"type:json"`
	CountNominators     int64  `json:"count_nominators" gorm:"not null"`
	UnstakeThreshold     int64  `json:"unstake_threshold" gorm:"not null"`
}
