package session

import (
	"github.com/jinzhu/gorm"
)

type Session struct {
	gorm.Model

	StartAtBlock       int64 `json:"start_at_block" gorm:"null"`
	Era                int64 `json:"era" gorm:"null"`
	EraIdx             int64 `json:"era_idx" gorm:"null"`
	CreatedAtBlock     int64 `json:"crated_at_block" gorm:"null"`
	CreatedAtExtrinsic int64 `json:"created_at_extrinsic" gorm:"null"`
	CreatedAtEvent     int64 `json:"created_at_event" gorm:"null"`
	CountValidators    int64 `json:"count_validators" gorm:"null"`
	CountNominators    int64 `json:"count_nominators" gorm:"null"`
}
