/*
Copyright © 2021 willbenica

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package internal

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestIsValidInput(t *testing.T) {
	cases := []struct {
		endpoint string
		isValid  bool
	}{
		{"leads", true},
		{"visits", true},
	}

	for _, test := range cases {
		t.Run(fmt.Sprintf("%s is a valid leadfeeder endpoint, %v", test.endpoint, test.isValid), func(t *testing.T) {
			got := IsValidEndpoint(test.endpoint)
			if got != test.isValid {
				t.Errorf("got %v, wanted %v", got, test.isValid)
			}
		})
	}

}

const (
	// URL is the URL for an lf instance
	URL = "https://api.leadfeeder.me/accounts"
	// TOKEN is the auth token generated in the lf UI
	TOKEN       string = "Bearer xxxxYYYYxxxxWWWWxxxxQQQ867512"
	ENDPOINT    string = "leads"
	ACCOUNT_ID  string = "123456"
	FILE        string = "/Users/wbenica/projects/lf-cli/test_with_data/test.json"
	PAGE_SIZE   int    = 100
	PAGE_NUMBER int    = 1
	// The following requires a custom JSON marshal implementation
	MOCK_RESPONSE        = `{"data":[{"id":"myLeadId","type":"leads","attributes":{"facebook_url":"","status":"new","twitter_handle":"","first_visit_date":"2021-05-23","last_visit_date":"2021-05-25","linkedin_url":"","name":"myCompany","website_url":"","business_id":"","revenue":"","assignee":"","emailed_to":"","view_in_leadfeeder":"https://me.leadfeeder.com/link/lf_id","industry":"N/A","phone":"","crm_lead_id":"","crm_organization_id":"","employee_count":0,"tags":[],"logo_url":"","visits":1,"quality":1},"relationships":{"location":{"data":{"id":"myLocationId","type":"locations"}}}}],"included":[{"id":"myLocationId","type":"locations","attributes":{"country":"Germany","country_code":"DE","region":"Saxony","region_code":"","city":"Dresden","state_code":""}}],"links":{"self":"https://me.leadfeeder.com/accounts/123456/leads?end_date=2021-05-24&page%5Bnumber%5D=1&page%5Bsize%5D=100&start_date=2021-01-01","next":"https://me.leadfeeder.com/accounts/123456/leads?end_date=2021-05-24&page%5Bnumber%5D=2&page%5Bsize%5D=100&start_date=2021-01-01","last":"https://me.leadfeeder.com/accounts/123456/leads?end_date=2021-05-24&page%5Bnumber%5D=45&page%5Bsize%5D=100&start_date=2021-01-01"}}`
	TEST_FOLDER   string = "./test_files/"
	// Test files
	L1 = "leads_test_1.json"
	L2 = "leads_test_2.json"
	V1 = "visits_test_1.json"
	V2 = "visits_test_2.json"
)

func TestGetEndPointData(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	expected := MOCK_RESPONSE
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	t.Run("GetEndPointData receives data", func(t *testing.T) {
		httpmock.RegisterResponder("GET", URL+"/"+ACCOUNT_ID+"/"+ENDPOINT,
			httpmock.NewStringResponder(200, expected))
		epData, err := GetEndPointData("leads", URL, TOKEN, ACCOUNT_ID, today, today, PAGE_SIZE, PAGE_NUMBER)
		if err != nil {
			t.Errorf("Error while retrieving an EndPoint:\n%s", err)
		}
		got, err := epData.String()
		if err != nil {
			t.Errorf("Error while extracting the desired data from the EndPoint:\n%s", err)
		}
		// Addinga a carriage return to the end of the expected block - solves the encoding/decoding that is happening
		if got != expected+"\n" {
			t.Errorf("   got: %q \nwanted: %q", got, expected)
		}
	})
}

func TestEndpointURLBuilder(t *testing.T) {
	want := fmt.Sprintf("%s/%s/%s?start_date=2002-10-15&end_date=2005-09-25&page%%5Bsize%%5D=%d&page%%5Bnumber%%5D=%d",
		URL, ACCOUNT_ID, ENDPOINT, PAGE_SIZE, PAGE_NUMBER)

	cases := []struct {
		name        string
		providedURL string
		endpoint    string
		accountID   string
		start       string
		end         string
		expected    string
	}{
		{name: "shortest valid URL provided", providedURL: "api.leadfeeder.me", endpoint: "leads", accountID: "123456", start: "2002-10-15", end: "2005-09-25", expected: want},
		{name: "valid URL with https", providedURL: "https://api.leadfeeder.me", endpoint: "leads", accountID: "123456", start: "2002-10-15", end: "2005-09-25", expected: want},
		{name: "valid URL with https and api", providedURL: "https://api.leadfeeder.me/", endpoint: "leads", accountID: "123456", start: "2002-10-15", end: "2005-09-25", expected: want},
		{name: "valid URL with trailing slash", providedURL: "api.leadfeeder.me/", endpoint: "leads", accountID: "123456", start: "2002-10-15", end: "2005-09-25", expected: want},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := EndpointURLBuilder(c.providedURL, c.endpoint, c.accountID, c.start, c.end, PAGE_SIZE, PAGE_NUMBER)
			if err != nil {
				if err.Error() != "why u no use 'https'?" {
					t.Errorf("got an unexpected error: %q", err)
				}
			}
			if got != c.expected {
				t.Errorf("got %q, wanted %q", got, c.expected)
			}
		})
	}
}

func TestTodayOrDate(t *testing.T) {

	cases := []struct {
		name         string
		providedDate string
		expected     string
	}{
		{name: "today is understood", providedDate: "tODaY", expected: time.Now().Format("2006-01-02")},
		{name: "valid date in proper format", providedDate: "2005-09-25", expected: "2005-09-25"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := TodayOrDate(c.providedDate)
			if got != c.expected {
				t.Errorf("got %q, wanted %q", got, c.expected)
			}
		})
	}
}

func TestParseApiResponseToLeadsStruct(t *testing.T) {
	// Setup
	leads_one, err := os.Open(TEST_FOLDER + L1)
	if err != nil {
		fmt.Println(err)
	}
	reader := bufio.NewReader(leads_one)
	file_content, _ := ioutil.ReadAll(reader)
	leads_data := ParseApiResponseToLeadsResponseStruct(file_content)

	// Let's test
	want := "https://me.leadfeeder.com/accounts/123456/leads?end_date=2021-05-24&page%5Bnumber%5D=1&page%5Bsize%5D=2&start_date=2021-01-01"
	if leads_data.Links.Self != want {
		t.Errorf("Got %s,\nWanted %s", leads_data.Links.Self, want)
	}
	// Are there 100 different ids in the response
	expected_number_of_leads := 2
	var l_ids []string
	for _, l := range leads_data.Data {
		l_ids = append(l_ids, l.ID)
	}
	if len(l_ids) != expected_number_of_leads {
		t.Errorf("Want %d, Got %d", expected_number_of_leads, len(l_ids))
	}
}

func TestParseApiResponseToVisitStruct(t *testing.T) {
	// Setup
	visits_one, err := os.Open(TEST_FOLDER + V1)
	if err != nil {
		fmt.Println(err)
	}
	reader := bufio.NewReader(visits_one)
	file_content, _ := ioutil.ReadAll(reader)
	visits_data := ParseApiResponseToVisitsResponseStruct(file_content)

	// Let's test
	want := "https://me.leadfeeder.com/accounts/123456/visits?end_date=2021-05-25&page%5Bnumber%5D=1&page%5Bsize%5D=2&start_date=2021-01-01"
	if visits_data.Links.Self != want {
		t.Errorf("Got %s,\nWanted %s", visits_data.Links.Self, want)
	}
	// Are there 100 different ids in the response
	expected_number_of_visits := 2
	var v_ids []string
	for _, v := range visits_data.Data {
		v_ids = append(v_ids, v.ID)
	}
	if len(v_ids) != expected_number_of_visits {
		t.Errorf("Want %d, Got %d", expected_number_of_visits, len(v_ids))
	}
}
