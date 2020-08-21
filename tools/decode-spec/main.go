package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log"
	"os"
)

// This is equivalent to running
//  base64 --decode | gunzip
// to decode specs returned by the apg tool, with an additional effort to
// handle the hex-encoded inputs produced by encode-spec.
func main() {

	// decode the spec
	reader := bufio.NewReader(os.Stdin)
	s, _ := reader.ReadString('\n')
	decoded, err := hex.DecodeString(s)
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(s)
	}
	if err != nil {
		log.Fatal(err)
	}

	// gunzip the spec
	buf := bytes.NewBuffer(decoded)
	zr, err := gzip.NewReader(buf)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := io.Copy(os.Stdout, zr); err != nil {
		log.Fatal(err)
	}
	if err := zr.Close(); err != nil {
		log.Fatal(err)
	}
}
