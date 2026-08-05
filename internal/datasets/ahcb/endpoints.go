package ahcb

import "github.com/chnm/apiary/internal/httpx"

func Endpoints(baseURL string) []httpx.Endpoint {
	return []httpx.Endpoint{
		{Name: "Historial U.S. county boundaries by date from the Atlas of Historical County Boundaries", URL: baseURL + "/ahcb/counties/1844-05-08/"},
		{Name: "Historial U.S. county boundaries by date and county ID from the Atlas of Historical County Boundaries", URL: baseURL + "/ahcb/counties/1844-05-08/id/mas_essex,mas_middlesex/"},
		{Name: "Historial U.S. county boundaries by date and state/territory ID from the Atlas of Historical County Boundaries", URL: baseURL + "/ahcb/counties/1834-05-08/state-terr-id/nc_state,sc_state/"},
		{Name: "Historial U.S. county boundaries by date and state code from the Atlas of Historical County Boundaries", URL: baseURL + "/ahcb/counties/1844-05-08/state-code/nh,vt/"},
		{Name: "Historial U.S. state boundaries by date from the Atlas of Historical County Boundaries", URL: baseURL + "/ahcb/states/1820-05-10/"},
	}
}
