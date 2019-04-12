package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
)

const (
	nameSep                 = "_"
	fallbackOutfileName     = "ralad.out"
	maxNumSuffix        int = 100000
)

var logger *log.Logger

func debugf(format string, args ...interface{}) {
	if os.Getenv("RALAD_DEBUG") == "1" {
		logger.Output(2, "<DEBUG> "+fmt.Sprintf(format, args...))
	}
}

func redirectPolicy(req *http.Request, via []*http.Request) error {
	return nil
}

func nameIsSignificant(n string) bool {
	switch n {
	case "", "/", "index.html", "index.htm":
		return false
	}
	if _, err := os.Stat(n); err == nil {
		return false
	}
	return true
}

func makeFilename(req *http.Request) string {
	var name string
	cdp := req.Header.Get("Content-Disposition")
	if cdp != "" {
		debugf("found Content-Disposition header: %+v", cdp)
		fmt.Println(cdp)
		_, params, _ := mime.ParseMediaType(cdp)
		name = params["filename"]
		debugf("filename from cdp header: %s", name)
		if nameIsSignificant(name) {
			return name
		}
	}
	path := strings.Trim(req.URL.Path, "/")
	pathElems := strings.Split(path, "/")
	name = pathElems[len(pathElems)-1]
	if nameIsSignificant(name) {
		debugf("last path element is significant")
		return name
	}
	name = strings.Join(pathElems, nameSep)
	if nameIsSignificant(name) {
		debugf("full path is significant")
		return name
	}
	name = req.URL.Host + nameSep + name
	if nameIsSignificant(name) {
		debugf("host + full path is significant")
		return name
	}
	if nameIsSignificant(fallbackOutfileName) {
		debugf("fallback output file name is significant")
		return fallbackOutfileName
	}
	for i := 1; i < maxNumSuffix; i++ {
		name = fmt.Sprintf("%s.%d", fallbackOutfileName, i)
		if nameIsSignificant(name) {
			return name
		}
	}
	return ""
}

func ralad(downloadUrl string) error {

	client := &http.Client{
		CheckRedirect: redirectPolicy,
	}
	resp, err := client.Get(downloadUrl)
	if err != nil {
		return fmt.Errorf("error getting: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad http status: %s", resp.Status)
	}
	debugf("Response header: %+v\n", resp.Header)

	fn := makeFilename(resp.Request)
	if fn == "" {
		return fmt.Errorf("unable to generate filename")
	}
	debugf("output filename will be: %s", fn)
	outf, err := os.Create(fn)
	if err != nil {
		return fmt.Errorf("error creating file: %s", err)
	}
	defer outf.Close()
	written, err := io.Copy(outf, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing file: %s", err)
	}

	fmt.Printf("%d bytes written\n", written)
	return nil
}

func main() {
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	if len(os.Args) != 2 {
		fmt.Printf("no url given\n")
		os.Exit(1)
	}
	downloadUrl := os.Args[1]
	err := ralad(downloadUrl)
	if err != nil {
		fmt.Printf("ralad failed: %s\n", err)
		os.Exit(1)
	}
}
