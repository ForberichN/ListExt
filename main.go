package main

import (
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type entry struct {
	occurecnes int
	totalSize  int64
}

type data struct {
	Key   string
	Value int
	size  int64
}

func main() {

	const orderOptions = "count|average|total"
	const outputOptions = "table|csv|extension"

	orderBy := flag.String("orderby", "count", "The Field which is used to order the Output After {"+orderOptions+"}")
	output := flag.String("output", "table", "Controls how the output looks {"+outputOptions+"}")
	flag.Parse()

	if !strings.Contains(orderOptions, *orderBy) {
		flag.PrintDefaults()
		return
	}

	if !strings.Contains(outputOptions, *output) {
		flag.PrintDefaults()
		return
	}

	regex, _ := regexp.Compile(".*\\.(.*)")

	root := "."

	if flag.Arg(0) != "" {
		root = flag.Arg(0)
	}

	m := make(map[string]entry)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if info.IsDir() {
			return err
		}
		results := regex.FindStringSubmatch(info.Name())
		switch len(results) {
		case 0:
			e := m[""]
			e.occurecnes += 1
			e.totalSize += info.Size()
			m[""] = e
		case 2:
			e := m[results[1]]
			e.occurecnes += 1
			e.totalSize += info.Size()
			m[results[1]] = e
		}
		return nil
	})

	if err != nil {
		fmt.Println("Failed to walk Directory", err)
		return
	}

	//Sort Extensions by Occurrence

	var extensionsSorted []data
	for k, v := range m {
		extensionsSorted = append(extensionsSorted, data{k, v.occurecnes, v.totalSize})
	}

	sort.Slice(extensionsSorted, func(i, j int) bool {
		switch *orderBy {
		case "count":
			return extensionsSorted[i].Value > extensionsSorted[j].Value
		case "average":
			return (extensionsSorted[i].size / int64(extensionsSorted[i].Value)) > (extensionsSorted[j].size / int64(extensionsSorted[j].Value))
		case "total":
			return extensionsSorted[i].size > extensionsSorted[j].size
		}
		return extensionsSorted[i].Value > extensionsSorted[j].Value
	})

	switch *output {
	case "table":
		printTable(extensionsSorted)
	case "csv":
		printCSV(extensionsSorted)
	case "extension":
		printExtension(extensionsSorted)
	}
}

func printExtension(extensionsSorted []data) {
	for _, kv := range extensionsSorted {
		fileName := "No Filename"
		if kv.Key != "" {
			fileName = kv.Key
		}
		fmt.Println(fileName)
	}
}

func printCSV(extensionsSorted []data) {
	for _, kv := range extensionsSorted {
		fileName := "No Filename"
		if kv.Key != "" {
			fileName = kv.Key
		}
		fmt.Printf("%s,%v,%v,%v\n", fileName, kv.Value, uint64(kv.size/int64(kv.Value)), uint64(kv.size))
	}
}

func printTable(extensionsSorted []data) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	headFoot := table.Row{"File Extension", "Count", "Average Size", "Total Size"}
	t.AppendHeader(headFoot)

	for _, kv := range extensionsSorted {
		fileName := "No Filename"
		if kv.Key != "" {
			fileName = kv.Key
		}
		t.AppendRow(table.Row{fileName, kv.Value, humanize.Bytes(uint64(kv.size / int64(kv.Value))), humanize.Bytes(uint64(kv.size))})
	}
	t.AppendFooter(headFoot)
	t.Render()
}
