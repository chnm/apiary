package apb

import "github.com/chnm/apiary/internal/httpx"

func Endpoints(baseURL string) []httpx.Endpoint {
	return []httpx.Endpoint{
		{Name: "APB: Featured verses", URL: baseURL + "/apb/index/featured"},
		{Name: "APB: Top verses", URL: baseURL + "/apb/index/top"},
		{Name: "APB: Verses in biblical order", URL: baseURL + "/apb/index/biblical"},
		{Name: "APB: Verses in chronological order of peak quotations", URL: baseURL + "/apb/index/peaks"},
		{Name: "APB: All verses in biblical order", URL: baseURL + "/apb/index/all"},
		{Name: "APB: Verse", URL: baseURL + "/apb/verse?ref=Luke+18:16"},
		{Name: "APB: Verse trend", URL: baseURL + "/apb/verse-trend?ref=Luke+18:16&corpus=chronam"},
		{Name: "APB: Verse quotations", URL: baseURL + "/apb/verse-quotations?ref=Luke+18:16"},
		{Name: "APB: Bible trend", URL: baseURL + "/apb/bible-trend"},
		{Name: "APB: Bible similarity", URL: baseURL + "/apb/bible-similarity"},
		{Name: "APB: Books of the Bible", URL: baseURL + "/apb/bible-books"},
	}
}
