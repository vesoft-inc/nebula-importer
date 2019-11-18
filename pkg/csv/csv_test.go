package csv

import (
	"encoding/csv"
	"os"
	"testing"
)

func TestCsvWriter(t *testing.T) {
	file, err := os.OpenFile("./test.csv", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	w := csv.NewWriter(file)

	if err := w.Write([]string{"hash(\"hello\")", "234"}); err != nil {
		t.Fatal(err)
	}
	w.Flush()
	if w.Error() != nil {
		t.Fatal(w.Error())
	}
}
