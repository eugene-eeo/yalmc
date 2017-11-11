package main

import "strconv"
import "io"
import "fmt"

const heatmapHTMLFrontMatter = `
<style>
table  { border-collapse: collapse; }
td     { padding: 0 0.5em; border: 1px solid #000; text-align: left; font-family: monospace; }
.count { font-weight: bold; }
</style>
<table>
`

type entry struct {
	text  string
	count int
}

func (e *entry) toHTML(max int) string {
	r := int(255 * (2 * e.count / max))
	g := int(255 * (2 * (1 - e.count/max)))
	color := fmt.Sprintf("rgba(%d, %d, 0, 0.35)", r, g)
	return fmt.Sprintf(
		"<tr><td style='background-color: %s' class='count'>%d</td><td><pre>%s</pre></td></tr>",
		color,
		e.count,
		e.text,
	)
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

func writeEntries(entries []entry, w io.Writer) error {
	_, err := w.Write([]byte(heatmapHTMLFrontMatter))
	if err != nil {
		return err
	}
	max := maxCount(entries)
	for _, entry := range entries {
		_, err := w.Write([]byte(entry.toHTML(max)))
		if err != nil {
			return err
		}
	}
	return nil
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
	code, errors := linesToInt(lines)
	if len(errors) != 0 {
		return nil, errors
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
	for i, line := range h.lines {
		entries[i] = entry{line.text, 0}
	}
	for i, count := range h.heatmap {
		text := strconv.Itoa(h.vm.mem[i])
		if len(h.lines) > i {
			text = h.lines[i].text
		}
		entries[i] = entry{text, count}
	}
	return entries
}
