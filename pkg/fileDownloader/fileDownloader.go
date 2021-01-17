package fileDownloader

import (
	"io/ioutil"
	"net/http"
)

func DownloadFileByURI(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	fileData, err := ioutil.ReadAll(resp.Body)
	return fileData, err
}
