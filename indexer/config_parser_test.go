package indexer

import (
	"fmt"
	"testing"

	"github.com/figment-networks/polkadothub-indexer/utils/test"
)

func TestConfigParser(t *testing.T) {
	t.Run("returns error when file is bad", func(t *testing.T) {
		fileName := "test_indexer_config.json"
		var targetsJsonBlob = []byte(`this is not a JSON`)

		test.CreateFile(t, fileName, targetsJsonBlob)
		defer test.CleanUp(t, fileName)

		_, err := NewConfigParser(fileName)
		if err == nil {
			t.Errorf("NewConfigParser should return error")
		}
	})
}

func TestConfigParser_GetAllTasks(t *testing.T) {
	t.Run("returns unique tasks for all targets", func(t *testing.T) {
		fileName := "test_indexer_config.json"
		var targetsJsonBlob = []byte(`
			{
			  "available_targets": [
				{
				  "id": 1,
				  "name": "target1",
				  "desc": "Test target1",
				  "tasks": [
					"Task1",
					"Task2"
				  ]
				},
				{
				  "id": 2,
				  "name": "target2",
				  "desc": "Test target2",
				  "tasks": [
					"Task1",
					"Task2",
					"Task3",
					"Task4",
					"Task5"
				  ]
				}
			  ]
			}
    	`)

		test.CreateFile(t, fileName, targetsJsonBlob)
		defer test.CleanUp(t, fileName)

		parser, err := NewConfigParser(fileName)
		if err != nil {
			t.Errorf("should not return error: err=%+v", err)
			return
		}

		tasks := parser.GetAllAvailableTasks()

		if len(tasks) != 5 {
			t.Errorf("unexpected tasks length, want: %d; got: %d", 5, len(tasks))
		}

		for i := 0; i < len(tasks); i++ {
			taskName := fmt.Sprintf("Task%d", i+1)
			if string(tasks[i]) != taskName {
				t.Errorf("unexpected task at index %d, want: %s, got: %s", i, taskName, tasks[i])
			}
		}
	})

	t.Run("returns tasks including shared tasks", func(t *testing.T) {
		fileName := "test_indexer_config.json"
		var targetsJsonBlob = []byte(`
			{
			  "shared_tasks": [
				"SharedTask1",
				"SharedTask2"
			  ],
			  "available_targets": [
				{
				  "id": 1,
				  "name": "target1",
				  "desc": "Test target1",
				  "tasks": [
					"Task1",
					"Task2"
				  ]
				},
				{
				  "id": 2,
				  "name": "target2",
				  "desc": "Test target2",
				  "tasks": [
					"Task1",
					"Task2",
					"Task3",
					"Task4",
					"Task5"
				  ]
				}
			  ]
			}
    	`)

		test.CreateFile(t, fileName, targetsJsonBlob)
		defer test.CleanUp(t, fileName)

		parser, err := NewConfigParser(fileName)
		if err != nil {
			t.Errorf("should not return error: err=%+v", err)
			return
		}

		tasks := parser.GetAllAvailableTasks()

		if len(tasks) != 7 {
			t.Errorf("unexpected tasks length, want: %d; got: %d", 7, len(tasks))
		}

		if string(tasks[0]) != "SharedTask1" {
			t.Errorf("unexpected task at index %d, want: %s, got: %s", 0, "SharedTask1", tasks[0])
		}

		if string(tasks[1]) != "SharedTask2" {
			t.Errorf("unexpected task at index %d, want: %s, got: %s", 1, "SharedTask2", tasks[1])
		}

		for i := 2; i < len(tasks); i++ {
			taskName := fmt.Sprintf("Task%d", i-1)
			if string(tasks[i]) != taskName {
				t.Errorf("unexpected task at index %d, want: %s, got: %s", i, taskName, tasks[i])
			}
		}
	})
}

