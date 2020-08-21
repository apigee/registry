package main

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

// This is equivalent to running
//  gzip --stdout | xxd -p | tr -d '\n'
// but with a platform-independent gzip encoding to provide more stable results.
func main() {

	// gzip the spec
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)

	if _, err := io.Copy(zw, os.Stdin); err != nil {
		log.Fatal(err)
	}

	if err := zw.Close(); err != nil {
		log.Fatal(err)
	}

	// hex-encode the spec
	s := hex.EncodeToString(buf.Bytes())

	fmt.Printf("%s", s)
}
