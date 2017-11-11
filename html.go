package main

import "strings"
import "strconv"
import "path/filepath"
import "fmt"
import "io"

const tableFrontmatter string = `
<style>
table { border-collapse: collapse; }
td,th { padding: 0 0.5em; border: 1px solid #000; }
</style>
<table>
<tr>
	<th>Filename</th>
	<th>Mailboxes</th>
	<th>Test Case</th>
	<th>Inputs</th>
	<th>Expected Output</th>
	<th>Output</th>
	<th>Max Cycles</th>
	<th>Cycles</th>
</tr>
`

func isliceToString(a []int) string {
	b := make([]string, len(a))
	for i, n := range a {
		b[i] = strconv.Itoa(n)
	}
	return strings.Join(b, ", ")
}

func isliceEq(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, x := range a {
		if b[i] != x {
			return false
		}
	}
	return true
}

type table struct {
	fragments []string
}

func newTable() *table {
	return &table{[]string{}}
}

func (t *table) addRow(path string, mailboxes int, results []testResult) {
	abs, _ := filepath.Abs(path)
	trs := []string{fmt.Sprintf(
		"<tr><th rowspan='%d'><a href='%s'>%s</a></th><td rowspan='%d'>%d</td></tr>",
		len(results)+1,
		abs,
		filepath.Base(path),
		len(results)+1,
		mailboxes,
	)}
	for _, res := range results {
		color := "#ffffff"
		if res.terminated || !isliceEq(res.output, res.tcase.output) {
			color = "#ff6666"
		}
		trs = append(trs, fmt.Sprintf(
			"<tr style='background-color:%s'><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%d</td></tr>",
			color,
			res.tcase.name,
			isliceToString(res.tcase.input),
			isliceToString(res.tcase.output),
			isliceToString(res.output),
			res.tcase.cycleLimit,
			res.cycles,
		))
	}
	t.fragments = append(t.fragments, strings.Join(trs, ""))
}

func (t *table) write(w io.Writer) error {
	_, err := w.Write([]byte(tableFrontmatter))
	if err != nil {
		return err
	}
	for _, tr := range t.fragments {
		_, err = w.Write([]byte(tr))
		if err != nil {
			return err
		}
	}
	return nil
}