func TestConfigParser_GetTasksByVersionIds(t *testing.T) {
	fileName := "test_indexer_config.json"
	var targetsJsonBlob = []byte(`
		{
		  "versions": [
			{
			  "id": 1,
			  "targets": [1, 2]
			},
			{
			  "id": 2,
			  "targets": [3]
			}
		  ],
		  "available_targets": [
			{
			  "id": 1,
			  "name": "target1",
			  "desc": "Test target1",
			  "tasks": [
				"Task1",
				"Task2"
			  ]
			},
			{
			  "id": 2,
			  "name": "target2",
			  "desc": "Test target2",
			  "tasks": [
				"Task1",
				"Task2",
				"Task3",
				"Task4",
				"Task5"
			  ]
			},
			{
			  "id": 3,
			  "name": "target3",
			  "desc": "Test target3",
			  "tasks": [
				"Task1",
				"Task2",
				"Task6",
				"Task7",
				"Task8"
			  ]
			}
		  ]
		}
	`)

	t.Run("returns tasks for given version id", func(t *testing.T) {
		test.CreateFile(t, fileName, targetsJsonBlob)
		defer test.CleanUp(t, fileName)

		parser, err := NewConfigParser(fileName)
		if err != nil {
			t.Errorf("NewConfigParser should not return error: err=%+v", err)
			return
		}

		tasks, err := parser.GetTasksByVersionIds([]int64{1})
		if err != nil {
			t.Errorf("GetTasksForVersion should not return error: err=%+v", err)
			return
		}

		if len(tasks) != 5 {
			t.Errorf("unexpected tasks length, want: %d; got: %d", 5, len(tasks))
		}

		for i := 0; i < len(tasks); i++ {
			taskName := fmt.Sprintf("Task%d", i+1)
			if string(tasks[i]) != taskName {
				t.Errorf("unexpected task at index %d, want: %s, got: %s", i, taskName, tasks[i])
			}
		}
	})

	t.Run("returns error when version is not found", func(t *testing.T) {
		test.CreateFile(t, fileName, targetsJsonBlob)
		defer test.CleanUp(t, fileName)

		parser, err := NewConfigParser(fileName)
		if err != nil {
			t.Errorf("NewConfigParser should not return error: err=%+v", err)
			return
		}

		_, err = parser.GetTasksByVersionIds([]int64{40})
		if err == nil {
			t.Errorf("GetTasksForVersion should return error")
		}
	})
}

func TestConfigParser_GetTasksByTargetIds(t *testing.T) {
	t.Run("returns unique tasks for selected target ids", func(t *testing.T) {
		fileName := "test_indexer_config.json"
		var targetsJsonBlob = []byte(`
			{
			  "available_targets": [
				{
				  "id": 1,
				  "name": "target1",
				  "desc": "Test target1",
				  "tasks": [
					"Task1",
					"Task2"
				  ]
				},
				{
				  "id": 2,
				  "name": "target2",
				  "desc": "Test target2",
				  "tasks": [
					"Task1",
					"Task2",
					"Task3",
					"Task4",
					"Task5"
				  ]
				},
				{
				  "id": 3,
				  "name": "target3",
				  "desc": "Test target3",
				  "tasks": [
					"Task1",
					"Task2",
					"Task6",
					"Task7",
					"Task8"
				  ]
				}
			  ]
			}
    	`)

		test.CreateFile(t, fileName, targetsJsonBlob)
		defer test.CleanUp(t, fileName)

		parser, err := NewConfigParser(fileName)
		if err != nil {
			t.Errorf("NewConfigParser should not return error: err=%+v", err)
			return
		}

		tasks, err := parser.GetTasksByTargetIds([]int64{1, 2})
		if err != nil {
			t.Errorf("GetTasksByTargetIds should not return error: err=%+v", err)
			return
		}

		if len(tasks) != 5 {
			t.Errorf("unexpected tasks length, want: %d; got: %d", 5, len(tasks))
		}

		for i := 0; i < len(tasks); i++ {
			taskName := fmt.Sprintf("Task%d", i+1)
			if string(tasks[i]) != taskName {
				t.Errorf("unexpected task at index %d, want: %s, got: %s", i, taskName, tasks[i])
			}
		}
	})
}

func TestConfigParser_GetCurrentVersion(t *testing.T) {
	t.Run("returns most recent version", func(t *testing.T) {
		fileName := "test_indexer_config.json"
		var targetsJsonBlob = []byte(`
			{
			  "versions": [
			  	{
				  "id": 1,
				  "targets": [1, 2]
				},
				{
				  "id": 2,
				  "targets": [3]
				}
   	          ]
			}
    	`)

		test.CreateFile(t, fileName, targetsJsonBlob)
		defer test.CleanUp(t, fileName)

		parser, err := NewConfigParser(fileName)
		if err != nil {
			t.Errorf("NewConfigParser should not return error: err=%+v", err)
			return
		}

		version := parser.GetCurrentVersionId()
		if version != 2 {
			t.Errorf("unexpected current version, want: %d; got: %d", 2, version)
		}
	})
}

