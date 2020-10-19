package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/niklasfasching/go-org/org"
	"io/ioutil"
	"log"
	"os"
)

func parseBiorgToJSON(inputPath string, outputPath string) {

	jsonArray := []map[string]interface{}{}

	input, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return
	}
	document := org.New().Parse(bytes.NewReader(input), inputPath)

	for _, node := range document.Nodes {
		switch node := node.(type) {
		case org.Headline:
			if node.Lvl == 1 {

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

				jsonArray = append(jsonArray, nodeMap)

			}
		}
	}

	// create a file at outputPath
	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	storg, _ := json.Marshal(jsonArray)

	file.WriteString(string(storg))

}

func main() {

	var inputPath string
	var outputPath string

	// flags declaration using flag package
	flag.StringVar(&inputPath, "i", "empty", "Please specify storg path")
	flag.StringVar(&outputPath, "o", "empty", "Please specify output path")

	flag.Parse() // after declaring flags we need to call it

	parseBiorgToJSON(inputPath, outputPath)
	// TODO: show usage when no arguments are specified
}
