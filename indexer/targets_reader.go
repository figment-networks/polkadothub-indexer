package indexer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/figment-networks/indexing-engine/pipeline"
	"io/ioutil"
)

const (
	TargetIndexBlockSequences = iota + 1
	TargetIndexValidatorSequences
	TargetIndexValidatorAggregates
)

// NewTargetsReader constructor for targetsReader
func NewTargetsReader(file string) (*targetsReader, error) {
	p := &targetsReader{}

	cfg, err := p.parseFile(file)
	if err != nil {
		return nil, err
	}

	p.cfg = cfg

	return p, nil
}

// targetsReader
type targetsReader struct {
	cfg *targetsCfg
}

type targetsCfg struct {
	Version int64    `json:"version"`
	Targets []target `json:"targets"`
}

type target struct {
	ID    int64               `json:"id"`
	Name  string              `json:"name"`
	Desc  string              `json:"desc"`
	Tasks []pipeline.TaskName `json:"tasks"`
}

func (p *targetsReader) GetCurrentVersion() int64 {
	return p.cfg.Version
}

// GetAll get lists of tasks for all files
func (p *targetsReader) GetAllTasks() []pipeline.TaskName {
	var allTaskNames []pipeline.TaskName
	for _, t := range p.cfg.Targets {
		allTaskNames = append(allTaskNames, t.Tasks...)
	}
	return uniqueStr(allTaskNames)
}

// GetTasksByTargetIds get lists of tasks for specific target ids
func (p *targetsReader) GetTasksByTargetIds(targetIds []int64) ([]pipeline.TaskName, error) {
	var allTaskNames []pipeline.TaskName
	for _, t := range targetIds {
		tasks, err := p.GetTasksByTargetId(t)
		if err != nil {
			return nil, err
		}
		allTaskNames = append(allTaskNames, tasks...)
	}
	return uniqueStr(allTaskNames), nil
}

// GetById get list of tasks for desired version file
func (p *targetsReader) GetTasksByTargetId(targetId int64) ([]pipeline.TaskName, error) {
	for _, t := range p.cfg.Targets {
		if t.ID == targetId {
			return uniqueStr(t.Tasks), nil
		}
	}
	return nil, errors.New(fmt.Sprintf("target id %d does not exists", targetId))
}

// parseFile gets tasks from json files from given directory
func (p *targetsReader) parseFile(file string) (*targetsCfg, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var cfg targetsCfg
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// UniqueStr return slice with unique elements
func uniqueStr(slice []pipeline.TaskName) []pipeline.TaskName {
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
