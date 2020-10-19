// Copyright © 2020 Anton Davydov <fetsorn@gmail.com>.

package main

// Biorgo generates reports from a JSON array of Biorg entries called storg
import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/niklasfasching/go-org/org"
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

func fileIsLegible(path string) bool {
	return !strings.HasSuffix(path, ".DS_Store")
}

// tbn

// feed json to a template, print a biorg entry
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

// tbn

type Biorg struct {
	Metadatum struct {
		Comment    string `json:"COMMENT"`
		Guest_date string `json:"GUEST_DATE"`
		Guest      string `json:"GUEST"`
		Host_date  string `json:"HOST_DATE"`
		Host       string `json:"HOST"`
		Label      string `json:"LABEL"`
		Module     string `json:"MODULE"`
		Type       string `json:"TYPE"`
	} `json:"metadatum"`
	Datum struct {
		Uuid  string `json:"uuid"`
		Entry string `json:"entry"`
	} `json:"datum"`
}

// parse Biorg node to json
func parseNodeToJSONOld(node org.Headline) string {

	var biorg Biorg

	biorg.Metadatum.Comment, _ = node.Properties.Get("COMMENT")
	biorg.Metadatum.Guest_date, _ = node.Properties.Get("GUEST_DATE")
	biorg.Metadatum.Guest, _ = node.Properties.Get("GUEST")
	biorg.Metadatum.Host_date, _ = node.Properties.Get("HOST_DATE")
	biorg.Metadatum.Host, _ = node.Properties.Get("HOST")
	biorg.Metadatum.Label, _ = node.Properties.Get("LABEL")
	biorg.Metadatum.Module, _ = node.Properties.Get("MODULE")
	biorg.Metadatum.Type, _ = node.Properties.Get("TYPE")
	biorg.Datum.Uuid, _ = node.Properties.Get("UUID")

	var entry string

	for _, child := range node.Children {
		entry += org.NewOrgWriter().WriteNodesAsString(child)
	}

	biorg.Datum.Entry = entry

	storg, _ := json.Marshal(biorg)

	return string(storg)
}

// parse Biorg file to json
func parseBiorgToJSONOld(inputPath string, outputPath string) {

	bs, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return
	}
	d := org.New().Parse(bytes.NewReader(bs), inputPath)

	if outputPath == "empty" {
		fmt.Println("[")
		for _, node := range d.Nodes {
			switch node := node.(type) {
			case org.Headline:
				if node.Lvl == 1 {
					fmt.Println(parseNodeToJSONOld(node))
					fmt.Print(",")
				}
			}
		}
		fmt.Println("{}")
		fmt.Println("]")
	} else {
		file, err := os.Create(outputPath)
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}

		file.WriteString("[")
		for _, node := range d.Nodes {
			switch node := node.(type) {
			case org.Headline:
				if node.Lvl == 1 {
					file.WriteString(parseNodeToJSONOld(node))
					file.WriteString(",")
				}
			}
		}
		file.WriteString("{}")
		file.WriteString("]")
		file.Close()
	}
}

func parseNodeToJSON(node org.Headline) map[string]interface{} {

	var properties [][]string = node.Properties.Properties

	var nodeMap = map[string]interface{}{}

	for _, pair := range properties {
		// map first element of the pair array to the second
		nodeMap[pair[0]] = pair[1]
	}

	return nodeMap
}

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
					// map first element of the pair array to the second
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

	// read a go template from templatePath
	// templateString, err := ioutil.ReadFile(templatePath)
	// if err != nil {
	// 	fmt.Println("File reading error", err)
	// 	return
	// }

	// prepare the template
	// customFunctions := template.FuncMap{}

	// templateStruct, err := template.New("nodesPrepared").Funcs(customFunctions).Parse(string(templateString))
	// if err != nil {
	// 	panic(err)
	// }

	// create a file at outputPath
	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	storg, _ := json.Marshal(jsonArray)
	file.WriteString(string(storg))

	// execute template over the array of FilePrepared, write to output file
	// err = templateStruct.Execute(file, jsonArray)
	// if err != nil {
	// 	panic(err)
	// }

}

type FilePrepared struct {
	Name       string `json:"name"`
	ModTime    string `json:"mod_time"`
	Size       int64  `json:"size"`
	Path       string `json:"path"`
	ParsedTime string `json:"parsed_time"`
	Contents   string `json:"contents"`
	// UUID       string `json:"uuid"`
}

