package main

import "io"
import "bufio"
import "errors"
import "fmt"
import "strings"
import "strconv"

var outOfBounds error = errors.New("Number is > 999")
var instrLookup = map[string]int{
	"ADD": 100,
	"SUB": 200,
	"STO": 300,
	"LDA": 500,
	"BR":  600,
	"BRZ": 700,
	"BRP": 800,
	"IN":  901,
	"OUT": 902,
	"HLT": 000,
	"DAT": -1, // special for DAT
}

func errorMessage(lineNo int, message string) error {
	return fmt.Errorf("Line %d: %s", lineNo, message)
}

type Line struct {
	lineNo int
	label  string
	instr  int
	addr   string
}

func stoi(s string) (int, error) {
	if len(s) > 3 {
		return 0, outOfBounds
	}
	return strconv.Atoi(s)
}

func line_to_int(line *Line, labels map[string]int) (int, error) {
	// HLT / IN / OUT instructions can be on their own without
	// any address component
	if line.instr == 0 || line.instr == 901 || line.instr == 902 {
		return line.instr, nil
	}
	// DAT [xxx], defaults to 0
	if line.instr == -1 {
		if line.addr == "" {
			return 0, nil
		}
		return stoi(line.addr)
	}
	// Instructions other than IN/OUT/HLT need a target address
	// so if we are not given one, error out.
	if line.addr == "" {
		return 0, errorMessage(line.lineNo, "no address given")
	}
	if i, ok := labels[line.addr]; ok {
		return line.instr + i, nil
	}
	i, err := stoi(line.addr)
	return line.instr + i, err
}

func lines_to_int(lines []*Line) ([]int, []error) {
	// Perform 1 pass to first index the positions of the
	// mailboxes in the code so that it is possible to reference
	// a mailbox after/before it is defined
	labels := map[string]int{}
	errors := []error{}
	for mailbox, line := range lines {
		if line.label == "" {
			continue
		}
		labels[line.label] = mailbox
	}
	// Fill up the mailboxes by parsing the instructions
	buff := make([]int, 100)
	for i, line := range lines {
		instr, err := line_to_int(line, labels)
		buff[i] = instr
		if err != nil {
			errors = append(errors, err)
			return nil, errors
		}
	}
	return buff, nil
}

func parseInstruction(s string) (int, error) {
	a, ok := instrLookup[s]
	if !ok {
		return -1, fmt.Errorf("%s: invalid instruction", s)
	}
	return a, nil
}

func stringToLine(lineNo int, s string) (*Line, error) {
	s = strings.SplitN(s, "#", 2)[0]
	if len(strings.TrimSpace(s)) == 0 {
		return nil, nil
	}
	c := strings.SplitN(s, "\t", 3)
	if len(c) < 2 {
		return nil, errorMessage(lineNo, "bad instruction")
	}
	label := strings.TrimSpace(c[0])
	instr := strings.TrimSpace(strings.ToUpper(c[1]))
	addr := ""
	if len(c) == 3 {
		addr = strings.TrimSpace(c[2])
	}
	opcode, err := parseInstruction(instr)
	if err != nil {
		return nil, errorMessage(lineNo, err.Error())
	}
	return &Line{
		label:  label,
		instr:  opcode,
		addr:   addr,
		lineNo: lineNo,
	}, nil
}

func compile(r io.Reader) ([]int, int, []error) {
	i := 0            // current mailbox number
	lineNo := 0       // current line number
	buff := []*Line{} // compile buffer
	errors := []error{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		lineNo++
		line, err := stringToLine(lineNo, s)
		// Empty line / only comments so don't bother incrementing
		// mailbox number
		if err == nil && line == nil {
			continue
		}
		i++
		// Reached mailbox limit, don't bother incrementing
		if i == 100 {
			errors = append(errors, errorMessage(lineNo, "Out of mailboxes"))
			break
		}
		if err != nil {
			errors = append(errors, err)
			continue
		}
		buff = append(buff, line)
	}
	if len(errors) != 0 {
		return nil, 0, errors
	}
	code, errors := lines_to_int(buff)
	return code, i, errors
}
