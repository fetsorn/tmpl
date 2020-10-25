package main

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/niklasfasching/go-org/org"
	// "io/ioutil"
	// "strings"
	"encoding/json"
)

func main() {
	// var str = `{"UUID": "1282","GUEST":"fetsorn"}`
	// var str = `[{"UUID": "1282","GUEST":"fetsorn"},{"UUID": "2522","GUEST":"mamagpa"}]`
	var str = "* heading\nsometext"
	// var str = "* heading\n:PROPERTIES:\n:GUEST: fetsorn\n:END:"
	// var str = "###"

	// check if string is a json element
	var jsonElement map[string]interface{}

	if err := json.Unmarshal([]byte(str), &jsonElement); err == nil {
		fmt.Println(jsonElement)
		return
	}

	// check if string is json array
	var jsonArray []map[string]interface{}
	if err := json.Unmarshal([]byte(str), &jsonArray); err == nil {
		fmt.Println(jsonArray)
		return
	}

	// check if string is biorg
	jsonArray = nil

	document := org.New().Parse(bytes.NewReader([]byte(str)), "")

	for _, node := range document.Nodes {

		switch node := node.(type) {

		case org.Headline:

			if node.Lvl == 1 {

				if node.Properties == nil {
					fmt.Println("Found a first level org heading without a property drawer, stopping...")
					fmt.Println(node)
					return
				}

				var nodeMap = map[string]interface{}{}

				for _, pair := range node.Properties.Properties {
					// map property key to the value
					nodeMap[pair[0]] = pair[1]
				}

				var datum string

				for _, child := range node.Children {
					datum += org.NewOrgWriter().WriteNodesAsString(child)
				}

				nodeMap["DATUM"] = datum

				nodeMap["UUID"] = uuid.New().String()

				jsonArray = append(jsonArray, nodeMap)

			}
		}
	}

	if jsonArray != nil {
		fmt.Println(jsonArray)
		return
	}

	fmt.Println("Input is not json or biorg, stopping...")
	return
}
