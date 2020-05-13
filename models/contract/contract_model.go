package contract

import (
	"github.com/jinzhu/gorm"
)

type Contract struct {
	gorm.Model

	CodeHash string `json:"code_hash" gorm:"not null; type:varchar(64)"`
	Bytecode string `json:"bytecode" gorm:"null; type:text"`
	Source string `json:"source" gorm:"null; type:text"`
	Abi string `json:"abi" gorm:"null; type:json"`
	Compiler string `json:"compiler" gorm:"not null; type:varchar(64)"`
	CreatedAtBlock int64 `json:"created_at_block" gorm:"null"`
	CreatedAtExtrinsic int64 `json:"created_at_extrinsic" gorm:"null"`
	CreatedAtEvent int64 `json:"created_at_event" gorm:"null"`
}
