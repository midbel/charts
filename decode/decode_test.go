package decode

import (
	"os"
	"testing"
)

func TestDecoder_Decode(t *testing.T) {
	r, err := os.Open("testdata/sample.chart")
	if err != nil {

	}
	defer r.Close()

	_, err = NewDecoder(r).Decode()
	if err != nil {
		t.Fatal(err)
	}
}
