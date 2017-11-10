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

func (c *context) fetchExecute() (err error) {
	instruction := c.mem[c.pc]
	c.pc = (c.pc + 1) % 1000
	opcode := instruction / 100
	addr := instruction % 100
	switch opcode {
	case 0: // HLT
		c.halted = true
		return
	case 1: // ADD
		c.acc += c.mem[addr]
		c.acc %= 1000
		return
	case 2: // SUB
		c.acc -= c.mem[addr]
		c.acc %= 1000
		return
	case 3: // STO
		c.mem[addr] = c.acc
		return
	case 5: // LDA
		c.acc = c.mem[addr]
		return
	case 6: // BR
		c.pc = addr
		return
	case 7: // BRZ
		if c.acc == 0 {
			c.pc = addr
		}
		return
	case 8:
		if c.acc >= 0 {
			c.pc = addr
		}
		return
	case 9:
		// 901 => IN
		if addr == 1 {
			if len(c.input) == 0 {
				c.halted = true
				err = noMoreInputs
				return
			}
			c.acc = c.input[0]
			c.input = c.input[1:]
			return
		}
		// 902 => OUT
		if addr == 2 {
			c.output = append(c.output, c.acc)
			return
		}
	}
	return
}

func (c *context) run() (output []int, err error) {
	for !c.halted {
		err = c.fetchExecute()
		if err != nil {
			return
		}
	}
	output = c.output
	return
}
