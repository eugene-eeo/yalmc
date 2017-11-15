package main

import "strconv"
import "errors"
import "strings"
import "bufio"
import "io"
import "fmt"

var outOfCycles = errors.New("out of cycles")
var invalidTestCase = errors.New("invalid test case")
var invalidTestCaseInputs = errors.New("invalid inputs")
var invalidTestCaseOutputs = errors.New("invalid outputs")
var invalidTestCaseCycles = errors.New("invalid cycles")

type testCase struct {
	name       string
	input      []int
	output     []int
	cycleLimit int
}

type testResult struct {
	tcase      testCase
	output     []int
	cycles     int
	terminated bool
}

func isliceEq(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, x := range a {
		if b[i] != x {
			return false
		}
	}
	return true
}

func (t *testResult) failed() bool {
	return t.terminated || !isliceEq(t.tcase.output, t.output)
}

func runWith(vm *context, t *testCase) (r testResult) {
	var err error = nil
	cycles := 0
	vm.input = t.input
	for !vm.halted {
		if cycles == t.cycleLimit {
			err = outOfCycles
			break
		}
		cycles++
		err = vm.fetchExecute()
		if err != nil {
			break
		}
	}
	r.cycles = cycles
	r.tcase = *t
	r.output = vm.output
	r.terminated = (err != nil)
	return
}

func batch(workers int, code []int, cases []testCase) []testResult {
	src := make(chan testCase, len(cases))
	dst := make(chan testResult, (len(cases)/4)+1)
	for _, t := range cases {
		src <- t
	}
	close(src)
	for w := 0; w < workers; w++ {
		go func() {
			vm := newContextFromSlice(code)
			for t := range src {
				dst <- runWith(vm, &t)
				vm.reset()
			}
		}()
	}
	res := make([]testResult, len(cases))
	for i := 0; i < len(cases); i++ {
		res[i] = <-dst
	}
	return res
}

func inputsToInts(strs []string) ([]int, error) {
	b := []int{}
	for i, s := range strs {
		if i == 0 && s == "" {
			continue
		}
		n, err := stoi(strings.TrimSpace(s), 999)
		if err != nil {
			return nil, fmt.Errorf("cannot convert '%s' to int: %s", s, err)
		}
		b = append(b, n)
	}
	return b, nil
}

func newTestCaseFromString(s string) (*testCase, error) {
	s = strings.SplitN(s, "#", 2)[0]
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return nil, nil
	}
	contents := strings.SplitN(s, ";", 4)
	if len(contents) != 4 {
		return nil, invalidTestCase
	}
	inputs, err := inputsToInts(strings.Split(contents[1], ","))
	if err != nil {
		return nil, invalidTestCaseInputs
	}
	outputs, err := inputsToInts(strings.Split(contents[2], ","))
	if err != nil {
		return nil, invalidTestCaseOutputs
	}
	cycles, err := strconv.Atoi(contents[3])
	if err != nil {
		return nil, invalidTestCaseCycles
	}
	return &testCase{
		name:       contents[0],
		input:      inputs,
		output:     outputs,
		cycleLimit: cycles,
	}, nil
}

func parseBatch(r io.Reader) (cases []testCase, errors []error) {
	// Batch file format:
	// # comment allowed
	// Name;Inputs;Outputs;Cycle Limit
	lineNo := 0
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		t, err := newTestCaseFromString(line)
		if err != nil {
			errors = append(errors, newError(lineNo, err.Error()))
			continue
		}
		if t == nil {
			continue
		}
		cases = append(cases, *t)
	}
	return
}
