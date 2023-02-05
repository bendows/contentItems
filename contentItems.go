package ContentItems

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	logger "github.com/bendows/gologger"
)

type ContentItem struct {
	Hash         string
	Filename     string
	Contenttypes []string
}

func GenerateFileTypesAndHash(r io.Reader) ContentItem {
	signature, err := ioutil.ReadAll(io.LimitReader(r, 512))
	if err != nil {
		return ContentItem{
			Hash:         "",
			Contenttypes: []string{err.Error()},
		}
	}
	hash := md5.New() // fast & good enough
	if _, err := io.Copy(hash, bytes.NewReader(signature)); err != nil {
		return ContentItem{
			Hash:         "",
			Contenttypes: []string{err.Error()},
		}
	}
	if _, err := io.Copy(hash, r); err != nil {
		return ContentItem{
			Hash:         "",
			Contenttypes: []string{err.Error()},
		}
	}
	contentTypes := []string{}
	meta := http.DetectContentType(signature)
	logger.Loginfo.Printf("meta %s\n", meta)
	contentTypes = append(contentTypes, meta)
	sub1 := strings.Split(meta, " ")
	for _, key := range sub1 {
		key = strings.TrimRight(key, ";")
		if len(sub1) > 1 {
			contentTypes = append(contentTypes, key)
		}
		sub2 := strings.Split(key, "/")
		if len(sub2) > 1 {
			contentTypes = append(contentTypes, sub2...)
		}
		sub3 := strings.Split(key, "=")
		if len(sub3) > 1 {
			contentTypes = append(contentTypes, sub3...)
		}
	}
	return ContentItem{
		Hash:         fmt.Sprintf("%x", hash.Sum(nil)),
		Contenttypes: contentTypes,
	}
}
