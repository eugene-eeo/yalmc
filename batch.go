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

func runWith(vm *context, t *testCase) (output []int, cycles int, err error) {
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
	output = vm.output
	return
}

func executeCases(vm context, src <-chan testCase, sink chan<- testResult) {
	for t := range src {
		output, cycles, err := runWith(&vm, &t)
		sink <- testResult{
			tcase:      t,
			output:     output,
			cycles:     cycles,
			terminated: err != nil,
		}
		vm.reset()
	}
}

func batch(workers int, code []int, cases []testCase) []testResult {
	casesQueue := make(chan testCase, len(cases))
	resultsQueue := make(chan testResult, (len(cases)/4)+1)
	for _, t := range cases {
		casesQueue <- t
	}
	for w := 0; w < workers; w++ {
		go executeCases(
			*newContextFromSlice(code),
			casesQueue,
			resultsQueue,
		)
	}
	res := make([]testResult, len(cases))
	for i := 0; i < len(cases); i++ {
		res[i] = <-resultsQueue
	}
	return res
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
		if line[0] == '#' {
			continue
		}
		contents := strings.SplitN(line, ";", 4)
		if len(contents) != 4 {
			errors = append(errors, fmt.Errorf("Line %d: Invalid test case"))
			continue
		}
		name := contents[0]
		inputs := mustInt(strings.Split(contents[1], ","))
		outputs := mustInt(strings.Split(contents[2], ","))
		cycles, err := strconv.Atoi(contents[3])
		if err != nil {
			errors = append(errors, fmt.Errorf("Line %d: Invalid cycle count"))
			break
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