func TestConfigParser_GetAllVersionedVersionIds(t *testing.T) {
	fileName := "test_indexer_config.json"
	var targetsJsonBlob = []byte(`
		{
		  "versions": [
			{
			  "id": 1,
			  "targets": [1, 2]
			},
			{
			  "id": 2,
			  "targets": [3]
			}
		  ]
		}
	`)

	t.Run("returns version ids for version 1 and 2", func(t *testing.T) {
		test.CreateFile(t, fileName, targetsJsonBlob)
		defer test.CleanUp(t, fileName)

		parser, err := NewConfigParser(fileName)
		if err != nil {
			t.Errorf("NewConfigParser should not return error: err=%+v", err)
			return
		}

		ids := parser.GetAllVersionedVersionIds()

		if len(ids) != 2 {
			t.Errorf("unexpected tasks length, want: %d; got: %d", 2, len(ids))
			return
		}

		for i := 0; i < len(ids); i++ {
			if ids[i] != int64(i)+1 {
				t.Errorf("unexpected id at index %d, want: %d, got: %d", i, i+1, ids[i])
			}
		}
	})
}

func TestConfigParser_GetAllVersionedTasks(t *testing.T) {
	fileName := "test_indexer_config.json"
	var targetsJsonBlob = []byte(`
		{
		  "versions": [
			{
			  "id": 1,
			  "targets": [1, 2]
			},
			{
			  "id": 2,
			  "targets": [3]
			}
		  ],
		  "available_targets": [
			{
			  "id": 1,
			  "name": "target1",
			  "desc": "Test target1",
			  "tasks": [
				"Task1",
				"Task2"
			  ]
			},
			{
			  "id": 2,
			  "name": "target2",
			  "desc": "Test target2",
			  "tasks": [
				"Task1",
				"Task2",
				"Task3",
				"Task4",
				"Task5"
			  ]
			},
			{
			  "id": 3,
			  "name": "target3",
			  "desc": "Test target3",
			  "tasks": [
				"Task1",
				"Task2",
				"Task6",
				"Task7",
				"Task8"
			  ]
			},
			{
			  "id": 4,
			  "name": "target4",
			  "desc": "Test target4",
			  "tasks": [
				"Task9",
				"Task10",
				"Task11"
			  ]
			}
		  ]
		}
	`)

	t.Run("returns tasks for targets 1, 2 and 3 from versions and not target 4", func(t *testing.T) {
		test.CreateFile(t, fileName, targetsJsonBlob)
		defer test.CleanUp(t, fileName)

		parser, err := NewConfigParser(fileName)
		if err != nil {
			t.Errorf("NewConfigParser should not return error: err=%+v", err)
			return
		}

		tasks, err := parser.GetAllVersionedTasks()
		if err != nil {
			t.Errorf("GetAllVersionedTasks should not return error: err=%+v", err)
			return
		}

		if len(tasks) != 8 {
			t.Errorf("unexpected tasks length, want: %d; got: %d", 5, len(tasks))
		}

		for i := 0; i < len(tasks); i++ {
			taskName := fmt.Sprintf("Task%d", i+1)
			if string(tasks[i]) != taskName {
				t.Errorf("unexpected task at index %d, want: %s, got: %s", i, taskName, tasks[i])
			}
		}
	})
}

func TestConfigParser_IsAnyVersionSequential(t *testing.T) {
	fileName := "test_indexer_config.json"
	tests := []struct {
		description string
		input       []byte
		expected    bool
	}{
		{
			input:    []byte(`{"versions": [{"id": 1, "parallel": false}, {"id": 2, "parallel": false}]}`),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			test.CreateFile(t, fileName, tt.input)
			defer test.CleanUp(t, fileName)

			parser, err := NewConfigParser(fileName)
			if err != nil {
				t.Errorf("NewConfigParser should not return error: err=%+v", err)
				return
			}

			got := parser.IsAnyVersionSequential([]int64{1, 2})
			if got != tt.expected {
				t.Errorf("unexpected value, want: %v; got: %v", tt.expected, got)
			}
		})
	}
}
