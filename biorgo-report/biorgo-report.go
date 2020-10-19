package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"
)

// tbn

// sort storg by key in ascending order
func sortStorg(storg []map[string]interface{}, key string) []map[string]interface{} {

	sort.SliceStable(storg, func(i, j int) bool {

		nodeLast := storg[i]
		nodeNext := storg[j]

		metaLast := nodeLast["metadatum"].(map[string]interface{})
		keyLast := metaLast[key].(string)

		metaNext := nodeNext["metadatum"].(map[string]interface{})
		keyNext := metaNext[key].(string)

		return keyLast < keyNext
	})

	return storg
}

// filter storg by key value
func filterStorg(storgOld []map[string]interface{}, key string, value string) []map[string]interface{} {

	storgNew := []map[string]interface{}{}

	for _, node := range storgOld {
		if keyEquals(node, key, value) {
			storgNew = append(storgNew, node)
		}
	}

	return storgNew
}

// filter storg by key in a time period between start and end
func betweenDates(storgOld []map[string]interface{}, key string, start string, end string) []map[string]interface{} {

	storgNew := []map[string]interface{}{}

	for _, node := range storgOld {
		if between(node, key, start, end) {
			storgNew = append(storgNew, node)
		}
	}

	return storgNew
}

// tell if node key value is between start and period
func between(node map[string]interface{}, key string, start string, end string) bool {

	meta := node["metadatum"].(map[string]interface{})
	date := meta[key].(string)

	layout := "<2006-01-02>"
	timeNode, err := time.Parse(layout, date)
	timeStart, err := time.Parse(layout, start)
	timeEnd, err := time.Parse(layout, end)

	if err != nil {
		// fmt.Println("not a date")
		return false
	}

	if timeNode.After(timeStart) && timeNode.Before(timeEnd) {
		return true
	}

	return false
}

// tell if a storg node has value
func keyEquals(node map[string]interface{}, key string, keyword string) bool {

	meta := node["metadatum"].(map[string]interface{})
	value := meta[key].(string)

	return value == keyword
}

// filter storg oldArray by unique key
func uniqueDate(storgOld []map[string]interface{}, key string) []map[string]interface{} {
	storgNew := []map[string]interface{}{}

	for _, node := range storgOld {
		if !storgContains(storgNew[:], node, key) {
			storgNew = append(storgNew, node)
		}
	}

	// r := []string{}
	// for _, s := range e {
	// 	if !contains(r[:], s) {
	// 		r = append(r, s)
	// 	}
	// }
	return storgNew
}

// tell if node index is divisible by 2400
// allows to partition dates in dot notation and avoid graphviz error
func reached2400(index int) bool {
	return index%2400 == 0
}

// tell if a storg array storgContains elements with the same key value as elementNew
func storgContains(array []map[string]interface{}, elementNew map[string]interface{}, key string) bool {

	metaNew := elementNew["metadatum"].(map[string]interface{})
	dateNew := metaNew[key].(string)

	for _, elementOld := range array {

		metaOld := elementOld["metadatum"].(map[string]interface{})
		dateOld := metaOld[key].(string)

		if dateOld == dateNew {
			return true
		}
	}
	// var dat map[string]interface{}
	// strs := dat["strs"].([]interface{})
	// str1 := strs[0].(string)
	// fmt.Println(str1)

	// for _, s := range e {
	// 	if s == c {
	// 		return true
	// 	}
	// }
	return false
}

// tbn

// parse storg json
func parseStorg(storgPath string) []map[string]interface{} {

	// read a storg file
	storgFile, err := ioutil.ReadFile(storgPath)
	if err != nil {
		fmt.Println("File reading error", err)
		return nil
	}

	// fmt.Println("Contents of file:", len(string(storgFile)))

	var storgMap []map[string]interface{}

	if err := json.Unmarshal(storgFile, &storgMap); err != nil {
		panic(err)
	}
	fmt.Println("Number of nodes:", len(storgMap))

	return storgMap
}

