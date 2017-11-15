package main

import "testing"
import "github.com/stretchr/testify/assert"

type batchLineTest struct {
	line       string
	name       string
	input      []int
	output     []int
	cycleLimit int
	err        error
}

func TestNewTestCaseFromString(t *testing.T) {
	tests := []batchLineTest{
		batchLineTest{"ABC", "", []int{}, []int{}, 0, invalidTestCase},
		batchLineTest{"a;1,2,3;4;5", "a", []int{1, 2, 3}, []int{4}, 5, nil},
		batchLineTest{"a;1,2,3;;5", "a", []int{1, 2, 3}, []int{}, 5, nil},
		batchLineTest{"name;;4,5;5", "name", []int{}, []int{4, 5}, 5, nil},
		batchLineTest{";1,2,3;;5", "", []int{1, 2, 3}, []int{}, 5, nil},
		batchLineTest{"name;a,d;;5", "", []int{}, []int{}, 5, invalidTestCaseInputs},
		batchLineTest{"name;;a;5", "", []int{}, []int{}, 5, invalidTestCaseOutputs},
		batchLineTest{"name;;;a", "", []int{}, []int{}, 5, invalidTestCaseCycles},
	}
	for _, c := range tests {
		tc, err := newTestCaseFromString(c.line)
		assert.Equal(t, c.err, err, c.line)
		if err == nil {
			assert.Equal(t, tc.name, c.name)
			assert.Equal(t, tc.input, c.input)
			assert.Equal(t, tc.output, c.output)
			assert.Equal(t, tc.cycleLimit, c.cycleLimit)
		}
	}
}
