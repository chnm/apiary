package main

import (
	"encoding/json"
	"net/http"
	"testing"

	apiary "github.com/chnm/apiary"
)

func TestAPBFeaturedVerses(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/index/featured", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.APBIndexItem
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if len(data) < 4 {
		t.Error("Not enough verses returned.")
	}

}

func TestAPBTopVerses(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/index/top", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.APBIndexItem
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if len(data) < 100 {
		t.Error("Not enough verses returned.")
	}

}

func TestAPBVersePeaks(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/index/peaks", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.APBIndexItem
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if len(data) < 100 {
		t.Error("Not enough verses returned.")
	}

}

func TestAPBVerse(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/verse?ref=Genesis+1:1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data apiary.Verse
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if data.Reference != "Genesis 1:1" {
		t.Error("Wrong verse returned.")
	}

}

func TestAPBVerseTrend(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/verse-trend?ref=Genesis+1:1&corpus=chronam", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data apiary.VerseTrendResponse
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if data.Reference != "Genesis 1:1" {
		t.Error("Wrong verse returned.")
	}

	if data.Corpus != "chronam" {
		t.Error("Wrong corpus returned.")
	}

	if len(data.Trend) < 50 {
		t.Error("Not enough data points returned.")
	}

}

func TestAPBBibleTrend(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/bible-trend", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data apiary.VerseTrendResponse
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if data.Reference != "bible" {
		t.Error("Wrong verse returned.")
	}

	if data.Corpus != "chronam" {
		t.Error("Wrong corpus returned.")
	}

	if len(data.Trend) < 50 {
		t.Error("Not enough data points returned.")
	}

}

func TestAPBVerseQuotations(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/verse-quotations?ref=Genesis+1:1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.VerseQuotation
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}

	if len(data) < 100 {
		t.Error("Not enough verses returned.")
	}

}

func TestBiblicalIndex(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/index/biblical", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.BibleBook
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestBibleAllIndex(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/index/all", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.APBIndexItemText
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestBibleSimilarity(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/bible-similarity", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.BibleSimilarityEdge
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}

func TestBibleBooks(t *testing.T) {
	// Check that we get the right response
	req, _ := http.NewRequest("GET", "/apb/bible-books", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Get the data
	var data []apiary.BibleBook
	err := json.Unmarshal(response.Body.Bytes(), &data)
	if err != nil {
		t.Error(err)
	}
}
