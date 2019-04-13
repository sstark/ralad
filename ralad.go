package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	nameSep                 = "_"
	fallbackOutfileName     = "ralad.out"
	maxNumSuffix        int = 100000
	maxRedirStrlen      int = 72
)

var (
	logger       *log.Logger
	funsafeTLS   bool
	fredirPolicy string
	frDisplay    string
	foutfileName string
)

func debugf(format string, args ...interface{}) {
	if os.Getenv("RALAD_DEBUG") == "1" {
		logger.Output(2, "<DEBUG> "+fmt.Sprintf(format, args...))
	}
}

func askOk(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	debugf(text)
	switch strings.TrimSpace(text) {
	case "yes", "y":
		debugf("ok!")
		return true
	}
	return false
}

func ellipsize(u *url.URL) string {
	switch frDisplay {
	case "truncate":
		s := u.String()
		if len(s) > maxRedirStrlen {
			return fmt.Sprintf("%s...", s[:maxRedirStrlen-1])
		} else {
			return s
		}
	case "part":
		return fmt.Sprintf("%s://%s...", u.Scheme, u.Host)
	case "full":
		return u.String()
	}
	return u.String()
}

func redirectPolicy(req *http.Request, via []*http.Request) error {
	debugf("redirect: %+v", req)
	for _, v := range via {
		debugf("via: %+v", v)
	}
	if fredirPolicy == "never" {
		return nil
	}
	lastScheme := via[len(via)-1].URL.Scheme
	lastHost := via[len(via)-1].URL.Host
	if fredirPolicy == "always" || req.URL.Scheme != lastScheme || req.URL.Host != lastHost {
		ans := askOk(fmt.Sprintf("redirect to %s? (y/n) ", ellipsize(req.URL)))
		if ans == true {
			debugf("allow redirect")
			return nil
		} else {
			debugf("deny redirect")
			return http.ErrUseLastResponse
		}
	}
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

func makeFilename(resp *http.Response) string {
	var name string
	cdp := resp.Header.Get("Content-Disposition")
	if cdp != "" {
		debugf("found Content-Disposition header: %+v", cdp)
		_, params, _ := mime.ParseMediaType(cdp)
		name = strings.Trim(params["filename"], "/")
		debugf("filename from cdp header: %s", name)
		if nameIsSignificant(name) {
			return name
		}
	}
	path := strings.Trim(resp.Request.URL.Path, "/")
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
	name = resp.Request.URL.Host + nameSep + name
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

func downloadBody(resp *http.Response, outf io.Writer) (int64, error) {
	cl := resp.ContentLength
	debugf("Content-Length: %d", cl)
	if cl == -1 {
		cl = 0
	}
	bar := pb.New64(cl).SetUnits(pb.U_BYTES)
	bar.ShowSpeed = true
	bar.Format("▰▰▰▱▰")
	bar.Start()
	rd := bar.NewProxyReader(resp.Body)
	written, err := io.Copy(outf, rd)
	if err != nil {
		return written, fmt.Errorf("error writing file: %s", err)
	}
	bar.Finish()
	if resp.ContentLength != -1 && resp.ContentLength != written {
		fmt.Printf("warning: bytes written (%d) is different from Content-Length header (%d)\n", written, resp.ContentLength)
	}
	return written, nil
}

func ralad(downloadUrl string) error {
	client := &http.Client{
		CheckRedirect: redirectPolicy,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: funsafeTLS,
			},
		},
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

	var fn string
	if foutfileName == "" {
		fn = makeFilename(resp)
		if fn == "" {
			return fmt.Errorf("unable to generate filename")
		}
	} else {
		fn = foutfileName
	}
	debugf("output filename will be: %s", fn)
	var outf *os.File
	if fn == "-" {
		outf = os.Stdout
	} else {
		outf, err = os.Create(fn)
		if err != nil {
			return fmt.Errorf("error creating file: %s", err)
		}
		defer outf.Close()
	}
	written, err := downloadBody(resp, outf)
	if err == nil {
		fmt.Printf("%d bytes written to %s\n", written, outf.Name())
	}
	return err
}

func Usage() {
	fmt.Printf("Usage:\n")
	fmt.Printf("\tralad [flags] url\n")
	fmt.Printf("Flags:\n")
	flag.PrintDefaults()
}

func validateFlags() error {
	switch fredirPolicy {
	case "always", "relaxed", "never":
		debugf(fredirPolicy)
	default:
		return fmt.Errorf("invalid value for rpolicy: %s", fredirPolicy)
	}
	switch frDisplay {
	case "full", "part", "truncate":
		debugf(frDisplay)
	default:
		return fmt.Errorf("invalid value for rdisplay: %s", frDisplay)
	}
	return nil
}

func main() {
	flag.BoolVar(&funsafeTLS, "unsafeTLS", false, "ignore TLS certificate errors")
	flag.StringVar(&fredirPolicy, "rpolicy", "relaxed", "set redirect confirmation policy: always|relaxed|never")
	flag.StringVar(&frDisplay, "rdisplay", "truncate", "redirect display: full|part|truncate")
	flag.StringVar(&foutfileName, "o", "", "output file name (use - for stdout)")
	flag.Usage = Usage
	flag.Parse()
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	err := validateFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(flag.Args()) != 1 {
		fmt.Printf("no url given\n")
		os.Exit(1)
	}
	downloadUrl := flag.Args()[0]
	err = ralad(downloadUrl)
	if err != nil {
		fmt.Printf("ralad failed: %s\n", err)
		os.Exit(1)
	}
}
