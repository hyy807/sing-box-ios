package v2rayxhttp

import (
	"bufio"
	"bytes"
	"io"
	"testing"
)

func TestEFANBodyRoundTrip(t *testing.T) {
	src := make([]byte, 256)
	for i := range src {
		src[i] = byte(i)
	}
	var b bytes.Buffer
	w := latin1Writer{w: &b}
	if _, err := w.Write(src); err != nil {
		t.Fatal(err)
	}
	r := latin1Reader{r: bufio.NewReaderSize(&b, 16)}
	dst, err := io.ReadAll(&r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(src, dst) {
		t.Fatal("latin1 body mismatch")
	}
}
