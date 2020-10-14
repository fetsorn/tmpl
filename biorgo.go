package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"
	// "time"
)

type Todo struct {
	Name        string
	Description string
}

type Storg struct {
	Metadata Storg_meta `json:"metadatum"`
	Data     Storg_data `json:"datum"`
}

type Storg_meta struct {
	Comment    string `json:"COMMENT"`
	Guest_date string `json:"GUEST_DATE"`
	Guest      string `json:"GUEST"`
	Host_date  string `json:"HOST_DATE"`
	Host       string `json:"HOST"`
	Label      string `json:"LABEL"`
	Module     string `json:"MODULE"`
	Type       string `json:"TYPE"`
}

type Storg_data struct {
	Uuid  string `json:"uuid"`
	Entry string `json:"entry"`
}

func main() {
	todo1 := Todo{"Test templates", "Let's test a template to see the magic."}

	var todoTemplateStr = "You have a task named \"{{ .Name }}\" with description: \"{{ .Description }}\""

	todoTemplateOut, err := template.New("todos").Parse(todoTemplateStr)
	if err != nil {
		panic(err)
	}

	err = todoTemplateOut.Execute(os.Stdout, todo1)
	if err != nil {
		panic(err)
	}

	fmt.Println(" ")

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

	var storgMap1 map[string]interface{}

	if err := json.Unmarshal(storgJSON, &storgMap1); err != nil {
		panic(err)
	}
	fmt.Println(storgMap1)

	fmt.Println(" ")

	storgMap2 := Storg{}

	json.Unmarshal([]byte(storgJSON), &storgMap2)

	fmt.Println(storgMap2)

	var storgTemplateStr = `
	* .
	:PROPERTIES:
	:HOST: {{ .metadatum.HOST }}
	:HOST_DATE: {{ if .metadatum.HOST_DATE }} {{ .metadatum.HOST_DATE }} {{ end }}
	`

	// Option("missinkey=zero") prevents the <no value> for non-existent fields
	storgTemplateOut, err := template.New("nodes").Option("missingkey=zero").Parse(storgTemplateStr)
	if err != nil {
		panic(err)
	}

	err = storgTemplateOut.Execute(os.Stdout, storgMap1)
	if err != nil {
		panic(err)
	}

}
