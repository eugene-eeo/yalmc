package main

import "testing"
import "github.com/stretchr/testify/assert"

func TestStringToLine(t *testing.T) {
	line, err := newLineFromString("\tLDA\taddr # comment")
	assert.Equal(t, err, nil)
	assert.Equal(t, line.label, "")
	assert.Equal(t, line.instr, "LDA")
	assert.Equal(t, line.addr, "addr")
	assert.Equal(t, line.text, "\tLDA\taddr ")
}

func TestStringToLineWithLabel(t *testing.T) {
	line, err := newLineFromString("label\tLDA\taddr # comment")
	assert.Equal(t, err, nil)
	assert.Equal(t, line.label, "label")
	assert.Equal(t, line.instr, "LDA")
	assert.Equal(t, line.addr, "addr")
	assert.Equal(t, line.text, "label\tLDA\taddr ")
}

func TestStringToLineWithSpaces(t *testing.T) {
	line, err := newLineFromString(" LDA\taddr # comment")
	assert.Equal(t, err, nil)
	assert.Equal(t, line.label, "")
	assert.Equal(t, line.instr, "LDA")
	assert.Equal(t, line.addr, "addr")
	assert.Equal(t, line.text, " LDA\taddr ")
}

func TestEmptyStringToLine(t *testing.T) {
	line, err := newLineFromString("# comment")
	assert.Equal(t, err, nil)
	if line != nil || err != nil {
		t.Fail()
	}
}
