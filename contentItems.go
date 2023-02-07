package ContentItems

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	logger "github.com/bendows/gologger"
)

type ContentItem struct {
	Hash         string
	Filename     string
	Contenttypes []string
}

func UploadFile(r io.Reader, filename, directory string) (i int, thefilename string) {
	diskFileName := ""
	var f *os.File
	var err error
	for {
		f, err = os.OpenFile(directory+"/"+filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err == nil {
			diskFileName = directory + "/" + filename
			logger.Loginfo.Printf("[%s] created\n", directory+"/"+filename)
			break
		}
		secondName := ""
		for i := 0; i < 100; i++ {
			secondName = filename + "_" + strconv.Itoa(i)
			f, err = os.OpenFile(directory+"/"+secondName, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
			if err == nil {
				filename = secondName
				diskFileName = directory + "/" + filename
				break
			}
		}
		break
	}
	if len(diskFileName) < 1 {
		return 0, ""
	}
	fileContents, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, ""
	}
	// filewriter := bufio.NewWriter(f)
	// _, err = filewriter.Write(fileContents)
	var n int
	n, err = f.Write(fileContents)
	if err != nil {
		return 0, ""
	}
	headers, err := Hashandcontents(bytes.NewReader(fileContents))
	logger.Loginfo.Printf("[%d] [%s] [%+v] file created\n", n, diskFileName, headers)
	return n, diskFileName
	// s, err := filewriter.Write(r)
	// if err != nil || s < 1 {
	// 	err = f.Close()
	// 	if err != nil {
	// 		return file, err
	// 	}
	// 	return file, err
	// }
	// err = filewriter.Flush()
	// if err != nil {
	// 	err = f.Close()
	// 	if err != nil {
	// 		return file, err
	// 	}
	// 	logger.Logerror.Printf("[%s] [%s] [%v]\n", key, name, err)
	// 	return file, err
	// }
	// logger.Loginfo.Printf("[%s] %d bytes saved\n", key, s)
	// fileInfo, err := f.Stat()
	// if err != nil {
	// 	logger.Logerror.Printf("[%s] [%s] [%v]\n", key, name, err)
	// 	return file, err
	// }
	// err = f.Close()
	// if err != nil {
	// 	logger.Logerror.Printf("[%s] [%s] [%v]\n", key, name, err)
	// 	return file, err
	// }
}

func Hashandcontents(r io.Reader) ([]string, error) {
	logger.Loginfo.Printf("Read():)\n")
	var b bytes.Buffer
	contentTypes := []string{}
	hashedReader := &Hasher{
		Reader: r,
		Hash:   sha1.New(),
	}
	n, err := io.Copy(&b, hashedReader)
	if err != nil {
		return contentTypes, err
	}
	logger.Loginfo.Printf("Read():) [%d]\n", n)
	meta := http.DetectContentType(b.Bytes())
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
	return contentTypes, nil
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
