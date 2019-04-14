package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type EllTest struct {
	in  *url.URL
	out []string // output for truncate/part/full
}

var elltests = []EllTest{
	{
		&url.URL{
			Scheme: "http",
			Host:   "www.example.com",
			Path:   "/file one&two",
		},
		[]string{
			"http://www.example.com/file%20one&two",
			"http://www.example.com...",
			"http://www.example.com/file%20one&two",
		},
	},
	{
		&url.URL{
			Scheme: "https",
			Host:   "www1.bla.blu.example.com",
			Path:   "/feawf/gser/gawef/agae/rfaw/efaw/eg/aerfa/w/fawg/awef/awet/t/4t3/a/wgw34/43t/t34/t34g/aw4f/f4.zip",
		},
		[]string{
			"https://www1.bla.blu.example.com/feawf/gser/gawef/agae/rfaw/efaw/eg/aer...",
			"https://www1.bla.blu.example.com...",
			"https://www1.bla.blu.example.com/feawf/gser/gawef/agae/rfaw/efaw/eg/aerfa/w/fawg/awef/awet/t/4t3/a/wgw34/43t/t34/t34g/aw4f/f4.zip",
		},
	},
}

func TestEllipsize(t *testing.T) {
	var got, want string
	var modes []string = []string{"truncate", "part", "full"}
	for _, et := range elltests {
		for i, mode := range modes {
			frDisplay = mode
			got = ellipsize(et.in)
			want = et.out[i]
			if got != want {
				t.Errorf("got %s, but wanted %s", got, want)
			}
		}
	}
}

type MakeFNTest struct {
	in   *http.Response
	out  string
	warn string
}

var makefntests = []MakeFNTest{
	{
		&http.Response{
			Header: http.Header{
				"Content-Disposition": {"attachment; filename=something.zip"},
			},
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.com",
					Path:   "/34g/aw4f/somethingNot.zip",
				},
			},
		},
		"something.zip",
		"",
	},
	{
		&http.Response{
			Header: http.Header{
				"Content-Disposition": {"attachment; filename=this/is/bogus"},
			},
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.com",
					Path:   "/34g/aw4f/bogus.zip",
				},
			},
		},
		"bogus.zip",
		"failed to parse Content-Disposition header: mime: invalid media parameter",
	},
	{
		&http.Response{
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.com",
					Path:   "/34g/aw4f/f4.tgz",
				},
			},
		},
		"f4.tgz",
		"",
	},
	{
		&http.Response{
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.com",
					Path:   "/34g/aw4f/index.html",
				},
			},
		},
		"34g_aw4f_index.html",
		"",
	},
	{
		&http.Response{
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "https",
					Host:   "www.example.com",
					Path:   "/index.html",
				},
			},
		},
		"www.example.com_index.html",
		"",
	},
	{
		&http.Response{
			Request: &http.Request{
				URL: &url.URL{
					Scheme: "https",
					Host:   "evil.example.com",
					Path:   "/files/%2fetc%2fpasswd",
				},
			},
		},
		"%2fetc%2fpasswd",
		"",
	},
}

func TestMakeFilename(t *testing.T) {
	var got, wanted string
	buf := new(bytes.Buffer)
	userWarnStream = buf
	for _, mt := range makefntests {
		buf.Reset()
		got = makeFilename(mt.in)
		wanted = mt.out
		if got != wanted {
			t.Errorf("got %s, but wanted %s", got, wanted)
		}
		got = buf.String()
		wanted = mt.warn
		if got != wanted {
			t.Errorf("user warning(s) should be\n\t\"%s\"\nbut is\n\t\"%s\"", got, wanted)
		}
	}
}

type AskOkTest struct {
	in  string
	out bool
}

var askOkTests = []AskOkTest{
	{"y\n", true},
	{"n\n", false},
	{"yes\n", true},
	{"safd\n", false},
	{"\n", false},
}

func TestAskOk(t *testing.T) {
	var got, want bool
	for _, aot := range askOkTests {
		userInputStream = bufio.NewReader(strings.NewReader(aot.in))
		got = askOk("")
		want = aot.out
		if got != want {
			t.Errorf("got %t, but wanted %t", got, want)
		}
	}
}

type RedirPolInput struct {
	req *http.Request
	via []*http.Request
}

