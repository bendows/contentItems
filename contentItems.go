package ContentItems

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	logger "github.com/bendows/gologger"
)

func SaveFile(filename, directory string, r io.Reader) (string, error) {
	fext := filepath.Ext(filename)
	fname := strings.TrimSuffix(filename, fext)
	diskFileName := ""
	//  os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	f, err := os.OpenFile(directory+"/"+fname+fext, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err == nil {
		diskFileName = directory + "/" + fname + fext
	} else {
		f.Close()
		secondName := ""
		for i := 0; i < 100; i++ {
			secondName = fname + "_" + strconv.Itoa(i)
			f, err = os.OpenFile(directory+"/"+secondName+fext, os.O_CREATE|os.O_EXCL, 0666)
			if err == nil {
				filename = secondName
				diskFileName = directory + "/" + secondName + fext
				logger.Loginfo.Printf("[%s] ", diskFileName)
				break
			}
		}
	}
	if len(diskFileName) < 1 {
		logger.Logerror.Printf("error [%v] ", err)
		return "", errors.New("empty pathname")
	}
	b, err := io.ReadAll(r)
	if err != nil {
		logger.Logerror.Printf("[%s] Error [%v]\n", diskFileName, err)
		return diskFileName, err
	}
	_, err = f.Write(b)
	f.Close()
	if err != nil {
		logger.Logerror.Printf("[%s] Error [%v]\n", diskFileName, err)
	}
	return diskFileName, nil
}

func GenerateHash(r io.Reader) (string, error) {
	hash := md5.New() // fast & good enough
	if _, err := io.Copy(hash, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func GenerateHashAndFileTypes(r io.Reader) []string {
	signature, err := ioutil.ReadAll(io.LimitReader(r, 512))
	if err != nil {
		return []string{err.Error()}
	}
	hash := md5.New() // fast & good enough
	if _, err := io.Copy(hash, bytes.NewReader(signature)); err != nil {
		return []string{err.Error()}
	}
	if _, err := io.Copy(hash, r); err != nil {
		return []string{err.Error()}
	}
	contentTypes := []string{}
	contentTypes = append(contentTypes, fmt.Sprintf("%x", hash.Sum(nil)))
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
	return contentTypes
}
