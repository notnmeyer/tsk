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

func TestConvertEnvToStringSlice(t *testing.T) {
	env := map[string]string{
		"FOO": "bar",
	}

	expected := []string{"FOO=bar"}
	actual := ConvertEnvToStringSlice(env)

	if len(actual) != len(expected) {
		t.Errorf("Expected %d, got %d", len(expected), len(actual))
	}

	for i, v := range actual {
		if v != expected[i] {
			t.Errorf("Expected %s, got %s", expected[i], v)
		}
	}
}
