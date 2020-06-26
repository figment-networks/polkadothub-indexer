package indexer

const (
	AccountAggCreatorTaskName   = "AccountAggCreator"
	ValidatorAggCreatorTaskName = "ValidatorAggCreator"
)

var (
	//_ pipeline.Task = (*validatorAggCreatorTask)(nil)
)



//func NewValidatorAggCreatorTask(db *store.Store) *validatorAggCreatorTask {
//	return &validatorAggCreatorTask{
//		db: db,
//	}
//}
//
//type validatorAggCreatorTask struct {
//	db *store.Store
//}
//
//func (t *validatorAggCreatorTask) GetName() string {
//	return ValidatorAggCreatorTaskName
//}
//
//func (t *validatorAggCreatorTask) Run(ctx context.Context, p pipeline.Payload) error {
//	defer metric.LogIndexerTaskDuration(time.Now(), t.GetName())
//
//	payload := p.(*payload)
//
//	logger.Info(fmt.Sprintf("running indexer task [stage=%s] [task=%s] [height=%d]", pipeline.StageAggregator, t.GetName(), payload.CurrentHeight))
//
//	var newValidatorAggs []model.ValidatorAgg
//	var updatedValidatorAggs []model.ValidatorAgg
//	for _, rawValidator := range payload.RawValidators {
//		existing, err := t.db.ValidatorAgg.FindByEntityUID(rawValidator.GetNode().GetEntityId())
//		if err != nil {
//			if err == store.ErrNotFound {
//				// Create new
//				validator := model.ValidatorAgg{
//					Aggregate: &model.Aggregate{
//						StartedAtHeight: payload.Syncable.Height,
//						StartedAt:       payload.Syncable.Time,
//						RecentAtHeight:  payload.Syncable.Height,
//						RecentAt:        payload.Syncable.Time,
//					},
//
//					EntityUID:               rawValidator.GetNode().EntityId,
//					RecentAddress:           rawValidator.GetAddress(),
//					RecentVotingPower:       rawValidator.GetVotingPower(),
//					RecentAsValidatorHeight: payload.Syncable.Height,
//				}
//
//				parsedValidator, ok := payload.ParsedValidators[rawValidator.GetNode().GetEntityId()]
//				if ok {
//					validator.RecentTotalShares = parsedValidator.TotalShares
//
//					if parsedValidator.PrecommitBlockIdFlag == 1 {
//						// Not validated
//						validator.AccumulatedUptime = 0
//						validator.AccumulatedUptimeCount = 1
//					} else if parsedValidator.PrecommitBlockIdFlag == 2 {
//						// Validated
//						validator.AccumulatedUptime = 1
//						validator.AccumulatedUptimeCount = 1
//					} else {
//						// Nil validated
//						validator.AccumulatedUptime = 0
//						validator.AccumulatedUptimeCount = 0
//					}
//
//					if parsedValidator.Proposed {
//						validator.RecentProposedHeight = payload.CurrentHeight
//						validator.AccumulatedProposedCount = 1
//					}
//				}
//
//				newValidatorAggs = append(newValidatorAggs, validator)
//			} else {
//				return err
//			}
//		} else {
//			// Update
//			validator := model.ValidatorAgg{
//				Aggregate: &model.Aggregate{
//					RecentAtHeight: payload.Syncable.Height,
//					RecentAt:       payload.Syncable.Time,
//				},
//
//				RecentAddress:           rawValidator.GetAddress(),
//				RecentVotingPower:       rawValidator.GetVotingPower(),
//				RecentAsValidatorHeight: payload.Syncable.Height,
//			}
//
//			parsedValidator, ok := payload.ParsedValidators[rawValidator.GetNode().GetEntityId()]
//			if ok {
//				validator.RecentTotalShares = parsedValidator.TotalShares
//
//				if parsedValidator.PrecommitBlockIdFlag == 1 {
//					// Not validated
//					validator.AccumulatedUptime = existing.AccumulatedUptime
//					validator.AccumulatedUptimeCount = existing.AccumulatedUptimeCount + 1
//				} else if parsedValidator.PrecommitBlockIdFlag == 2 {
//					// Validated
//					validator.AccumulatedUptime = existing.AccumulatedUptime + 1
//					validator.AccumulatedUptimeCount = existing.AccumulatedUptimeCount + 1
//				} else {
//					// Validated nil
//					validator.AccumulatedUptime = existing.AccumulatedUptime
//					validator.AccumulatedUptimeCount = existing.AccumulatedUptimeCount
//				}
//
//				if parsedValidator.Proposed {
//					validator.RecentProposedHeight = payload.Syncable.Height
//					validator.AccumulatedProposedCount = existing.AccumulatedProposedCount + 1
//				} else {
//					validator.RecentProposedHeight = existing.RecentProposedHeight
//					validator.AccumulatedProposedCount = existing.AccumulatedProposedCount
//				}
//			}
//
//			existing.Update(validator)
//
//			updatedValidatorAggs = append(updatedValidatorAggs, *existing)
//		}
//	}
//	payload.NewAggregatedValidators = newValidatorAggs
//	payload.UpdatedAggregatedValidators = updatedValidatorAggs
//	return nil
//}
