package indexer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/pkg/errors"
)

const (
	TargetIndexBlockSequences = iota + 1
	TargetIndexValidatorSessionSequences
	TargetIndexValidatorEraSequences
	TargetIndexValidatorAggregates
	TargetIndexEventSequences
)

var (
	_ ConfigParser = (*configParser)(nil)
)

type ConfigParser interface {
	GetCurrentVersionId() int64
	IsForLastOfSessionsByVersionId(versionId int64) bool
	IsForLastOfErasByVersionId(versionId int64) bool
	GetAllVersionedVersionIds() []int64
	IsAnyVersionSequential(versionIds []int64) bool
	GetAllAvailableTasks() []pipeline.TaskName
	GetAllVersionedTasks() ([]pipeline.TaskName, error)
	GetTasksByVersionIds([]int64) ([]pipeline.TaskName, error)
	GetTasksByTargetIds([]int64) ([]pipeline.TaskName, error)
}

type indexerConfig struct {
	Versions         []version           `json:"versions"`
	SharedTasks      []pipeline.TaskName `json:"shared_tasks"`
	AvailableTargets []target            `json:"available_targets"`
}

type version struct {
	ID            int64   `json:"id"`
	Targets       []int64 `json:"targets"`
	Parallel      bool    `json:"parallel"`
	LastInSession bool    `json:"last_in_session"`
	LastInEra     bool    `json:"last_in_era"`
}

type target struct {
	ID    int64               `json:"id"`
	Name  string              `json:"name"`
	Desc  string              `json:"desc"`
	Tasks []pipeline.TaskName `json:"tasks"`
}

func NewConfigParser(file string) (*configParser, error) {
	o := &configParser{
		file: file,
	}

	tr, err := o.Parse()
	if err != nil {
		return nil, err
	}

	o.targets = tr

	return o, nil
}

type configParser struct {
	file    string
	targets *indexerConfig
}

func (o *configParser) Parse() (*indexerConfig, error) {
	data, err := ioutil.ReadFile(o.file)
	if err != nil {
		return nil, err
	}

	var tgs indexerConfig
	err = json.Unmarshal(data, &tgs)
	if err != nil {
		return nil, err
	}

	return &tgs, nil
}

//GetCurrentVersionId gets the most recent version id
func (o *configParser) GetCurrentVersionId() int64 {
	lastVersion := o.targets.Versions[len(o.targets.Versions)-1]
	return lastVersion.ID
}

//IsForLastOfSessionsByVersionId check if this version is for last of sessions
func (o *configParser) IsForLastOfSessionsByVersionId(versionId int64) bool {
	lastVersion := o.targets.Versions[versionId]
	return lastVersion.LastInSession
}

//IsForLastOfEraByVersionId check if this version is for last of eras
func (o *configParser) IsForLastOfErasByVersionId(versionId int64) bool {
	lastVersion := o.targets.Versions[versionId]
	return lastVersion.LastInEra
}

// GetAllAvailableTasks get lists of tasks for all available targets
func (o *configParser) GetAllAvailableTasks() []pipeline.TaskName {
	var allAvailableTaskNames []pipeline.TaskName

	allAvailableTaskNames = o.appendSharedTasks(allAvailableTaskNames)

	for _, t := range o.targets.AvailableTargets {
		allAvailableTaskNames = append(allAvailableTaskNames, t.Tasks...)
	}

	return getUniqueTaskNames(allAvailableTaskNames)
}

// GetAllVersionedVersionIds gets a slice with all version ids in the targets file
func (o *configParser) GetAllVersionedVersionIds() []int64 {
	currentVersionId := o.GetCurrentVersionId()
	var ids []int64
	for i := int64(1); i <= currentVersionId; i++ {
		ids = append(ids, i)
	}
	return ids
}

// GetAllVersionedTasks get lists of tasks for provided versions
func (o *configParser) GetAllVersionedTasks() ([]pipeline.TaskName, error) {
	var allAvailableTaskNames []pipeline.TaskName

	allAvailableTaskNames = o.appendSharedTasks(allAvailableTaskNames)

	ids := o.GetAllVersionedVersionIds()

	versionedTaskNames, err := o.GetTasksByVersionIds(ids)
	if err != nil {
		return nil, err
	}

	allAvailableTaskNames = append(allAvailableTaskNames, versionedTaskNames...)

	return getUniqueTaskNames(allAvailableTaskNames), nil
}

// GetTasksByTargetIds get lists of tasks for specific version ids
func (o *configParser) GetTasksByVersionIds(versionIds []int64) ([]pipeline.TaskName, error) {
	var allTaskNames []pipeline.TaskName

	allTaskNames = o.appendSharedTasks(allTaskNames)

	for _, t := range versionIds {
		tasks, err := o.getTasksByVersionId(t)
		if err != nil {
			return nil, err
		}
		allTaskNames = append(allTaskNames, tasks...)
	}

	return getUniqueTaskNames(allTaskNames), nil
}

// IsAnyVersionSequential checks if any version in targets file is sequential
func (o *configParser) IsAnyVersionSequential(versionIds []int64) bool {
	for _, v := range o.targets.Versions {
		for _, needleVersionId := range versionIds {
			if v.ID == needleVersionId && !v.Parallel {
				return true
			}
		}
	}

	return false
}

// getTasksByVersionId get lists of tasks for specific version id
func (o *configParser) getTasksByVersionId(versionId int64) ([]pipeline.TaskName, error) {
	var targetIds []int64
	versionFound := false
	for _, version := range o.targets.Versions {
		if version.ID == versionId {
			targetIds = version.Targets
			versionFound = true
		}
	}

	if !versionFound {
		return nil, errors.New(fmt.Sprintf("version %d not found", versionId))
	}

	return o.GetTasksByTargetIds(targetIds)
}

// GetTasksByTargetIds get lists of tasks for specific target ids
func (o *configParser) GetTasksByTargetIds(targetIds []int64) ([]pipeline.TaskName, error) {
	var allTaskNames []pipeline.TaskName

	allTaskNames = o.appendSharedTasks(allTaskNames)

	for _, t := range targetIds {
		tasks, err := o.getTasksByTargetId(t)
		if err != nil {
			return nil, err
		}
		allTaskNames = append(allTaskNames, tasks...)
	}

	return getUniqueTaskNames(allTaskNames), nil
}

// getTasksByTargetId get list of tasks for desired target id
func (o *configParser) getTasksByTargetId(targetId int64) ([]pipeline.TaskName, error) {
	for _, t := range o.targets.AvailableTargets {
		if t.ID == targetId {
			return getUniqueTaskNames(t.Tasks), nil
		}
	}
	return nil, errors.New(fmt.Sprintf("target id %d does not exists", targetId))
}

// appendSharedTasks appends shared tasks
func (o *configParser) appendSharedTasks(tasks []pipeline.TaskName) []pipeline.TaskName {
	tasks = append(tasks, o.targets.SharedTasks...)
	return tasks
}

// getUniqueTaskNames return slice with unique task names
func getUniqueTaskNames(slice []pipeline.TaskName) []pipeline.TaskName {
	keys := make(map[pipeline.TaskName]bool)
	var list []pipeline.TaskName
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
