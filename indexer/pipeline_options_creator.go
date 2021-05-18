package indexer

import "github.com/figment-networks/indexing-engine/pipeline"

// pipelineOptionsCreator is responsible for creating pipeline options
type pipelineOptionsCreator struct {
	configParser ConfigParser

	desiredVersionIds []int64
	desiredTargetIds  []int64

	dry bool
}

func (o *pipelineOptionsCreator) parse() (*pipeline.Options, error) {
	taskWhitelist, err := o.configParser.GetAllTasks(o.desiredVersionIds, o.desiredTargetIds)
	if err != nil {
		return nil, err
	}

	return &pipeline.Options{
		TaskWhitelist:   taskWhitelist,
		StagesBlacklist: o.getStagesBlacklist(),
	}, nil
}

func (o *pipelineOptionsCreator) getStagesBlacklist() []pipeline.StageName {
	var stagesBlacklist []pipeline.StageName
	if o.dry {
		stagesBlacklist = append(stagesBlacklist, pipeline.StagePersistor)
	}
	return stagesBlacklist
}
