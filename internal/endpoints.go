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
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// EndPoint interface contains methods that help to work with the various endpoints
type EndPoint interface {
	Type() string
	GetData() ([]LeadData, []Location, []VisitData, Links)
	GetLastPageNumber() (int, error)
	String() (string, error)
}

type EndPointData interface {
	GetAllData()
}

//LeadsResponse is a struct that used for mapping lead response JSONs to a native struct for (un)marshalling
type LeadsResponse struct {
	Data     []LeadData `json:"data"`
	Included []Location `json:"included"`
	Links    Links      `json:"links"`
}

func (lr LeadsResponse) Type() string {
	return "LeadsResponse"
}

func (lr LeadsResponse) GetData() ([]LeadData, []Location, []VisitData, Links) {
	return lr.Data, lr.Included, nil, lr.Links
}

func (lr LeadsResponse) GetLastPageNumber() (int, error) {

	lastUrl := lr.Links.Last
	logger.Debug("Getting last page from Links.last", zap.String("Links.Last", lastUrl))
	if len(lastUrl) > 0 {
		lastUrlIdxStart := strings.Index(lastUrl, "page%5Bnumber%5D=") + 17
		lastUrlIdxEnd := strings.Index(lastUrl, "&page%5Bsize")
		n, err := strconv.Atoi(lastUrl[lastUrlIdxStart:lastUrlIdxEnd])
		if err != nil {
			logger.Error("Reading from Links.Last failed", zap.Error(err))
			return 0, err
		}
		logger.Debug("Last page determined", zap.Int("last page", n))
		return n, nil
	}
	logger.Debug("Links.Last is nil", zap.String("Links.Last", lr.Links.Last))
	return 0, nil
}

func (lr LeadsResponse) String() (string, error) {
	logger.Debug("Converting LeadsResponse to String")
	var buf bytes.Buffer
	e := json.NewEncoder(&buf)
	e.SetEscapeHTML(false)
	e.Encode(lr)
	return buf.String(), nil
}

func ParseApiResponseToLeadsResponseStruct(data []byte) (lr LeadsResponse) {
	logger.Debug("Parsing API response to LeadsResponse")
	err := json.Unmarshal(data, &lr)
	if err != nil {
		logger.Error("Unmarshalling data to string has failed", zap.Error(err))
	}
	return lr
}

type LeadData struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Attributes    LeadAttributes `json:"attributes"`
	Relationships Relationships  `json:"relationships"`
}

type Leads struct {
	Data []LeadData
}

func (l Leads) GetAllData() string {
	logger.Debug("Creating []byte to print VisitData")
	var buffer bytes.Buffer
	e := json.NewEncoder(&buffer)
	e.SetEscapeHTML(false)
	for _, lead := range l.Data {
		e.Encode(lead)
	}
	return buffer.String()
}

func LoopThroughLeadsData(d []LeadData, l []Location, start int, end int, f Flags) ([]LeadData, []Location, error) {
	logger.Debug("Looping through LeadsData")
	for i := start; i <= end; i++ {
		logger.Info("Starting loops", zap.Int("current", i), zap.Int("end", end))
		ep_data, err := GetEndPointData("leads", f.BaseURL, f.Token, f.AccountID, TodayOrDate(f.StartDate), TodayOrDate(f.EndDate), f.PageSize, i)
		if err != nil {
			return nil, nil, err
		}
		leads, locations, _, _ := ep_data.GetData()

		d = append(d, leads...)
		l = append(l, locations...)
	}
	return d, l, nil
}

type LeadAttributes struct {
	FacebookURL       string   `json:"facebook_url"`
	Status            string   `json:"status"`
	TwitterHandle     string   `json:"twitter_handle"`
	FirstVisitDate    string   `json:"first_visit_date"`
	LastVisitDate     string   `json:"last_visit_date"`
	LinkedinURL       string   `json:"linkedin_url"`
	Name              string   `json:"name"`
	WebsiteURL        string   `json:"website_url"`
	BusinessID        string   `json:"business_id"`
	Revenue           string   `json:"revenue"`
	Assignee          string   `json:"assignee"`
	EmailedTo         string   `json:"emailed_to"`
	ViewInLeadfeeder  string   `json:"view_in_leadfeeder"`
	Industry          string   `json:"industry"`
	Phone             string   `json:"phone"`
	CrmLeadID         string   `json:"crm_lead_id"`
	CrmOrganizationID string   `json:"crm_organization_id"`
	EmployeeCount     int      `json:"employee_count"`
	Tags              []string `json:"tags"`
	LogoURL           string   `json:"logo_url"`
	Visits            int      `json:"visits"`
	Quality           int      `json:"quality"`
}

type Relationships struct {
	Location RelatedLocaction `json:"location"`
}

type RelatedLocaction struct {
	Data LocData `json:"data"`
}

type LocData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ------------------------------------