// generate Biorg desmi from storg
func generateDesmi(storg []map[string]interface{}, templatePath string, outputPath string) {

	// read a go template
	templateString, err := ioutil.ReadFile(templatePath)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	// 	var storgFileTemplateStr = `{{ range . }}
	// * .
	// :PROPERTIES: {{ with .metadatum }}
	// :COMMENT: {{ with .COMMENT }}{{ . }}{{ end }}
	// :GUEST_DATE: {{ with .GUEST_DATE }}{{ . }}{{ end }}
	// :GUEST: {{ with .GUEST }}{{ . }}{{ end }}
	// :HOST_DATE: {{ with .HOST_DATE }} {{ . }}{{ end }}
	// :HOST: {{ with .HOST }}{{ . }}{{ end }}
	// :LABEL: {{ with .LABEL }}{{ . }}{{ end }}
	// :MODULE: {{ with .MODULE }}{{ . }}{{ end }}
	// :TYPE: {{ with .TYPE }}{{ . }}{{ end }}{{ end }}
	// :UUID: {{ with .datum.uuid }}{{ . }}{{ end }}
	// :END:
	// {{ with .datum.entry }}{{ . }}{{ end }}{{ end }}`

	customFunctions := template.FuncMap{"sortStorg": sortStorg, "filterStorg": filterStorg, "betweenDates": betweenDates}

	templateStruct, err := template.New("nodesDesmi").Funcs(customFunctions).Parse(string(templateString))
	if err != nil {
		panic(err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	//	var tpl bytes.Buffer
	err = templateStruct.Execute(file, storg)
	if err != nil {
		panic(err)
	}
}

// generate dot notation
func generateDot(storg []map[string]interface{}, templatePath string, outputPath string) {

	// format storg entries to prevent graphviz errors
	for _, node := range storg {
		datum := node["datum"].(map[string]interface{})
		// remove symbols instead of escaping because backslashes might otherwise escape closing quotes
		// DO NOT REUSE FOR RAVDIA, BREAKS VALIDITY
		// remote newlines
		entry := datum["entry"].(string)
		datum["entry"] = strings.Replace(entry, "\n", "", -1)
		// remove quotes
		entry = datum["entry"].(string)
		datum["entry"] = strings.Replace(entry, "\"", "", -1)
		// remove the line tabulation character
		entry = datum["entry"].(string)
		datum["entry"] = strings.Replace(entry, "", "", -1)
	}

	// read a template
	templateString, err := ioutil.ReadFile(templatePath)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	customFunctions := template.FuncMap{"uniqueDate": uniqueDate, "reached2400": reached2400, "sortStorg": sortStorg, "filterStorg": filterStorg, "betweenDates": betweenDates}

	templateStruct, err := template.New("nodesDot").Funcs(customFunctions).Parse(string(templateString))
	if err != nil {
		panic(err)
	}

	file, err := os.Create(outputPath)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	//	var tpl bytes.Buffer
	err = templateStruct.Execute(file, storg)
	if err != nil {
		panic(err)
	}

}

// generate ravdia org files
func generateRavdia(storg []map[string]interface{}, templatePath string, outputPath string) {

	// read a template
	templateString, err := ioutil.ReadFile(templatePath)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	customFunctions := template.FuncMap{"uniqueDate": uniqueDate, "reached2400": reached2400, "sortStorg": sortStorg, "filterStorg": filterStorg, "betweenDates": betweenDates}

	templateStruct, err := template.New("nodesRavdia").Funcs(customFunctions).Parse(string(templateString))
	if err != nil {
		panic(err)
	}

	for _, node := range storg {

		data := node["datum"].(map[string]interface{})
		uuid := data["uuid"].(string)

		file, err := os.Create(outputPath + "/" + uuid + ".org")
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}

		err = templateStruct.Execute(file, node)
		if err != nil {
			panic(err)
		}

		file.Close()

	}

}

func main() {

	var reportType string
	var inputPath string
	var templatePath string
	var outputPath string

	// flags declaration using flag package
	flag.StringVar(&reportType, "r", "empty", "Please specify report: desmi, dot, ravdia")
	flag.StringVar(&inputPath, "i", "empty", "Please specify storg path")
	flag.StringVar(&templatePath, "t", "empty", "Please specify template path")
	flag.StringVar(&outputPath, "o", "empty", "Please specify output path")

	flag.Parse() // after declaring flags we need to call it
	// TODO: show usage when no arguments are specified

	if reportType == "desmi" {
		var storg = parseStorg(inputPath)
		generateDesmi(storg, templatePath, outputPath)
	} else if reportType == "dot" {
		var storg = parseStorg(inputPath)
		generateDot(storg, templatePath, outputPath)
	} else if reportType == "ravdia" {
		var storg = parseStorg(inputPath)
		generateRavdia(storg, templatePath, outputPath)
	}
}
