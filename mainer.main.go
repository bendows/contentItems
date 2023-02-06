package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log"
	"os"

	logger "github.com/bendows/gologger"
)

// func hashAndSend(r io.Reader) {
// 	w := sha256.New()
// 	//any reads from tee will read from r and write to w
// 	tee := io.TeeReader(r, w)
// 	sendReader(tee)
// 	sha := hex.EncodeToString(w.Sum(nil))
// 	fmt.Println(sha)
// }

//sendReader sends the contents of an io.Reader to stdout using a 256 byte buffer
// func sendReader(data io.Reader) {
// 	buff := make([]byte, 256)
// 	for {
// 		_, err := data.Read(buff)
// 		if err == io.EOF {
// 			break
// 		}
// 		fmt.Print(string(buff))
// 	}
// 	fmt.Println("")
// }

// bytes, _ := ioutil.ReadAll(r) //All bytes are now in memory
//	https://stackoverflow.com/questions/25671305/golang-io-copy-twice-on-the-request-body
//
//	this works for either a reader or writer,
//
// but if you use both in the same time the hash will be wrong.
type Hasher struct {
	io.Writer
	io.Reader
	hash.Hash
	Size uint64
}

func (h *Hasher) Write(p []byte) (n int, err error) {
	n, err = h.Writer.Write(p)
	logger.Loginfo.Printf("Hasher.Write [%d]\n", n)
	h.Hash.Write(p)
	h.Size += uint64(n)
	return
}

func (h *Hasher) Read(p []byte) (n int, err error) {
	n, err = h.Reader.Read(p)
	logger.Loginfo.Printf("Hasher.Read [%d]\n", n)
	h.Hash.Write(p[:n]) //on error n is gonna be 0 so this is still safe.
	return
}

func (h *Hasher) Sum() string {
	return hex.EncodeToString(h.Hash.Sum(nil))
}

type UploadHandle struct {
	Contents  io.Reader
	Contents2 io.Reader
}

func (h *UploadHandle) Read() (io.Reader, string, int64, error) {
	logger.Loginfo.Printf("Read():)\n")
	var b bytes.Buffer
	hashedReader := &Hasher{Reader: h.Contents, Hash: sha1.New()}
	n, err := io.Copy(&b, hashedReader)
	if err != nil {
		return nil, "", 0, err
	}
	return &b, hashedReader.Sum(), n, nil
}

// updated version based on @Dustin's comment since I complete
func (h *UploadHandle) ReadnewTee() (io.Reader, string, int64, error) {
	logger.Loginfo.Printf("ReadnewTee():)\n")
	var b bytes.Buffer
	hash := sha1.New()
	n, err := io.Copy(&b, io.TeeReader(h.Contents2, hash))
	if err != nil {
		return nil, "", 0, err
	}
	return &b, hex.EncodeToString(hash.Sum(nil)), n, nil
}

func main() {
	logger.LogOn = true
	h := UploadHandle{}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	// convert byte slice to io.Reader
	h.Contents = bytes.NewReader(b)
	_, cnt, d, err := h.Read()
	if err != nil {
		log.Fatal()
	}
	fmt.Printf("\nhash[%s]\nread-count[%d]\nerr[%s]\n", cnt, d, err)
	h.Contents2 = bytes.NewReader(b)
	// updated version based on @Dustin's comment since I complete
	_, cnt, d, err = h.ReadnewTee()
	if err != nil {
		log.Fatal()
	}
	fmt.Printf("\nhash[%s]\nread-count[%d]\nerr[%s]\n", cnt, d, err)
}
// https://golang.cafe/blog/golang-convert-byte-slice-to-io-reader.html
