package ContentItems

import (
	"bytes"
	"crypto/md5"
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

func SaveFile(r io.Reader, directory string, filename string) (size int, path string, err error) {
	fext := filepath.Ext(filename)
	fname := strings.TrimSuffix(filename, fext)
	diskFileName := ""
	f, err := os.OpenFile(directory+"/"+fname+fext, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err == nil {
		diskFileName = directory + "/" + fname + fext
		b, err := ioutil.ReadAll(r)
		if err != nil {
			f.Close()
			return 0, diskFileName, err
		}
		fsize, err := f.Write(b)
		f.Close()
		if err == nil {
			return fsize, diskFileName, nil
		}
		return fsize, diskFileName, err
	}
	secondName := ""
	for i := 0; i < 100; i++ {
		secondName = fname + "_" + strconv.Itoa(i)
		f, err = os.OpenFile(directory+"/"+secondName+fext, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			continue
		}
		filename = secondName
		diskFileName = directory + "/" + secondName + fext
		b, err := ioutil.ReadAll(r)
		if err != nil {
			f.Close()
			return 0, diskFileName, err
		}
		fsize, err := f.Write(b)
		f.Close()
		if err == nil {
			return fsize, diskFileName, nil
		}
		return fsize, diskFileName, err
	}
	return 0, diskFileName, err
}

// func SaveFile(filename, directory string, r io.Reader) (string, string, error) {
// 	fext := filepath.Ext(filename)
// 	fname := strings.TrimSuffix(filename, fext)
// 	diskFileName := ""
// 	f, err := os.OpenFile(directory+"/"+fname+fext, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
// 	if err == nil {
// 		diskFileName = directory + "/" + fname + fext
// 	} else {
// 		secondName := ""
// 		for i := 0; i < 100; i++ {
// 			secondName = fname + "_" + strconv.Itoa(i)
// 			f, err = os.OpenFile(directory+"/"+secondName+fext, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
// 			if err == nil {
// 				filename = secondName
// 				diskFileName = directory + "/" + secondName + fext
// 				logger.Loginfo.Printf("[%s] ", diskFileName)
// 				break
// 			}
// 		}
// 	}
// 	if len(diskFileName) < 1 {
// 		logger.Logerror.Printf("error [%v] ", err)
// 		return "", "", errors.New("empty pathname")
// 	}
// 	var b bytes.Buffer
// 	hash := md5.New()
// 	_, err = io.Copy(&b, io.TeeReader(r, hash))
// 	if err != nil {
// 		f.Close()
// 		return diskFileName, "", err
// 	}
// 	_, err = f.Write(b.Bytes())
// 	if err != nil {
// 		logger.Logerror.Printf("[%s] Error [%v]\n", diskFileName, err)
// 	}
// 	f.Close()
// 	return diskFileName, hex.EncodeToString(hash.Sum(nil)), err
// }

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
