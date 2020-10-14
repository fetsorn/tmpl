package main

import (
	"encoding/json"
	"fmt"
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

}
