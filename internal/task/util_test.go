package task

import "testing"

func TestAlphabetizeTaskList(t *testing.T) {
	taskList := *alphabetizeTaskList(&map[string]Task{
		"c": {}, "b": {}, "a": {},
	})
	expected := []string{"a", "b", "c"}

	if compareSlices(taskList, expected) != true {
		t.Errorf("Expected %v, got %v", expected, taskList)
	}
}

func compareSlices(a, b []string) bool {
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
