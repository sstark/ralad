package main

import (
	"net/url"
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
