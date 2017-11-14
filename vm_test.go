package main

import "strings"
import "testing"
import "github.com/stretchr/testify/assert"

func TestVMIO(t *testing.T) {
	r := strings.NewReader(`
	IN
	STO	small
	IN		
	STO	big
	SUB	small # Next check that we received the smaller first
	BRP	ok	# no need to exchange
	LDA	small	# exchange
	STO	temp
	LDA	big
	STO	small
	LDA	temp
	STO	big 
ok	LDA	small	# add 1 to small
	ADD	one	# constant
	STO	small
while	SUB	big
	BRP	endw	# not smaller any more
	LDA	small	# display small
	OUT
	ADD	one	# add 1 to small
	STO	small
	BR	while	# loop back
endw	HLT		# stop
one	DAT	001	# defining a constant
small	DAT		# space for data value
big	DAT		# space for data value
temp	DAT	# space for data value
	`)
	code, _, errors := compile(r)
	assert.Equal(t, len(errors), 0, "No errors in compilation")
	vm := newContextFromSlice(code)
	vm.input = []int{10, 20}
	output, err := vm.run()
	assert.Equal(t, err, nil)
	assert.Equal(t, output, []int{11, 12, 13, 14, 15, 16, 17, 18, 19})
	// check that once we reset, all inputs are reset and the
	// noMoreInputs error is returned
	vm.reset()
	_, err = vm.run()
	assert.Equal(t, err, noMoreInputs)
}