type Location struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Attributes LocationAttributes `json:"attributes"`
}

type Locations struct {
	Data []Location
}

func (l Locations) GetAllData() string {
	logger.Debug("Creating []byte to print Locations")
	var buffer bytes.Buffer
	e := json.NewEncoder(&buffer)
	e.SetEscapeHTML(false)
	for _, location := range l.Data {
		e.Encode(location)
	}
	return buffer.String()
}

type LocationAttributes struct {
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Region      string `json:"region"`
	RegionCode  string `json:"region_code"`
	City        string `json:"city"`
	StateCode   string `json:"state_code"`
}

// -------------------------------------

type VisitsResponse struct {
	Data  []VisitData `json:"data"`
	Links Links       `json:"links"`
}

func (vr VisitsResponse) GetData() ([]LeadData, []Location, []VisitData, Links) {
	return nil, nil, vr.Data, vr.Links
}

func (vr VisitsResponse) Type() string {
	return "VisitsResponse"
}

func (vr VisitsResponse) GetLastPageNumber() (int, error) {
	lastUrl := vr.Links.Last
	logger.Debug("Getting last page from Links.last", zap.String("Links.Last", lastUrl))
	if len(lastUrl) > 0 {
		lastUrlIdxStart := strings.Index(lastUrl, "page%5Bnumber%5D=") + 17
		lastUrlIdxEnd := strings.Index(lastUrl, "&page%5Bsize")
		n, err := strconv.Atoi(lastUrl[lastUrlIdxStart:lastUrlIdxEnd])
		if err != nil {
			logger.Error("Reading from Links.Last failed", zap.Error(err))
			return 0, err
		}
		logger.Debug("Last page determined", zap.Int("last page", n))
		return n, nil
	}
	logger.Debug("Links.Last is nil", zap.String("Links.Last", vr.Links.Last))
	return 0, nil
}

func (vr VisitsResponse) String() (string, error) {
	logger.Debug("Converting VisitResponse to String")
	var buf bytes.Buffer
	e := json.NewEncoder(&buf)
	e.SetEscapeHTML(false)
	e.Encode(vr)
	return buf.String(), nil
}

func ParseApiResponseToVisitsResponseStruct(data []byte) (vr VisitsResponse) {
	logger.Debug("Parsing API response to VisitsResponse")
	err := json.Unmarshal(data, &vr)
	if err != nil {
		logger.Error("Unmarshalling data to string has failed", zap.Error(err))
	}
	return vr
}

type VisitData struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Attributes VisitAttributes `json:"attributes"`
}

type Visits struct {
	Data []VisitData
}

func (v Visits) GetAllData() string {
	logger.Debug("Creating []byte to print VisitData")
	var buffer bytes.Buffer
	e := json.NewEncoder(&buffer)
	e.SetEscapeHTML(false)
	for _, visit := range v.Data {
		e.Encode(visit)
	}
	return buffer.String()
}

func LoopThroughVistsData(d Visits, start int, end int, f Flags) (Visits, error) {
	logger.Debug("Looping through LeadsData")
	for i := start; i <= end; i++ {
		logger.Info("Starting loops", zap.Int("current", i), zap.Int("end", end))
		ep_data, err := GetEndPointData("visits", f.BaseURL, f.Token, f.AccountID, TodayOrDate(f.StartDate), TodayOrDate(f.EndDate), f.PageSize, i)
		if err != nil {
			return Visits{}, err
		}
		_, _, data, _ := ep_data.GetData()

		d.Data = append(d.Data, data...)
	}
	return d, nil
}

type VisitAttributes struct {
	Source       string       `json:"source"`
	Medium       string       `json:"medium"`
	ReferringURL string       `json:"referring_url"`
	PageDepth    int          `json:"page_depth"`
	VisitRoute   []VisitRoute `json:"visit_route"`
	Keyword      string       `json:"keyword"`
	QueryTerm    string       `json:"query_term"`
	VisitLength  int          `json:"visit_length"`
	StartedAt    time.Time    `json:"started_at"`
	Campaign     string       `json:"campaign"`
	Date         string       `json:"date"`
	Hour         int          `json:"hour"`
	LfClientID   string       `json:"lf_client_id"`
	GaClientIDs  []string     `json:"ga_client_ids"`
	LeadID       string       `json:"lead_id"`
}

type VisitRoute struct {
	Hostname         string `json:"hostname"`
	PagePath         string `json:"page_path"`
	PreviousPagePath string `json:"previous_page_path"`
	TimeOnPage       int    `json:"time_on_page"`
	PageTitle        string `json:"page_title"`
	PageURL          string `json:"page_url"`
	DisplayPageName  string `json:"display_page_name"`
}

// --------------------------------------
type Links struct {
	Self     string `json:"self"`
	First    string `json:"first,omitempty"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
	Last     string `json:"last,omitempty"`
}
