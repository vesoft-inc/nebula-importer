package csv

import (
	"encoding/csv"
	"os"
	"testing"
)

func TestCsvWriter(t *testing.T) {
	file, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()

	w := csv.NewWriter(file)

	if err := w.Write([]string{"hash(\"hello\")", "234"}); err != nil {
		t.Fatal(err)
	}
	w.Flush()
	if w.Error() != nil {
		t.Fatal(w.Error())
	}
}
