package indexer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/model"
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
	IsLastInSession() bool
	IsLastInEra() bool
	GetTransactionKinds() []model.TransactionKind
	GetAllVersionedVersionIds() []int64
	IsAnyVersionSequential(versionIds []int64) bool
	GetAllTasks(verionIds, targetIds []int64) ([]pipeline.TaskName, error)
}

type indexerConfig struct {
	Versions          []version           `json:"versions"`
	SharedTasks       []pipeline.TaskName `json:"shared_tasks"`
	IncompatibleTasks []incompatible      `json:"incompatible_tasks"`
	AvailableTargets  []target            `json:"available_targets"`
}

type version struct {
	ID            int64                   `json:"id"`
	Targets       []int64                 `json:"targets"`
	Parallel      bool                    `json:"parallel"`
	LastInSession bool                    `json:"last_in_session"`
	LastInEra     bool                    `json:"last_in_era"`
	TrxKinds      []model.TransactionKind `json:"transaction_kind"`
}

type target struct {
	ID    int64               `json:"id"`
	Name  string              `json:"name"`
	Desc  string              `json:"desc"`
	Tasks []pipeline.TaskName `json:"tasks"`
}

type incompatible struct {
	Name      pipeline.TaskName   `json:"name"`
	Blacklist []pipeline.TaskName `json:"blacklist"`
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

func (o *configParser) getCurrentVersion() version {
	return o.targets.Versions[len(o.targets.Versions)-1]
}

//GetCurrentVersionId gets the most recent version id
func (o *configParser) GetCurrentVersionId() int64 {
	return o.getCurrentVersion().ID
}

//IsLastInSession check if this version is for last of sessions
func (o *configParser) IsLastInSession() bool {
	return o.getCurrentVersion().LastInSession
}

//IsLastInEra check if this version is for last of eras
func (o *configParser) IsLastInEra() bool {
	return o.getCurrentVersion().LastInEra
}

//GetTransactionKinds gets transaction kinds info
func (o *configParser) GetTransactionKinds() []model.TransactionKind {
	return o.getCurrentVersion().TrxKinds
}

// GetAllAvailableTasks get lists of tasks for all available targets
func (o *configParser) GetAllAvailableTasks() []pipeline.TaskName {
	var allAvailableTaskNames []pipeline.TaskName

	allAvailableTaskNames = o.appendSharedTasks(allAvailableTaskNames)

	for _, t := range o.targets.AvailableTargets {
		allAvailableTaskNames = append(allAvailableTaskNames, t.Tasks...)
	}

	return o.getUniqueTaskNames(allAvailableTaskNames)
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

func (o *configParser) GetAllTasks(verionIds, targetIds []int64) ([]pipeline.TaskName, error) {
	var allTasks []pipeline.TaskName

	if len(verionIds) > 0 {
		tasks, err := o.getTasksByVersionIds(verionIds)
		if err != nil {
			return nil, err
		}
		allTasks = append(allTasks, tasks...)
	}

	if len(targetIds) > 0 {
		tasks, err := o.getTasksByTargetIds(targetIds)
		if err != nil {
			return nil, err
		}
		allTasks = append(allTasks, tasks...)
	}

	return o.getUniqueTaskNames(allTasks), nil
}

// GetAllVersionedTasks get lists of tasks for provided versions
func (o *configParser) GetAllVersionedTasks() ([]pipeline.TaskName, error) {
	var allAvailableTaskNames []pipeline.TaskName

	allAvailableTaskNames = o.appendSharedTasks(allAvailableTaskNames)

	ids := o.GetAllVersionedVersionIds()

	versionedTaskNames, err := o.getTasksByVersionIds(ids)
	if err != nil {
		return nil, err
	}

	allAvailableTaskNames = append(allAvailableTaskNames, versionedTaskNames...)

	return o.getUniqueTaskNames(allAvailableTaskNames), nil
}

// GetTasksByTargetIds get lists of tasks for specific version ids
func (o *configParser) getTasksByVersionIds(versionIds []int64) ([]pipeline.TaskName, error) {
	var allTaskNames []pipeline.TaskName

	allTaskNames = o.appendSharedTasks(allTaskNames)

	for _, t := range versionIds {
		tasks, err := o.getTasksByVersionId(t)
		if err != nil {
			return nil, err
		}
		allTaskNames = append(allTaskNames, tasks...)
	}

	return allTaskNames, nil
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

	return o.getTasksByTargetIds(targetIds)
}

// getTasksByTargetIds get lists of tasks for specific target ids
func (o *configParser) getTasksByTargetIds(targetIds []int64) ([]pipeline.TaskName, error) {
	var allTaskNames []pipeline.TaskName

	allTaskNames = o.appendSharedTasks(allTaskNames)

	for _, t := range targetIds {
		tasks, err := o.getTasksByTargetId(t)
		if err != nil {
			return nil, err
		}
		allTaskNames = append(allTaskNames, tasks...)
	}

	return allTaskNames, nil
}

// getTasksByTargetId get list of tasks for desired target id
func (o *configParser) getTasksByTargetId(targetId int64) ([]pipeline.TaskName, error) {
	for _, t := range o.targets.AvailableTargets {
		if t.ID == targetId {
			return t.Tasks, nil
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
func (o *configParser) getUniqueTaskNames(slice []pipeline.TaskName) []pipeline.TaskName {
	keys := make(map[pipeline.TaskName]bool)
	var list []pipeline.TaskName

	for _, entry := range slice {
		keys[entry] = true
	}

	for _, entry := range o.targets.IncompatibleTasks {
		if _, ok := keys[entry.Name]; ok {
			for _, task := range entry.Blacklist {
				if _, ok2 := keys[task]; ok2 {
					delete(keys, task)
				}
			}
		}
	}

	for key := range keys {
		list = append(list, key)
	}

	return list
}
