package main

import "io"
import "fmt"

type entry struct {
	mailbox int
	text    string
	count   int
}

func maxCount(entries []entry) int {
	max := 1
	for _, entry := range entries {
		if entry.count > max {
			max = entry.count
		}
	}
	return max
}

type heatmapVM struct {
	vm      *context
	code    []int
	lines   []*Line
	heatmap map[int]int
}

func newHeatmapVM(r io.Reader) (*heatmapVM, []error) {
	lines, errors := parse(r)
	if len(errors) != 0 {
		return nil, errors
	}
	code, err := linesToInt(lines)
	if err != nil {
		return nil, []error{err}
	}
	return &heatmapVM{
		vm:      newContextFromSlice(code),
		code:    code,
		lines:   lines,
		heatmap: map[int]int{},
	}, nil
}

func (h *heatmapVM) run(input []int) (output []int, err error) {
	vm := h.vm
	heatmap := h.heatmap
	vm.input = input
	for !vm.halted {
		heatmap[vm.pc]++
		err = vm.fetchExecute()
		if err != nil {
			break
		}
	}
	output = vm.output
	return
}

func (h *heatmapVM) format() []entry {
	entries := make([]entry, 100)
	for i, _ := range entries {
		count, ok := h.heatmap[i]
		text := ""
		// first check if the mailbox is a line of code
		if len(h.lines) > i {
			text = h.lines[i].text
		} else if ok {
			// else check that we have executed this mailbox
			text = fmt.Sprintf("%03d", h.vm.mem[i])
		}
		entries[i] = entry{i, text, count}
	}
	return entries
}

func writeEntries(entries []entry, w io.Writer) error {
	_, err := w.Write([]byte(`
	<style>
	table  { border-collapse: collapse; }
	td     { padding: 0 0.5em; border: 1px solid #000; text-align: left; }
	pre,td { font-family: 'Inconsolata', monospace; }
	.count { font-weight: bold; text-align: right; }
	</style>
	<table>
	`))
	if err != nil {
		return err
	}
	max := maxCount(entries)
	for _, e := range entries {
		r := int(255 * (2 * e.count / max))
		g := int(255 * (2 * (1 - e.count/max)))
		s := fmt.Sprintf(
			"<tr><td style='background-color: rgba(%d, %d, 0, 0.35)' class='count'>%d</td><td>%02d</td><td><pre>%s</pre></td></tr>",
			r, g,
			e.count,
			e.mailbox,
			e.text,
		)
		_, err := w.Write([]byte(s))
		if err != nil {
			return err
		}
	}
	return nil
}