type RedirMode struct {
	mode       string
	userInputs []string
	outs       []error
	uiErr      []error
}

type RedirPolTest struct {
	in    RedirPolInput
	modes []RedirMode
}

var redirpoltests = []RedirPolTest{
	{
		in: RedirPolInput{
			req: &http.Request{
				// the request we are redirected to
				Method: "GET",
				URL: &url.URL{
					Scheme: "http",
					Host:   "smeik",
					Path:   "/r/bla.zip",
				},
				// the response that lead to the redirection
				Response: &http.Response{
					StatusCode: 301,
					Status:     "301 Moved Permanently",
					Header: http.Header{
						"Location": []string{"http://smeik/r/bla.zip"},
					},
				},
			},
			via: []*http.Request{
				// The earlier request that lead to the redirection
				&http.Request{
					Method: "GET",
					URL: &url.URL{
						Scheme: "http",
						Host:   "smeik",
						Path:   "/r2/bla.zip",
					},
				},
			},
		},
		modes: []RedirMode{
			{
				mode:       "always",
				userInputs: []string{"y\n", "n\n"},
				outs:       []error{nil, http.ErrUseLastResponse},
				uiErr:      []error{io.EOF, io.EOF},
			},
			{
				mode:       "relaxed",
				userInputs: []string{"\n", "\n"},
				outs:       []error{nil, nil},
				uiErr:      []error{nil, nil},
			},
			{
				mode:       "never",
				userInputs: []string{"\n", "\n"},
				outs:       []error{nil, nil},
				uiErr:      []error{nil, nil},
			},
		},
	},
	{
		in: RedirPolInput{
			req: &http.Request{
				// the request we are redirected to
				Method: "GET",
				URL: &url.URL{
					Scheme: "http",
					Host:   "cdn.example.com",
					Path:   "/g2342353/bla.zip",
				},
				// the response that lead to the redirection
				Response: &http.Response{
					StatusCode: 307,
					Status:     "307 Temporary Redirect",
					Header: http.Header{
						"Location": []string{"http://cdn.example.com/g2342353/bla.zip"},
					},
				},
			},
			via: []*http.Request{
				// The earlier request that lead to the redirection
				&http.Request{
					Method: "GET",
					URL: &url.URL{
						Scheme: "http",
						Host:   "dl.example.com",
						Path:   "/bla.zip",
					},
				},
			},
		},
		modes: []RedirMode{
			{
				mode:       "always",
				userInputs: []string{"y\n", "n\n"},
				outs:       []error{nil, http.ErrUseLastResponse},
				uiErr:      []error{io.EOF, io.EOF},
			},
			{
				mode:       "relaxed",
				userInputs: []string{"y\n", "n\n"},
				outs:       []error{nil, http.ErrUseLastResponse},
				uiErr:      []error{io.EOF, io.EOF},
			},
			{
				mode:       "never",
				userInputs: []string{"\n", "\n"},
				outs:       []error{nil, nil},
				uiErr:      []error{nil, nil},
			},
		},
	},
}

func TestRedirPolicy(t *testing.T) {
	var got, want error
	userPromptStream = ioutil.Discard
	userWarnStream = ioutil.Discard
	for _, rpt := range redirpoltests {
		t.Log(rpt.in.req)
		t.Log(rpt.in.req.URL)
		t.Log(rpt.in.req.Response.StatusCode)
		for _, mode := range rpt.modes {
			fredirPolicy = mode.mode
			t.Logf("mode: %s", mode.mode)
			for i, ui := range mode.userInputs {
				userInputStream = bufio.NewReader(strings.NewReader(ui))
				want = mode.outs[i]
				got = redirectPolicy(rpt.in.req, rpt.in.via)
				if want != got {
					t.Errorf("wanted %v, got %v", want, got)
				}
				// here we read the userInputStream again, like askOk would. If
				// we get io.EOF, we know that the input was consumed earlier
				// and the user was interactively asked to confirm the redirect
				_, uiErr := userInputStream.ReadString('\n')
				if uiErr != mode.uiErr[i] {
					if uiErr == io.EOF {
						t.Errorf("user was prompted, but should not have been")
					} else {
						t.Error("user was not prompted, but should have been")
					}
				}
			}
		}
	}
}
