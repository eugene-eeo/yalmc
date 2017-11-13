package main

import "strconv"
import "errors"
import "strings"
import "bufio"
import "io"
import "fmt"

var outOfCycles = errors.New("out of cycles")

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
	for _, s := range strs {
		n, err := stoi(strings.TrimSpace(s), 999)
		if err != nil {
			return nil, fmt.Errorf("cannot convert %s to int: %e", s, err)
		}
		b = append(b, n)
	}
	return b, nil
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
		line = strings.SplitN(line, "#", 2)[0]
		if len(line) == 0 {
			continue
		}
		contents := strings.SplitN(line, ";", 4)
		if len(contents) != 4 {
			errors = append(errors, newError(lineNo, "invalid test case"))
			continue
		}
		name := contents[0]
		inputs, err := inputsToInts(strings.Split(contents[1], ","))
		if err != nil {
			errors = append(errors, newError(lineNo, "invalid inputs"))
			continue
		}
		outputs, err := inputsToInts(strings.Split(contents[2], ","))
		if err != nil {
			errors = append(errors, newError(lineNo, "invalid expected outputs"))
			continue
		}
		cycles, err := strconv.Atoi(contents[3])
		if err != nil {
			errors = append(errors, newError(lineNo, "invalid cycle count"))
			continue
		}
		cases = append(cases, testCase{
			name:       name,
			input:      inputs,
			output:     outputs,
			cycleLimit: cycles,
		})
	}
	return
}
