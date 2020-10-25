package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"testing"
)

// https://eli.thegreenplace.net/2020/faking-stdin-and-stdout-in-go/
// FakeStdio can be used to fake stdin and capture stdout.
// Between creating a new FakeStdio and calling ReadAndRestore on it,
// code reading os.Stdin will get the contents of stdinText passed to New.
// Output to os.Stdout will be captured and returned from ReadAndRestore.
// FakeStdio is not reusable; don't attempt to use it after calling
// ReadAndRestore, but it should be safe to create a new FakeStdio.
type FakeStdio struct {
	origStdout   *os.File
	stdoutReader *os.File

	outCh chan []byte

	origStdin   *os.File
	stdinWriter *os.File
}

func NewStdio(stdinText string) (*FakeStdio, error) {
	// Pipe for stdin.
	//
	//                 ======
	//  w ------------->||||------> r
	// (stdinWriter)   ======      (os.Stdin)
	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	// Pipe for stdout.
	//
	//               ======
	//  w ----------->||||------> r
	// (os.Stdout)   ======      (stdoutReader)
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	origStdin := os.Stdin
	os.Stdin = stdinReader

	_, err = stdinWriter.Write([]byte(stdinText))
	if err != nil {
		stdinWriter.Close()
		os.Stdin = origStdin
		return nil, err
	}

	origStdout := os.Stdout
	os.Stdout = stdoutWriter

	outCh := make(chan []byte)

	// This goroutine reads stdout into a buffer in the background.
	go func() {
		var b bytes.Buffer
		if _, err := io.Copy(&b, stdoutReader); err != nil {
			log.Println(err)
		}
		outCh <- b.Bytes()
	}()

	return &FakeStdio{
		origStdout:   origStdout,
		stdoutReader: stdoutReader,
		outCh:        outCh,
		origStdin:    origStdin,
		stdinWriter:  stdinWriter,
	}, nil
}

// ReadAndRestore collects all captured stdout and returns it; it also restores
// os.Stdin and os.Stdout to their original values.
func (sf *FakeStdio) ReadAndRestore() ([]byte, error) {
	if sf.stdoutReader == nil {
		return nil, fmt.Errorf("ReadAndRestore from closed FakeStdio")
	}

	// Close the writer side of the faked stdout pipe. This signals to the
	// background goroutine that it should exit.
	os.Stdout.Close()
	out := <-sf.outCh

	os.Stdout = sf.origStdout
	os.Stdin = sf.origStdin

	if sf.stdoutReader != nil {
		sf.stdoutReader.Close()
		sf.stdoutReader = nil
	}

	if sf.stdinWriter != nil {
		sf.stdinWriter.Close()
		sf.stdinWriter = nil
	}

	return out, nil
}

// CloseStdin closes the fake stdin. This may be necessary if the process has
// logic for reading stdin until EOF; otherwise such code would block forever.
func (sf *FakeStdio) CloseStdin() {
	if sf.stdinWriter != nil {
		sf.stdinWriter.Close()
		sf.stdinWriter = nil
	}
}

func TestIsInputFromPipe(t *testing.T) {

	// Create a new fakestdio with some input to feed into Stdin.
	_, err := NewStdio("input text")
	if err != nil {
		log.Fatal(err)
	}

	expected := true
	actual := isInputFromPipe()
	if actual != expected {
		t.Errorf("Test failed, expected: '%t', got:  '%t'", expected, actual)
	}

}

func TestParseJSONElement(t *testing.T) {

	input := []byte(`{"KEY":"value"}`)
	var actual map[string]interface{}
	actual, _ = parseJSONElement(input)

	var expected = map[string]interface{}{"KEY": "value"}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}

func TestParseJSONArray(t *testing.T) {

	input := []byte(`[{"KEY1":"value1"},{"KEY2":"value2"}]`)
	var actual []map[string]interface{}
	actual, _ = parseJSONArray(input)

	var map1 = map[string]interface{}{"KEY1": "value1"}
	var map2 = map[string]interface{}{"KEY2": "value2"}
	var expected = []map[string]interface{}{map1, map2}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}

func TestParseBiorg(t *testing.T) {

	input := []byte(`
* heading1
:PROPERTIES:
:KEY1: value1
:END:
* heading2
:PROPERTIES:
:KEY2: value2
:END:
`)
	var actual []map[string]interface{}
	actual, _ = parseBiorg(input)

	var map1 = map[string]interface{}{"DATUM": "", "KEY1": "value1", "UUID": "mock"}
	var map2 = map[string]interface{}{"DATUM": "", "KEY2": "value2", "UUID": "mock"}
	var expected = []map[string]interface{}{map1, map2}

	if expected[0]["KEY1"] != actual[0]["KEY1"] {
		t.Errorf("Test failed, expected: '%v', got:  '%v'", expected, actual)
	}
}
