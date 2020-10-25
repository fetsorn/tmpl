package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/niklasfasching/go-org/org"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
	// "strings"
	"encoding/json"
)

// tbn

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

// format strings for dot
func formatStringDot(str string) string {

	// remove symbols instead of escaping
	// otherwise backslashes might escape closing quotes in dot
	// remove newlines
	str = strings.Replace(str, "\n", "", -1)
	// remove quotes
	str = strings.Replace(str, "\"", "", -1)
	// remove the line tabulation character
	str = strings.Replace(str, "", "", -1)

	return str
}

// format strings for biorg
func formatStringBiorg(str string) string {

	// increment all org headings by one level
	// legacy solution that imitates previous manual commits
	re := regexp.MustCompile(`(?m)^\*`)
	str = re.ReplaceAllString(str, "**")

	return str
}

// tbn

// tell if input is from a pipe
// from https://stackoverflow.com/a/26567513
func isInputFromPipe() bool {

	fileInfo, _ := os.Stdin.Stat()

	return fileInfo.Mode()&os.ModeCharDevice == 0
}

// check if string is a json element
func parseJSONElement(input []byte) (map[string]interface{}, error) {

	var jsonElement map[string]interface{}

	err := json.Unmarshal(input, &jsonElement)

	return jsonElement, err
}

// check if string is json array
func parseJSONArray(input []byte) ([]map[string]interface{}, error) {

	var jsonArray []map[string]interface{}

	err := json.Unmarshal(input, &jsonArray)

	return jsonArray, err
}

// check if string is biorg
func parseBiorg(input []byte) ([]map[string]interface{}, error) {

	var jsonArray = []map[string]interface{}{}

	document := org.New().Parse(bytes.NewReader(input), "")

	// append all first level headlines to jsonArray
	for _, node := range document.Nodes {

		switch node := node.(type) {

		case org.Headline:

			if node.Lvl == 1 {

				if node.Properties == nil {
					log.Println("Found a first level org heading without a property drawer, stopping...")
					log.Println(node)
					err := errors.New("Org heading without a property drawer")
					return jsonArray, err
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

	var err error = nil

	if jsonArray == nil {
		err = errors.New("Biorg parser returned empty")
	}

	return jsonArray, err
}

// prepare temlate
func prepareTemplate(templatePath string) (*template.Template, error) {

	templateString, err := ioutil.ReadFile(templatePath)
	if err != nil {

		log.Println("File reading error", err)

		return nil, err
	}

	customFunctions := template.FuncMap{"uniqueDate": uniqueDate, "reached2400": reached2400, "sortStorg": sortStorg, "filterStorg": filterStorg, "betweenDates": betweenDates, "formatStringDot": formatStringDot}

	return template.New("nodesDesmi").Funcs(customFunctions).Parse(string(templateString))
}

// marshal model or execute template to stdout
func outputElement(element map[string]interface{}, templatePath string) error {

	if templatePath == "" {

		jsonBytes, _ := json.Marshal(element)

		_, err := fmt.Println(string(jsonBytes))

		return err
	} else {

		templateStruct, err := prepareTemplate(templatePath)

		if err != nil {

			return err
		}

		return templateStruct.Execute(os.Stdout, element)
	}

}

// marshal model or execute template to stdout
func outputArray(array []map[string]interface{}, templatePath string) error {

	if templatePath == "" {

		jsonBytes, _ := json.Marshal(array)

		_, err := fmt.Println(string(jsonBytes))
		return err
	} else {

		templateStruct, err := prepareTemplate(templatePath)
		if err != nil {
			return err
		}

		return templateStruct.Execute(os.Stdout, array)
	}

}

func main() {

	app := cli.NewApp()

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "template",
			Aliases: []string{"t"},
			Value:   "",
			Usage:   "path to go template",
		},
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Value:   "",
			Usage:   "path to input file",
		},
	}

	app.Action = func(c *cli.Context) error {

		var input []byte

		// read input from stdin if it is not empty
		if isInputFromPipe() {

			var err error
			input, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
		} else {

			// read input from file if it is provided
			file, err := os.Open(c.String("file"))
			if err != nil {
				// TODO: show help if now file or stdin is provided
				return err
			}

			input, err = ioutil.ReadAll(file)
			if err != nil {
				return err
			}

			file.Close()
		}

		// try to parse input as json element
		if element, err := parseJSONElement(input); err == nil {

			return outputElement(element, c.String("template"))
		}
		log.Println("Not a json element")

		// try to parse input as json array
		if array, err := parseJSONArray(input); err == nil {

			return outputArray(array, c.String("template"))
		}
		log.Println("Not a json array")

		// try to parse input as biorg
		if array, err := parseBiorg(input); err == nil {

			return outputArray(array, c.String("template"))
		}
		log.Println("Not biorg")

		log.Println("Input is not json or biorg, stopping...")

		err := errors.New("All parsers failed")
		return err
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
