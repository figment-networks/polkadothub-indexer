package nominator

import (
	"github.com/jinzhu/gorm"
)

type SessionNominator struct {
	gorm.Model

	SessionId           int64  `json:"session_id" gorm:"null"`
	RankValidator       int64  `json:"rank_validator" gorm:"null; index:ix_data_session_nominator_rank_validator"`
	RankNominator       int64  `json:"rank_nominator" gorm:"null; index:ix_data_session_nominator_rank_nominator"`
	NominatorStash      string `json:"nominator_stash" gorm:"null; type:varchar(64); index:ix_data_session_nominator_nominator_stash"`
	NominatorController string `json:"nominator_controller" gorm:"null; type:varchar(64); index:ix_data_session_nominator_nominator_controller"`
	Bonded              int64  `json:"bonded" gorm:"type:numeric(65, 0)"`
}
