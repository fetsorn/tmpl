package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

type FilePrepared struct {
	Name         string `json:"FILE_NAME"`
	ModDate      string `json:"MOD_DATE"`
	Size         int64  `json:"SIZE"`
	PathRelative string `json:"PATH_RELATIVE"`
	PathAbsolute string `json:"PATH_ABSOLUTE"`
}

// tell if input is from a pipe
// from https://stackoverflow.com/a/26567513
func isInputFromPipe() bool {

	fileInfo, _ := os.Stdin.Stat()

	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func newUUID() string {
	return uuid.New().String()
}

func parseStat(filePath string) ([]byte, error) {

	var jsonBytes []byte

	// get os.FileInfo for the file
	info, err := os.Stat(filePath)
	if err != nil {
		log.Println("Failed to get os.FileInfo", err)
		return jsonBytes, nil
	}

	// initialize struct for gathering stats
	var input FilePrepared

	// ISO format dates
	layout := "<2006-01-02>"
	// last modification time
	input.ModDate = info.ModTime().Format(layout)

	// file name
	input.Name = info.Name()
	// file size in Kb
	input.Size = info.Size()
	// file path relative to current directory
	input.PathRelative = filePath
	// file path absolute
	input.PathAbsolute, _ = filepath.Abs(filePath)

	// encode prepared data to json
	jsonBytes, err = json.Marshal(input)
	if err != nil {
		return jsonBytes, err
	}

	return jsonBytes, err

}

func outputStat(jsonBytes []byte, templatePath string) error {

	// print json to stdout by default
	if templatePath == "" {

		_, err := fmt.Println(string(jsonBytes))

		return err
	} else {

		// decode json to get a map that templates can read
		var jsonMap map[string]interface{}

		if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
			log.Println("Failed to unmarshal", err)
			return err
		}

		// read template from path
		templateString, err := ioutil.ReadFile(templatePath)
		if err != nil {

			log.Println("Template file reading error", err)
			return err
		}

		// specify custom template functions
		customFunctions := template.FuncMap{"newUUID": newUUID}

		// initialize template with custom functions
		templateStruct, err := template.New("nodes").Funcs(customFunctions).Parse(string(templateString))
		if err != nil {
			return err
		}

		// execute template on json to stdout
		return templateStruct.Execute(os.Stdout, jsonMap)
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
			Usage:   "path to input file",
		},
	}

	app.Action = func(c *cli.Context) error {

		if isInputFromPipe() {

			scanner := bufio.NewScanner(os.Stdin)

			for scanner.Scan() {
				jsonBytes, err := parseStat(scanner.Text())
				if err != nil {
					log.Println("Stat parsing error", err)
					return nil
				}

				// get path to golang template from flag
				outputStat(jsonBytes, c.String("template"))
			}

			if err := scanner.Err(); err != nil {
				log.Println(err)
				return nil
			}
		} else {

			// get path to input file from flag
			filePath := c.String("file")
			if filePath != "" {
				jsonBytes, err := parseStat(filePath)
				if err != nil {
					log.Println("Stat parsing error", err)
					return nil
				}

				// get path to golang template from flag
				return outputStat(jsonBytes, c.String("template"))
			}
		}

		return nil
	}

	// parse cli flags
	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
	}

}
