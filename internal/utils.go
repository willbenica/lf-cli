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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger      *zap.Logger
	LogConfig   *zap.Config
	initialized bool
)

// // initLogger is used to initialize logging for the program
func InitLogger() (*zap.Logger, *zap.Config) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build()
	defer logger.Sync()
	return logger, &config
}

// Init checks if a logger has been initialized and if not, initializes one
func Init() {
	if !initialized {
		logger, LogConfig = InitLogger()
	}
	initialized = true
}

// IsValidEndpoint returns `true` if an endpoint is valid or `false`
func IsValidEndpoint(ep string) bool {
	Init()
	logger.Debug("Checking if provided EndPoint is valid (leads, visits)", zap.String("provided EndPoint", ep))
	switch ep {
	case
		"leads",
		"visits":
		return true
	}
	return false
}

// GetEndPointData returns the data from a specified endpoint
func GetEndPointData(ep string, baseURL string, token string, accountID string, startDate string, endDate string, pageSize int, pageNumber int) (EndPoint, error) {
	Init()
	logger.Debug("Creating URL")
	url, err := EndpointURLBuilder(baseURL, ep, accountID, startDate, endDate, pageSize, pageNumber)
	if err != nil {
		logger.Error("You are trying to use https - this is unsafe", zap.Error(err))
		return nil, err
	}
	logger.Debug("URL created", zap.String("URL", url))
	logger.Debug("Requesting data from URL", zap.String("URL", url))
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Request did not complete successfully", zap.String("HTTP Verb", "GET"), zap.String("URL", url), zap.Error(err))
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Add("User-Agent", "lf-cli")
	request.Header.Add("Accept", "*/*")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Error("Issue with http.DefaultClient", zap.Error(err))
		return nil, err
	}
	defer response.Body.Close()

	// body holds the data that needs to be parsed!
	logger.Debug("Reading data returned from leadfeeder")
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Error("Error reading the response body", zap.Error(err))
		return nil, err
	}

	switch ep {
	case "leads":
		logger.Debug("Parsing leads into a LeadsResponse")
		lr := ParseApiResponseToLeadsResponseStruct(body)
		return lr, nil
	case "visits":
		logger.Debug("Parsing leads into a VisitsResponse")
		vr := ParseApiResponseToVisitsResponseStruct(body)
		return vr, nil
	default:
		logger.DPanic("this branch should never be reached")
		return nil, nil
	}
}

// EndpointURLBuilder is used to ensure that URL provided are formatted correctly
func EndpointURLBuilder(rawBaseURL string, endpoint string, accountID string, startDate string, endDate string, pageSize int, pageNumber int) (string, error) {
	Init()
	baseURL := baseURLBuilder(rawBaseURL)
	return "https://" + baseURL +
		"/accounts/" + accountID +
		"/" +
		endpoint +
		"?start_date=" + startDate +
		"&end_date=" + endDate +
		"&page%5Bsize%5D=" + fmt.Sprint(pageSize) +
		"&page%5Bnumber%5D=" + fmt.Sprint(pageNumber), nil
}

// baseURLBuilder determines the base URL (e.g. api.leadfeeder.me) stripping protocol trailing `/`
func baseURLBuilder(rawBaseURL string) (baseURL string) {
	Init()
	baseURL = rawBaseURL
	if strings.Contains(baseURL, "https://") {
		baseURL = baseURL[8:]
	}
	if strings.Contains(baseURL, "http://") {
		baseURL = baseURL[7:]
	}
	idxSlash := strings.Index(baseURL, "/")
	if idxSlash >= 0 {
		baseURL = baseURL[:idxSlash]
	}
	return
}

// WriteToFile should write data to the file provided under path
func WriteToFile(path string, filename string, data string) error {
	Init()
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if path == "" {
		path = "."
	}
	logger.Debug("Determining current working directory and its relation to the desired path", zap.String("cwd", cwd), zap.String("desired path", path))
	CreateDirectoryIfNotExists(path)
	file, err := os.Create(path + "/" + filename)
	if err != nil {
		logger.Error("Failed to create file", zap.String("file", fmt.Sprintf("%s/%s", path, filename)), zap.Error(err))
		return err
	}
	logger.Info("File created successfully")
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	file.Sync()
	return nil
}

func CreateDirectoryIfNotExists(name string) {
	Init()
	path, _ := os.Getwd()
	logger.Info("Trying to create a folder", zap.String("folder", fmt.Sprintf("%s/%s", path, name)))
	if _, err := os.Stat(fmt.Sprintf("%s/%s", path, name)); os.IsNotExist(err) {
		err = os.Mkdir(name, os.ModePerm)
		if err != nil {
			logger.Error("creating folder failed", zap.Error(err))
		}
	}
}

// CreateFileName is used to create standardized file names for outputs
func CreateFileName(ep string, f Flags) string {
	Init()
	if f.StartDate == f.EndDate {
		return fmt.Sprintf("%s_from_%s.json", ep, f.StartDate)
	}
	return fmt.Sprintf("%s_from_%s_to_%s.json", ep, f.StartDate, f.EndDate)
}

func TodayOrDate(possibleDate string) string {
	Init()
	if strings.ToLower(possibleDate) == "today" {
		return time.Now().Format("2006-01-02")
	}
	return possibleDate
}

type Flags struct {
	All        bool
	StartDate  string
	EndDate    string
	PageSize   int
	PageNumber int
	BaseURL    string
	Token      string
	AccountID  string
	Verbose    bool
}
