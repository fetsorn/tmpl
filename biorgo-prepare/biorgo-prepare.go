package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	// "time"
)

type FilePrepared struct {
	Name       string `json:"name"`
	ModTime    string `json:"mod_time"`
	Size       int64  `json:"size"`
	Path       string `json:"path"`
	ParsedTime string `json:"parsed_time"`
}

// tell if the folder content is noise and should be skipped
func recursiveIsNoise(path string, array []string) bool {

	for noise := range array {
		re := regexp.MustCompile(array[noise])
		if re.MatchString(path) {
			return true
		}
	}

	return false
}

// tell if the file extension is noise and it should be ignored
func fileIsNoise(path string, array []string) bool {

	for noise := range array {
		re := regexp.MustCompile(array[noise])
		if re.MatchString(path) {
			return true
		}
	}

	return false
}

// prepare files in directory inputPath according to a go template in templatePath, output as desmi to outputPath
func prepareFiles(inputPath string, templatePath string, outputPath string, ignorePath string) {

	jsonArray := []map[string]interface{}{}

	var recursiveArray []string
	var fileArray []string

	if ignorePath == "empty" {

		recursiveArray = []string{}
		fileArray = []string{}
	} else {

		// read ignore file
		ignoreFile, err := ioutil.ReadFile(ignorePath)
		if err != nil {
			fmt.Println("Failed to read ignore file", err)
			panic(err)
		}

		ignoreStr := string(ignoreFile)
		ignoreArray := strings.Split(ignoreStr, "\n---\n")

		recursiveArray = strings.Split(ignoreArray[0], "\n")
		// remove last empty element or it will match everything
		recursiveArray = recursiveArray[:len(recursiveArray)-1]

		fileArray = strings.Split(ignoreArray[1], "\n")
		// remove last empty element or it will match everything
		fileArray = fileArray[:len(fileArray)-1]
	}

	// parse the file into FilePrepared struct, append to array
	var walkFunction = func(path string, info os.FileInfo, err error) error {

		var file FilePrepared

		if info.IsDir() && recursiveIsNoise(path, recursiveArray) {
			return filepath.SkipDir
		}

		if !info.IsDir() && !fileIsNoise(path, fileArray) {

			layout := "<2006-01-02>"
			modTime := info.ModTime().Format(layout)

			file.Name = info.Name()
			file.ModTime = modTime
			//file.Size = info.Size()
			file.Path = path
			//file.ParsedTime = time.Now().Format(layout)

			jsonStr, _ := json.Marshal(file)

			var jsonMap map[string]interface{}
			if err := json.Unmarshal(jsonStr, &jsonMap); err != nil {
				fmt.Println("Failed to unmarshal", err)
				panic(err)
			}

			jsonArray = append(jsonArray, jsonMap)

		}

		return err

	}

	// walk the directory inputPath, apply walkFunction to each file
	err := filepath.Walk(inputPath, walkFunction)
	if err != nil {
		fmt.Println("error walking the path %q: %v\n", inputPath, err)
	}

	fmt.Println(inputPath, " - ", len(jsonArray), "nodes")

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
		fmt.Println("Failed to prepare template", err)
		panic(err)
	}

	if outputPath == "empty" {

		err = templateStruct.Execute(os.Stdout, jsonArray)
		if err != nil {
			fmt.Println("Failed to execute template", err)
			panic(err)
		}
	} else {

		// create a file at outputPath
		file, err := os.Create(outputPath)
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}

		// execute template over the array of FilePrepared, write to output file
		err = templateStruct.Execute(file, jsonArray)
		if err != nil {
			fmt.Println("Failed to execute template", err)
			panic(err)
		}
	}

}

func main() {

	var inputPath string
	var templatePath string
	var outputPath string
	var ignorePath string

	// flags declaration using flag package
	flag.StringVar(&inputPath, "i", "empty", "Please specify storg path")
	flag.StringVar(&templatePath, "t", "empty", "Please specify template path")
	flag.StringVar(&outputPath, "o", "empty", "Please specify output path")
	flag.StringVar(&ignorePath, "e", "empty", "Please specify path to ignore file")

	flag.Parse() // after declaring flags we need to call it
	// TODO: show usage when no arguments are specified

	prepareFiles(inputPath, templatePath, outputPath, ignorePath)

}
