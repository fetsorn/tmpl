// Copyright © 2020 Anton Davydov <fetsorn@gmail.com>.

package main

// Biorgo generates reports from a JSON array called storg that holds Biorg entries
import (
	// "bytes"
	"encoding/json"
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

// tell if date1 is earlier than date2
func earlierDate(date1 string, date2 string) bool {

	layout := "<2006-01-02>"
	time1, err := time.Parse(layout, date1)
	time2, err := time.Parse(layout, date2)

	if err != nil {
		fmt.Println("not a date")
		return false
	}

	if time1.Before(time2) {
		return true
	}
	return false
}

// parse storg json
func parseStorg() []map[string]interface{} {

	// read a storg file
	storgFile, err := ioutil.ReadFile("../test/storg7.json")
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
		if equals(node, key, value) {
			storgNew = append(storgNew, node)
		}
	}

	return storgNew
}

// tell if a storg node has value
func equals(node map[string]interface{}, key string, keyword string) bool {

	meta := node["metadatum"].(map[string]interface{})
	value := meta[key].(string)

	return value == keyword
}

// filter storg oldArray by unique key
func uniqueDate(storgOld []map[string]interface{}, key string) []map[string]interface{} {
	storgNew := []map[string]interface{}{}

	for _, node := range storgOld {
		if !contains(storgNew[:], node, key) {
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

// tell if a storg array contains elements with the same key value as elementNew
func contains(array []map[string]interface{}, elementNew map[string]interface{}, key string) bool {

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

// tell if node index is divisible by 2400
// allows to partition dates in dot notation and avoid graphviz error
func reached2400(index int) bool {
	return index%2400 == 0
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

// tbn

// feeds json to a template, prints a biorg entry
func templateTest() {

	/***********************/
	/*feed json to template*/
	/***********************/

	storgJSON := []byte(`{
  "metadatum": {
    "COMMENT": "",
    "GUEST_DATE": "<2013-01-26>",
    "GUEST": "fetsorn",
    "HOST_DATE": "<2013-01-26>",
    "HOST": "fetsorn",
    "LABEL": "",
    "MODULE": "winter2019-transcript",
    "TYPE": ""
  },
  "datum": {
    "uuid": "5339DE8B-2E14-47C3-8F2F-4954B9A09E6E",
    "entry": "26.01.2013 аудио: \"My recording #12\", вчера был четверг, полный <> день, он мои поверг мозги набекрень, инфаркт миокарда перенес в аптеке и в картину эдварда о смеющемся человеке я вперил взгляд с ожирением вряд-ли он как то может двигаться сам, особенно по четвергам, бада, бурду эту я сделал сам.\n"
  }
}`)

	var storgMap map[string]interface{}

	if err := json.Unmarshal(storgJSON, &storgMap); err != nil {
		panic(err)
	}
	fmt.Println(storgMap)

	fmt.Println(" ")

	// var testTemplate *template.Template
	// var err error
	// testTemplate, err = template.New("hello.gohtml").Funcs(template.FuncMap{
	// 	"earlierDate": func(date1 string, date2 string) bool {

	// 		layout := "<2006-01-02>"
	// 		time1, err := time.Parse(layout, date1)
	// 		time2, err := time.Parse(layout, date2)

	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}

	// 		if time1.Before(time2) {
	// 			return true
	// 		}
	// 		return false
	// 	},
	// }).ParseFiles("hello.gohtml")
	// if err != nil {
	// 	panic(err)
	// }

	// date2 := "<2010-01-01>"
	// date1 := "<2005-01-01>"
	// layout := "<2006-01-02>"
	// time1, err := time.Parse(layout, date1)
	// time2, err := time.Parse(layout, date2)

	// if time1.Before(time2) {
	// 	fmt.Println("yes")
	// } else {
	// 	fmt.Println("no")
	// }

	var templateTestStr = `{{ $variable := .datum.entry }}
	* .
	:PROPERTIES:
	:HOST: {{ .metadatum.HOST }}
	:HOST_DATE: {{ if .metadatum.HOST_DATE }} {{ .metadatum.HOST_DATE }} {{ end }}
    :GUEST_DATE: {{ if earlierDate .metadatum.HOST_DATE "<2021-01-01>" }} {{ .metadatum.HOST_DATE }} {{ else }} {{ .metadatum.HOST }} {{ end }}
    {{ $variable }}
	`
	customFunctionsTest := template.FuncMap{"earlierDate": earlierDate}

	templateTestOut, err := template.New("nodes").Funcs(customFunctionsTest).Parse(templateTestStr)
	if err != nil {
		panic(err)
	}

	err = templateTestOut.Execute(os.Stdout, storgMap)
	if err != nil {
		panic(err)
	}
}

// generates Biorg desmi from storg
func generateDesmi(storg []map[string]interface{}) {

	/* read a go template*/
	templateDesmiStr, err := ioutil.ReadFile("desmi.tmpl")

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

	customFunctionsDesmi := template.FuncMap{"sortStorg": sortStorg, "filterStorg": filterStorg, "betweenDates": betweenDates}

	templateDesmiOut, err := template.New("nodesDesmi").Funcs(customFunctionsDesmi).Parse(string(templateDesmiStr))
	if err != nil {
		panic(err)
	}

	fileDesmi, err := os.Create("test.org")

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	//	var tpl bytes.Buffer
	err = templateDesmiOut.Execute(fileDesmi, storg)
	if err != nil {
		panic(err)
	}
}

// generates dot notation
func generateDot(storg []map[string]interface{}) {

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
	templateDotStr, err := ioutil.ReadFile("dot.tmpl")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	customFunctionsDot := template.FuncMap{"uniqueDate": uniqueDate, "reached2400": reached2400, "sortStorg": sortStorg, "filterStorg": filterStorg, "betweenDates": betweenDates}

	templateDotOut, err := template.New("nodesDot").Funcs(customFunctionsDot).Parse(string(templateDotStr))
	if err != nil {
		panic(err)
	}

	fileDot, err := os.Create("test.dot")

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	//	var tpl bytes.Buffer
	err = templateDotOut.Execute(fileDot, storg)
	if err != nil {
		panic(err)
	}

}

func main() {

	// templateTest()

	var storg = parseStorg()

	generateDesmi(storg)

	generateDot(storg)

}
