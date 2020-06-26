package accountagg

import (
	"github.com/jinzhu/gorm"
)

type Account struct {
	gorm.Model

	Address              string `json:"address" gorm:"index:ix_data_account_address"`
	IndexAddress         string `json:"index_address" gorm:"index:ix_data_account_index_address"`
	IsReaped             bool   `json:"is_reaped"`
	IsValidator          bool   `json:"is_validator"`
	IsNominator          bool   `json:"is_nominator"`
	IsContract           bool   `json:"is_contract"`
	CountReaped          int64  `json:"count_reaped"`
	Balance              int64  `json:"count_reaped" gorm:"type:numeric(65, 0)"`
	HashBlake2B          string `json:"hash_blake2b" gorm:"null; index:ix_data_account_hash_blake2b"`
	IdentityDisplay      string `json:"identity_display" gorm:"null; index:ix_data_account_identity_display"`
	IdentityLegal        string `json:"identity_legal" gorm:"null"`
	IdentityWeb          string `json:"identity_web" gorm:"null"`
	IdentityRiot         string `json:"identity_riot" gorm:"null"`
	IdentityEmail        string `json:"identity_email" gorm:"null"`
	IdentityTwitter      string `json:"identity_twitter" gorm:"null"`
	IdentityJudgmentGood int64  `json:"identity_judgment_good" gorm:"null"`
	IdentityJudgmentBad  int64  `json:"identity_judgment_bad" gorm:"null"`
	CreatedAtBlock       int64  `json:"created_at_block" gorm:"not null"`
	UpdatedAtBlock       int64  `json:"updated_at_block" gorm:"not null"`
}
