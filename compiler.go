package main

import "io"
import "bufio"
import "fmt"
import "strings"
import "strconv"

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

func stoi(s string, max int) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if i > max || i < 0 {
		return 0, fmt.Errorf("number not in range 0-%d", max)
	}
	return i, nil
}

type Line struct {
	text   string
	lineNo int
	label  string
	instr  string
	addr   string
}

func newLineFromString(lineNo int, s string) (*Line, error) {
	s = strings.SplitN(s, "#", 2)[0]
	if len(strings.TrimSpace(s)) == 0 {
		return nil, nil
	}
	parts := strings.Fields(s)
	if len(parts) > 3 {
		return nil, newError(lineNo, "unexpected content after address section")
	}
	if s[0] == ' ' || s[0] == '\t' {
		c := []string{""}
		parts = append(c, parts...)
	}
	if len(parts) < 2 {
		return nil, newError(lineNo, "got label with no instruction")
	}
	label := parts[0]
	instr := parts[1]
	addr := ""
	if len(parts) == 3 {
		addr = parts[2]
	}
	return &Line{
		text:   s,
		lineNo: lineNo,
		label:  label,
		instr:  strings.ToUpper(instr),
		addr:   addr,
	}, nil
}

func (l *Line) toData(labels map[string]int) (int, error) {
	op, ok := instrLookup[l.instr]
	if !ok {
		return 0, newError(l.lineNo, fmt.Sprintf("invalid instruction '%s'", l.instr))
	}
	// HLT / IN / OUT instructions can be on their own without
	// any address component
	if op == 0 || op == 901 || op == 902 {
		return op, nil
	}
	// DAT [xxx], defaults to 0
	if op == -1 {
		if l.addr == "" {
			return 0, nil
		}
		return stoi(l.addr, 999)
	}
	// Instructions other than IN/OUT/HLT need a target address
	// so if we are not given one, error out.
	if l.addr == "" {
		return 0, newError(l.lineNo, "no address given")
	}
	if i, ok := labels[l.addr]; ok {
		return op + i, nil
	}
	i, err := stoi(l.addr, 99) // addresses are bounded from 0-99
	if err != nil {
		err = newError(l.lineNo, fmt.Sprintf("invalid address/label: %s", l.addr))
	}
	return op + i, err
}

func linesToInt(lines []*Line) ([]int, error) {
	// Perform 1 pass to first index the positions of the
	// mailboxes in the code so that it is possible to reference
	// a label after/before it is defined
	labels := map[string]int{}
	for mailbox, line := range lines {
		if len(line.label) > 0 {
			labels[line.label] = mailbox
		}
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

func parse(r io.Reader) ([]*Line, []error) {
	i := 0            // current mailbox number
	lineNo := 0       // current line number
	buff := []*Line{} // compile buffer
	errors := []error{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		lineNo++
		line, err := newLineFromString(lineNo, s)
		if err == nil && line == nil {
			// Empty line / only comments so don't bother incrementing
			// mailbox number
			continue
		}
		i++
		if i == 100 { // Reached mailbox limit
			errors = append(errors, newError(lineNo, "out of mailboxes"))
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
		errors = []error{err}
	}
	return code, len(lines), errors
}
