package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/goccy/go-graphviz"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"
)

// sort storg by key in ascending order
func sortStorg(storg []map[string]interface{}, key string) []map[string]interface{} {

	sort.SliceStable(storg, func(i, j int) bool {

		nodeLast := storg[i]
		nodeNext := storg[j]

		keyLast := nodeLast[key].(string)
		keyNext := nodeNext[key].(string)

		return keyLast < keyNext
	})

	return storg
}

// filter storg by key value
func filterStorg(storgOld []map[string]interface{}, key string, keyword string) []map[string]interface{} {

	storgNew := []map[string]interface{}{}

	for _, node := range storgOld {

		value := node[key].(string)

		if value == keyword {
			storgNew = append(storgNew, node)
		}
	}

	return storgNew
}

// filter storg by key in a time period between start and end
func betweenDates(storgOld []map[string]interface{}, key string, start string, end string) []map[string]interface{} {

	storgNew := []map[string]interface{}{}

	for _, node := range storgOld {

		// get the key value
		date := node[key].(string)

		// parse dates to time.time
		layout := "<2006-01-02>"

		timeNode, err := time.Parse(layout, date)
		if err != nil {
			fmt.Println("Failed to parse the node date: ", date)
			continue
		}

		timeStart, err := time.Parse(layout, start)
		if err != nil {
			fmt.Println("Failed to parse the start date: ", start)
			continue
		}

		timeEnd, err := time.Parse(layout, end)
		if err != nil {
			fmt.Println("Failed to parse the end date: ", end)
			continue
		}

		// if the node time is between start and end, append it to the array
		if timeNode.After(timeStart) && timeNode.Before(timeEnd) {
			storgNew = append(storgNew, node)
		}
	}

	// return nodes with key between start and end
	return storgNew
}

// filter storg oldArray by unique key
func uniqueDate(storgArray []map[string]interface{}, key string) []map[string]interface{} {

	storgSet := []map[string]interface{}{}

	// iterate over the array, append unique nodes to the set
loopOuter:
	for _, nodeY := range storgArray {

		// get the key value of the next node
		dateY := nodeY[key].(string)

		// check if the set already has an element with that key value
		for _, nodeX := range storgSet {

			// get the key value of the previous node
			dateX := nodeX[key].(string)

			// if an element of the set has the same value, continue to the next node
			if dateX == dateY {
				continue loopOuter
			}
		}

		// if the node is unique, append it to the set
		storgSet = append(storgSet, nodeY)
	}

	return storgSet
}

// tell if node index is divisible by 2400
// allows to partition dates in dot notation and avoid graphviz error
func reached2400(index int) bool {
	return index != 0 && index%2400 == 0
}

// parse storg json
func parseStorg(storgPath string) []map[string]interface{} {

	// read a storg file
	storgFile, err := ioutil.ReadFile(storgPath)
	if err != nil {
		fmt.Println("File reading error", err)
		return nil
	}

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

	customFunctions := template.FuncMap{"sortStorg": sortStorg, "filterStorg": filterStorg, "betweenDates": betweenDates}

	templateStruct, err := template.New("nodesDesmi").Funcs(customFunctions).Parse(string(templateString))
	if err != nil {
		panic(err)
	}

	if outputPath == "empty" {

		err = templateStruct.Execute(os.Stdout, storg)
		if err != nil {
			panic(err)
		}
	} else {

		file, err := os.Create(outputPath)
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}

		err = templateStruct.Execute(file, storg)
		if err != nil {
			panic(err)
		}
	}
}

// generate dot notation
func generateDot(storg []map[string]interface{}, templatePath string, outputPath string) {

	// format storg entries to prevent graphviz errors
	for _, node := range storg {
		// remove symbols instead of escaping because backslashes might otherwise escape closing quotes
		// DO NOT REUSE FOR RAVDIA, BREAKS VALIDITY
		// remote newlines
		datum := node["DATUM"].(string)
		node["DATUM"] = strings.Replace(datum, "\n", "", -1)
		// remove quotes
		datum = node["DATUM"].(string)
		node["DATUM"] = strings.Replace(datum, "\"", "", -1)
		// remove the line tabulation character
		datum = node["DATUM"].(string)
		node["DATUM"] = strings.Replace(datum, "", "", -1)
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

	if outputPath == "empty" {

		err = templateStruct.Execute(os.Stdout, storg)
		if err != nil {
			panic(err)
		}
	} else {

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

}

// generate ravdia org files
func generateRavdia(storg []map[string]interface{}, templatePath string, outputPath string) {

	// create outputPath if it does not not exist
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		err = os.Mkdir(outputPath, 0755)
		if err != nil {
			panic(err)
		}
	}

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

		uuid := node["UUID"].(string)

		file, err := os.Create(outputPath + "/" + uuid + ".html")
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

// generate svg from dot
func generateSVG(dotPath string) {

	g := graphviz.New()

	b, err := ioutil.ReadFile(dotPath)
	if err != nil {
		log.Fatal(err)
	}
	graph, err := graphviz.ParseBytes(b)

	if err := g.RenderFilename(graph, graphviz.SVG, "test.svg"); err != nil {
		log.Fatal(err)
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
		// generateSVG(outputPath)
	} else if reportType == "ravdia" {
		var storg = parseStorg(inputPath)
		generateRavdia(storg, templatePath, outputPath)
	}
}
