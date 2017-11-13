package main

import "strings"
import "testing"
import "github.com/stretchr/testify/assert"

type lineFromStringTest struct {
	text   string // text
	lineNo int    // lineNo
	label  string // label
	instr  string // instruction
	addr   string // address
	line   bool   // line != nil
	err    bool   // err != nil
}

func TestStringToLine(t *testing.T) {
	tests := []lineFromStringTest{
		lineFromStringTest{"\tLDA\tabc", 10, "", "LDA", "abc", true, false},
		lineFromStringTest{" LDA\t123", 10, "", "LDA", "123", true, false},
		lineFromStringTest{"abc\tLDA\tghi", 10, "abc", "LDA", "ghi", true, false},
		lineFromStringTest{"\tLDA\tabc #def", 10, "", "LDA", "abc", true, false},
		lineFromStringTest{"abc", 10, "", "", "", true, true},
		lineFromStringTest{"", 10, "", "", "", false, false},
		lineFromStringTest{"abc LDA\tabc\tdef", 10, "", "", "", false, true},
	}
	for _, c := range tests {
		line, err := newLineFromString(c.lineNo, c.text)
		assert.Equal(t, err != nil, c.err, c.text)
		if c.line && !c.err {
			assert.Equal(t, line.lineNo, c.lineNo)
			assert.Equal(t, line.label, c.label)
			assert.Equal(t, line.instr, c.instr)
			assert.Equal(t, line.addr, c.addr)
		}
	}
}

type lineToDataTest struct {
	line   Line
	labels map[string]int
	err    bool
	data   int
}

func TestLineToData(t *testing.T) {
	tests := []lineToDataTest{
		lineToDataTest{
			line:   Line{lineNo: 5, label: "", instr: "LDA", addr: "057"},
			labels: map[string]int{},
			err:    false,
			data:   557,
		},
		lineToDataTest{
			line:   Line{lineNo: 5, label: "", instr: "LDA", addr: "abc"},
			labels: map[string]int{"abc": 12},
			err:    false,
			data:   512,
		},
		lineToDataTest{
			line:   Line{lineNo: 5, label: "", instr: "DAT", addr: "009"},
			labels: map[string]int{},
			err:    false,
			data:   9,
		},
		lineToDataTest{
			line:   Line{lineNo: 5, label: "", instr: "DAT", addr: "label"},
			labels: map[string]int{"label": 1},
			err:    true,
			data:   0,
		},
		lineToDataTest{
			line:   Line{lineNo: 5, label: "abc", instr: "IN", addr: ""},
			labels: map[string]int{},
			err:    false,
			data:   901,
		},
		lineToDataTest{
			line:   Line{lineNo: 5, label: "abc", instr: "DAT", addr: "1000"},
			labels: map[string]int{},
			err:    true,
			data:   901,
		},
		lineToDataTest{
			line:   Line{lineNo: 5, label: "abc", instr: "STO", addr: ""},
			labels: map[string]int{},
			err:    true,
			data:   0,
		},
		lineToDataTest{
			line:   Line{lineNo: 5, label: "abc", instr: "DAT", addr: ""},
			labels: map[string]int{},
			err:    false,
			data:   0,
		},
		lineToDataTest{
			line:   Line{lineNo: 5, label: "", instr: "STO", addr: "def"},
			labels: map[string]int{"abc": 1},
			err:    true,
			data:   0,
		},
		lineToDataTest{
			line:   Line{lineNo: 5, label: "", instr: "FOO", addr: "def"},
			labels: map[string]int{"abc": 1},
			err:    true,
			data:   0,
		},
	}
	for _, c := range tests {
		data, err := c.line.toData(c.labels)
		assert.Equal(t, err != nil, c.err)
		if !c.err {
			assert.Equal(t, data, c.data)
		}
	}
}

func TestCompile(t *testing.T) {
	r := strings.NewReader(`
# test program
st	IN
	STO inp
	BRZ lab
	BRP st
lab	HLT
inp	DAT 100
	`)
	code, mailboxes, errors := compile(r)
	buff := make([]int, 100)
	buff[0] = 901
	buff[1] = 305
	buff[2] = 704
	buff[3] = 800
	buff[4] = 0
	buff[5] = 100
	assert.Equal(t, len(errors), 0)
	assert.Equal(t, code, buff)
	assert.Equal(t, mailboxes, 6)
}
