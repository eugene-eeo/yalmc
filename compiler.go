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

type parseError struct {
	line   int
	reason string
}

func newError(line int, reason string) error {
	return parseError{line, reason}
}

func (e parseError) Error() string {
	return fmt.Sprintf("Line %d: %s", e.line, e.reason)
}

func stoi(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if i > 1000 || i < 0 {
		return 0, outOfBounds
	}
	return i, nil
}

type Line struct {
	text   string
	lineNo int
	label  string
	instr  int
	addr   string
}

func newLineFromString(s string) (*Line, error) {
	s = strings.SplitN(s, "#", 2)[0]
	if len(strings.TrimSpace(s)) == 0 {
		return nil, nil
	}
	parts := strings.Fields(s)
	if len(parts) > 3 {
		return nil, fmt.Errorf("trailing content, expected [LABEL] INSTRUCTION [ADDR]")
	}
	if s[0] == ' ' || s[0] == '\t' {
		c := []string{""}
		parts = append(c, parts...)
	}
	if len(parts) < 2 {
		return nil, fmt.Errorf("bad instruction, expected [LABEL] INSTRUCTION [ADDR]")
	}
	label := strings.TrimSpace(parts[0])
	instr := strings.TrimSpace(strings.ToUpper(parts[1]))
	addr := ""
	if len(parts) == 3 {
		addr = strings.TrimSpace(parts[2])
	}
	opcode, err := parseInstruction(instr)
	if err != nil {
		return nil, err
	}
	return &Line{
		text:  s,
		label: label,
		instr: opcode,
		addr:  addr,
	}, nil
}

func (l *Line) toData(labels map[string]int) (int, error) {
	// HLT / IN / OUT instructions can be on their own without
	// any address component
	if l.instr == 0 || l.instr == 901 || l.instr == 902 {
		return l.instr, nil
	}
	// DAT [xxx], defaults to 0
	if l.instr == -1 {
		if l.addr == "" {
			return 0, nil
		}
		return stoi(l.addr)
	}
	// Instructions other than IN/OUT/HLT need a target address
	// so if we are not given one, error out.
	if l.addr == "" {
		return 0, newError(l.lineNo, "no address given")
	}
	if i, ok := labels[l.addr]; ok {
		return l.instr + i, nil
	}
	i, err := stoi(l.addr)
	return l.instr + i, err
}

func linesToInt(lines []*Line) ([]int, error) {
	// Perform 1 pass to first index the positions of the
	// mailboxes in the code so that it is possible to reference
	// a mailbox after/before it is defined
	labels := map[string]int{}
	for mailbox, line := range lines {
		if line.label == "" {
			continue
		}
		labels[line.label] = mailbox
	}
	// Fill up the mailboxes by parsing the instructions
	buff := make([]int, 100)
	for i, line := range lines {
		instr, err := line.toData(labels)
		buff[i] = instr
		if err != nil {
			return nil, err
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

func parse(r io.Reader) ([]*Line, []error) {
	i := 0            // current mailbox number
	lineNo := 0       // current line number
	buff := []*Line{} // compile buffer
	errors := []error{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		lineNo++
		line, err := newLineFromString(s)
		if err == nil && line == nil {
			// Empty line / only comments so don't bother incrementing
			// mailbox number
			continue
		}
		line.lineNo = lineNo
		i++
		if i == 100 { // Reached mailbox limit
			errors = append(errors, newError(lineNo, "Out of mailboxes"))
			break
		}
		if err != nil {
			errors = append(errors, err)
			continue
		}
		buff = append(buff, line)
	}
	return buff, errors
}

func compile(r io.Reader) ([]int, int, []error) {
	lines, errors := parse(r)
	if len(errors) != 0 {
		return nil, 0, errors
	}
	code, err := linesToInt(lines)
	errors = []error{}
	if err != nil {
		errors = append(errors, err)
	}
	return code, len(lines), errors
}
