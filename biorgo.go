package main

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
)

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

func uniqueDate(oldArray []map[string]interface{}, key string) []map[string]interface{} {
	newArray := []map[string]interface{}{}

	for _, node := range oldArray {
		if !contains(newArray[:], node, key) {
			newArray = append(newArray, node)
		}
	}

	// r := []string{}
	// for _, s := range e {
	// 	if !contains(r[:], s) {
	// 		r = append(r, s)
	// 	}
	// }
	return newArray
}

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

// func contains(e []string, c string) bool {
// 	for _, s := range e {
// 		if s == c {
// 			return true
// 		}
// 	}
// 	return false
// }

func reached2400(index int) bool {
	return index%2400 == 0
}

func main() {

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

	/****************/
	/*generate desmi*/
	/****************/

	/* read a storg file*/
	storgFile, err := ioutil.ReadFile("../test/storg7.json")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	fmt.Println("Contents of file:", len(string(storgFile)))

	var storgFileMap []map[string]interface{}

	if err := json.Unmarshal(storgFile, &storgFileMap); err != nil {
		panic(err)
	}
	fmt.Println(len(storgFileMap))

	/* format entries */
	for _, node := range storgFileMap {
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

	/* read a go template*/
	templateDesmiStr, err := ioutil.ReadFile("desmi.txt")
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

	templateDesmiOut, err := template.New("nodesDesmi").Parse(string(templateDesmiStr))
	if err != nil {
		panic(err)
	}

	fileDesmi, err := os.Create("test.org")

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	//	var tpl bytes.Buffer
	err = templateDesmiOut.Execute(fileDesmi, storgFileMap)
	if err != nil {
		panic(err)
	}

	/**************/
	/*generate svg*/
	/**************/

	templateDotStr, err := ioutil.ReadFile("dot.txt")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	customFunctionsDot := template.FuncMap{"uniqueDate": uniqueDate, "reached2400": reached2400}

	templateDotOut, err := template.New("nodesDot").Funcs(customFunctionsDot).Parse(string(templateDotStr))
	if err != nil {
		panic(err)
	}

	fileDot, err := os.Create("test.dot")

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	//	var tpl bytes.Buffer
	err = templateDotOut.Execute(fileDot, storgFileMap)
	if err != nil {
		panic(err)
	}

}
