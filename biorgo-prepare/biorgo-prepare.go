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
	"time"
)

type FilePrepared struct {
	Name       string `json:"name"`
	ModTime    string `json:"mod_time"`
	Size       int64  `json:"size"`
	Path       string `json:"path"`
	ParsedTime string `json:"parsed_time"`
	Contents   string `json:"contents"`
	// UUID       string `json:"uuid"`
}

// tell if the file has useful data
func fileIsUseful(path string) bool {
	return !strings.HasSuffix(path, ".DS_Store")
}

// tell if the file is plain text and its contents should be read
func fileIsText(path string) bool {
	fileExt := filepath.Ext(path)
	return fileExt == ".txt" || fileExt == ".org" || fileExt == ".md"
}

// prepare files in directory inputPath according to a go template in templatePath, output as desmi to outputPath
func prepareFiles(inputPath string, templatePath string, outputPath string) {

	jsonArray := []map[string]interface{}{}

	// parse the file into FilePrepared struct, append to array
	var walkFunction = func(path string, info os.FileInfo, err error) error {

		if !info.IsDir() && fileIsUseful(path) {
			var file FilePrepared

			layout := "<2006-01-02>"
			modTime := info.ModTime().Format(layout)

			file.Name = info.Name()
			file.ModTime = modTime
			file.Size = info.Size()
			file.Path = path
			file.ParsedTime = time.Now().Format(layout)

			// add contents of plain text files
			if fileIsText(path) {

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

	if outputPath == "empty" {

		err = templateStruct.Execute(os.Stdout, jsonArray)
		if err != nil {
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
			panic(err)
		}
	}

}

func main() {

	var inputPath string
	var templatePath string
	var outputPath string

	// flags declaration using flag package
	flag.StringVar(&inputPath, "i", "empty", "Please specify storg path")
	flag.StringVar(&templatePath, "t", "empty", "Please specify template path")
	flag.StringVar(&outputPath, "o", "empty", "Please specify output path")

	flag.Parse() // after declaring flags we need to call it
	// TODO: show usage when no arguments are specified

	prepareFiles(inputPath, templatePath, outputPath)
}
