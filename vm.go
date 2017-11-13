package main

import "errors"

var noMoreInputs error = errors.New("no input given")

type context struct {
	mem    [100]int
	acc    int
	pc     int
	neg    bool
	input  []int
	output []int
	halted bool
}

func newContextFromSlice(mailboxes []int) *context {
	ctx := context{}
	// Size of mailboxes bounded from 0-100 since
	// we're accepting input from the `compile`
	// function.
	for i, m := range mailboxes {
		ctx.mem[i] = m
	}
	return &ctx
}

func (c *context) reset() {
	c.input = []int{}
	c.output = []int{}
	c.halted = false
	c.pc = 0
}

func (c *context) fetchExecute() (err error) {
	instruction := c.mem[c.pc]
	c.pc = (c.pc + 1) % 1000
	opcode := instruction / 100
	addr := instruction % 100
	switch opcode {
	case 0: // HLT
		c.halted = true
	case 1: // ADD
		c.neg = false
		c.acc += c.mem[addr]
		c.acc %= 1000
	case 2: // SUB
		c.acc -= c.mem[addr]
		if c.acc < 0 {
			c.neg = true
		}
		c.acc %= 1000
	case 3: // STO
		c.mem[addr] = c.acc
	case 5: // LDA
		c.neg = false
		c.acc = c.mem[addr]
	case 6: // BR
		c.pc = addr
	case 7: // BRZ
		if c.acc == 0 {
			c.pc = addr
		}
	case 8: // BRP
		if !c.neg {
			c.pc = addr
		}
	case 9:
		// 901 => IN
		if addr == 1 {
			if len(c.input) == 0 {
				c.halted = true
				err = noMoreInputs
				return
			}
			c.acc = c.input[0]
			c.neg = false
			c.input = c.input[1:]
		}
		// 902 => OUT
		if addr == 2 {
			c.output = append(c.output, c.acc)
		}
	}
	return
}

func (c *context) run() (output []int, err error) {
	for !c.halted {
		err = c.fetchExecute()
		if err != nil {
			break
		}
	}
	output = c.output
	return
}
