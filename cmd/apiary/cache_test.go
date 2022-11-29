package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var response *httptest.ResponseRecorder

func benchmarkEndpoint(path string, cache bool) {
	// Add nocache to the URL
	base, err := url.Parse(path)
	if err != nil {
		return
	}
	if !cache {
		params := url.Values{}
		params.Add("nocache", "")
		base.RawQuery = params.Encode()
	}
	path = base.String()

	// Run the request
	req, _ := http.NewRequest("GET", path, nil)
	response = executeRequest(req)
}

func BenchmarkCachedCounties(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkEndpoint("/ahcb/counties/1874-05-08/", true)
	}
}

func BenchmarkNotCachedCounties(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkEndpoint("/ahcb/counties/1874-05-08/", false)
	}
}

func BenchmarkCachedNorthAmerica(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkEndpoint("/ne/northamerica/", true)
	}
}

func BenchmarkNotCachedNorthAmerica(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkEndpoint("/ne/northamerica/", false)
	}
}

func BenchmarkCachedVerseTrend(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkEndpoint("/apb/verse-trend?ref=Luke+18:16&corpus=chronam", true)
	}
}

func BenchmarkNotCachedVerseTrend(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkEndpoint("/apb/verse-trend?ref=Luke+18:16&corpus=chronam", false)
	}
}
