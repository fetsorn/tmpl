package main

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"
	"time"
)

func main() {

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

	var storgTemplateStr = `{{ $variable := .datum.entry }}
	* .
	:PROPERTIES:
	:HOST: {{ .metadatum.HOST }}
	:HOST_DATE: {{ if .metadatum.HOST_DATE }} {{ .metadatum.HOST_DATE }} {{ end }}
    :GUEST_DATE: {{ if earlierDate .metadatum.HOST_DATE "test" }} {{ .metadatum.HOST_DATE }} {{ else }} {{ .metadatum.HOST }} {{ end }}
    {{ $variable }}
	`

	storgTemplateOut, err := template.New("nodes").Funcs(template.FuncMap{
		"earlierDate": func(date1 string, date2 string) bool {

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
		},
	}).Parse(storgTemplateStr)
	if err != nil {
		panic(err)
	}

	err = storgTemplateOut.Execute(os.Stdout, storgMap)
	if err != nil {
		panic(err)
	}

	/* read a storg file*/
	storgFile, err := ioutil.ReadFile("../test/storg7.json")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	fmt.Println("Contents of file:", len(string(storgFile)))

	var storgFileMap []interface{}

	if err := json.Unmarshal(storgFile, &storgFileMap); err != nil {
		panic(err)
	}
	fmt.Println(len(storgFileMap))

	var storgFileTemplateStr = `{{ range . }}
* .
:PROPERTIES: {{ with .metadatum }}
:COMMENT: {{ with .COMMENT }}{{ . }}{{ end }}
:GUEST_DATE: {{ with .GUEST_DATE }}{{ . }}{{ end }}
:GUEST: {{ with .GUEST }}{{ . }}{{ end }}
:HOST_DATE: {{ with .HOST_DATE }} {{ . }}{{ end }}
:HOST: {{ with .HOST }}{{ . }}{{ end }}
:LABEL: {{ with .LABEL }}{{ . }}{{ end }}
:MODULE: {{ with .MODULE }}{{ . }}{{ end }}
:TYPE: {{ with .TYPE }}{{ . }}{{ end }}{{ end }}
:UUID: {{ with .datum.uuid }}{{ . }}{{ end }}
:END:
{{ with .datum.entry }}{{ . }}{{ end }}{{ end }}`

	storgFileTemplateOut, err := template.New("nodes").Parse(storgFileTemplateStr)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("test.org")

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	//	var tpl bytes.Buffer
	err = storgFileTemplateOut.Execute(file, storgFileMap)
	if err != nil {
		panic(err)
	}

}
