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
	taskWhitelist, err := o.getTasksWhitelist()
	if err != nil {
		return nil, err
	}

	return &pipeline.Options{
		TaskWhitelist:   taskWhitelist,
		StagesBlacklist: o.getStagesBlacklist(),
	}, nil
}

func (o *pipelineOptionsCreator) getTasksWhitelist() ([]pipeline.TaskName, error) {
	var taskWhitelist []pipeline.TaskName

	if len(o.desiredVersionIds) > 0 {
		tasks, err := o.configParser.GetTasksByVersionIds(o.desiredVersionIds)
		if err != nil {
			return nil, err
		}
		taskWhitelist = append(taskWhitelist, tasks...)
	}

	if len(o.desiredTargetIds) > 0 {
		tasks, err := o.configParser.GetTasksByTargetIds(o.desiredTargetIds)
		if err != nil {
			return nil, err
		}
		taskWhitelist = append(taskWhitelist, tasks...)
	}

	return getUniqueTaskNames(taskWhitelist), nil
}

func (o *pipelineOptionsCreator) getStagesBlacklist() []pipeline.StageName {
	var stagesBlacklist []pipeline.StageName
	if o.dry {
		stagesBlacklist = append(stagesBlacklist, pipeline.StagePersistor)
	}
	return stagesBlacklist
}
