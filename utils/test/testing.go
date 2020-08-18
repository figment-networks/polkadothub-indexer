package test

import (
	"io/ioutil"
	"os"
	"testing"
)

// CreateFile create a file with given data
func CreateFile(t *testing.T, fileName string, data []byte) {
	f, err := os.Create(fileName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	err = ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

// CleanUp remove file
func CleanUp(t *testing.T, fileName string) {
	if err := os.RemoveAll(fileName); err != nil {
		t.Error("could not remove file", err)
	}
}