// prepare files in directory inputPath according to a go template in templatePath, output as desmi to outputPath
func prepareFiles(inputPath string, templatePath string, outputPath string) {

	jsonArray := []map[string]interface{}{}

	// parse the file into FilePrepared struct, append to array
	var walkFunction = func(path string, info os.FileInfo, err error) error {

		if !info.IsDir() && fileIsLegible(path) {
			var file FilePrepared

			layout := "<2006-01-02>"
			modTime := info.ModTime().Format(layout)

			file.Name = info.Name()
			file.ModTime = modTime
			file.Size = info.Size()
			file.Path = path
			file.ParsedTime = time.Now().Format(layout)

			//add contents of text files
			fileExt := filepath.Ext(path)
			if fileExt == ".org" || fileExt == ".org" || fileExt == ".md" {

				entry, err := ioutil.ReadFile(path)
				if err != nil {
					fmt.Printf("error reading file %v", path)
				}

				file.Contents = string(entry)

				// increment all org headings by one level
				// legacy solution that imitates previous manual commits
				re := regexp.MustCompile(`(?m)^\*`)
				file.Contents = re.ReplaceAllString(file.Contents, "**")
			}

			//add uuid to identify the node in Storg
			// file.UUID = uuid.New().String()

			jsonStr, _ := json.Marshal(file)

			var jsonMap map[string]interface{}
			if err := json.Unmarshal(jsonStr, &jsonMap); err != nil {
				panic(err)
			}

			jsonArray = append(jsonArray, jsonMap)

		}

		return err
	}

	// walk the directory inputPath, apply walkFunction to each file
	err := filepath.Walk(inputPath, walkFunction)
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", inputPath, err)
	}

	// read a go template from templatePath
	templateString, err := ioutil.ReadFile(templatePath)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	// prepare the template
	customFunctions := template.FuncMap{}

	templateStruct, err := template.New("nodesPrepared").Funcs(customFunctions).Parse(string(templateString))
	if err != nil {
		panic(err)
	}

	// create a file at outputPath
	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	// execute template over the array of FilePrepared, write to output file
	err = templateStruct.Execute(file, jsonArray)
	if err != nil {
		panic(err)
	}

}

// prepare files in directory inputPath for biorg
func prepareFilesOld(inputPath string, templatePath string, outputPath string) {

	storgArray := []map[string]interface{}{}

	// parse file into the Biorg struct
	var walkFunction = func(path string, info os.FileInfo, err error) error {

		if !info.IsDir() {
			var biorg Biorg

			layout := "<2006-01-02>"
			modTime := info.ModTime().Format(layout)

			biorg.Metadatum.Comment = ""
			biorg.Metadatum.Guest_date = modTime
			biorg.Metadatum.Guest = "fetsorn"
			biorg.Metadatum.Host_date = modTime
			biorg.Metadatum.Host = "fetsorn"
			biorg.Metadatum.Label = ""
			biorg.Metadatum.Module = ""
			biorg.Metadatum.Type = ""
			biorg.Datum.Uuid = uuid.New().String()

			entry, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Printf("error reading file %v", path)
			}

			biorg.Datum.Entry = string(entry)

			// increment all org headings by one level
			// legacy solution that imitates previous manual commits
			re := regexp.MustCompile(`(?m)^\*`)
			biorg.Datum.Entry = re.ReplaceAllString(biorg.Datum.Entry, "**")

			storgNode, _ := json.Marshal(biorg)

			var storgMap map[string]interface{}

			if err := json.Unmarshal(storgNode, &storgMap); err != nil {
				panic(err)
			}

			storgArray = append(storgArray, storgMap)
		}
		return err
	}

	err := filepath.Walk(inputPath, walkFunction)
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", inputPath, err)
	}

	outputType := filepath.Ext(outputPath)
	if outputType == ".org" {
		generateDesmi(storgArray, templatePath, outputPath)
	} else if outputType == ".json" {
		storg, _ := json.Marshal(storgArray)
		file, _ := os.Create(outputPath)
		file.WriteString(string(storg))
	}

}

func main() {

	var commandType string
	var inputPath string
	var templatePath string
	var outputPath string

	// flags declaration using flag package
	flag.StringVar(&commandType, "c", "empty", "Please specify command: desmi, dot, ravdia, json, prepare")
	flag.StringVar(&inputPath, "i", "empty", "Please specify storg path")
	flag.StringVar(&templatePath, "t", "empty", "Please specify template path")
	flag.StringVar(&outputPath, "o", "empty", "Please specify output path")

	flag.Parse() // after declaring flags we need to call it

	if commandType == "desmi" {
		var storg = parseStorg(inputPath)
		generateDesmi(storg, templatePath, outputPath)
	} else if commandType == "dot" {
		var storg = parseStorg(inputPath)
		generateDot(storg, templatePath, outputPath)
	} else if commandType == "ravdia" {
		var storg = parseStorg(inputPath)
		generateRavdia(storg, templatePath, outputPath)
	} else if commandType == "json" {
		parseBiorgToJSON(inputPath, outputPath)
	} else if commandType == "prepare" {
		prepareFiles(inputPath, templatePath, outputPath)
	}
	// TODO: show usage when no arguments are specified
}
