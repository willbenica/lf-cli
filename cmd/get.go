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
package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/willbenica/lf-cli/internal"
	"go.uber.org/zap"
)

const (
	missingEndpointMsg = "an endpoint is required"
	invalidEndPointMsg = "invalid endpoint specified: %s"
)

var (
	// Folder is where data should be written to
	Folder string
	Cwd    string
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <endpoint name>",
	Short: "Get the data from an endpoint, e.g. 'leads' or 'visits'",
	Long: `Get data from one of the following endpoints:
https://api.leadfeeder.com/accounts/...
	leads
	visits (all visits)
  Unsuported: Getting an invidvidual lead's visists`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New(missingEndpointMsg)
		}
		if internal.IsValidEndpoint(args[0]) {
			return nil
		}
		return fmt.Errorf(invalidEndPointMsg, args[0])
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		flags := internal.Flags{
			StartDate:  internal.TodayOrDate(startDate),
			EndDate:    internal.TodayOrDate(endDate),
			PageSize:   pageSize,
			PageNumber: pageNumber,
			BaseURL:    baseURL,
			Token:      token,
			AccountID:  accountID,
		}

		// Raise loglevel to Error if use is printing response to the console
		if !all || len(Folder) > 0 || quiet {
			logConfig.Level.SetLevel(zap.ErrorLevel)
			internal.LogConfig.Level.SetLevel(zap.ErrorLevel)
		}
		//Check if user wants expanded log output and if necessary override the above setting
		if verbose {
			logConfig.Level.SetLevel(zap.DebugLevel)
			internal.LogConfig.Level.SetLevel(zap.DebugLevel)
		}

		if !all {
			logger.Debug("Retrieving ONLY one response, not looping to the last page")
			data, err := internal.GetEndPointData(args[0], baseURL, token, accountID, flags.StartDate, flags.EndDate, flags.PageSize, flags.PageNumber)
			if err != nil {
				return err
			}

			// Are we writing to the default folder or just printing to the console?
			if len(Folder) != 0 {
				switch args[0] {
				case "leads":
					var ld internal.Leads
					var lc internal.Locations
					epld, eplc, _, _ := data.GetData()
					ld.Data = epld
					lc.Data = eplc
					errWrite := internal.WriteToFile(Folder, internal.CreateFileName("leads", flags), ld.GetAllData())
					if errWrite != nil {
						logger.Error("failed to write to file", zap.Error(errWrite))
					}
					errLoc := internal.WriteToFile(Folder, internal.CreateFileName("locations", flags), lc.GetAllData())
					if errLoc != nil {
						logger.Error("failed to write to file", zap.Error(errLoc))
					}
				case "visits":
					var v internal.Visits
					_, _, epv, _ := data.GetData()
					v.Data = epv
					errWrite := internal.WriteToFile(Folder, internal.CreateFileName("visits", flags), v.GetAllData())
					if errWrite != nil {
						logger.Error("failed to write to file", zap.Error(errWrite))
					}
				default:
					logger.DPanic("We should never reach this statement")
				}
			}

			// If no folder is specified to write, then just print to the console
			if data == nil {
				return nil
			}
			dataAsString, err := data.String()
			if err != nil {
				return err
			}
			fmt.Println(dataAsString)
			return nil
		}
		//-----------------------------//
		// Parse the complete endpoint //
		//-----------------------------//

		// Parse all endpoint data
		startTime := time.Now()

		// Can be removed after implementation
		logger.Info("Getting All Data")

		baseData, err := internal.GetEndPointData(args[0], baseURL, token, accountID, internal.TodayOrDate(startDate), internal.TodayOrDate(endDate), pageSize, pageNumber)
		if err != nil {
			return err
		}
		if baseData != nil {
			lastpage, _ := baseData.GetLastPageNumber()
			switch baseData.Type() {
			case "LeadsResponse":
				logger.Debug("Getting LeadData & LocationData")
				leads, locations, _, _ := baseData.GetData()
				logger.Debug("Starting to loop through LeadData & LocationData")
				var allLeads internal.Leads

				loopedLeads, loopedLocations, err := internal.LoopThroughLeadsData(leads, locations, pageNumber+1, lastpage, flags)
				if err != nil {
					return err
				}
				logger.Debug("Finished looping through LeadData & LocationData")

				allLeads.Data = loopedLeads

				leadsFile := internal.CreateFileName("leads", flags)
				logger.Info("Writing to file:", zap.String("file", leadsFile))
				errLeads := internal.WriteToFile(Folder, leadsFile, allLeads.GetAllData())
				if errLeads != nil {
					return errLeads
				}
				logger.Info("File written", zap.String("file", leadsFile))

				var allLocations internal.Locations
				allLocations.Data = loopedLocations
				locationsFile := internal.CreateFileName("locations", flags)
				logger.Info("Writing to file", zap.String("file", locationsFile))
				errLocations := internal.WriteToFile(Folder, locationsFile, allLocations.GetAllData())
				if errLocations != nil {
					return errLocations
				}
				logger.Info("File written", zap.String("file", locationsFile))

			case "VisitsResponse":
				logger.Debug("Getting VisitData")
				_, _, data, _ := baseData.GetData()
				var v internal.Visits
				v.Data = data
				logger.Debug("Starting to loop through VisitData")
				loopedVisits, err := internal.LoopThroughVistsData(v, pageNumber+1, lastpage, flags)
				if err != nil {
					return err
				}
				logger.Debug("Finished looping through VisitData")
				visitsFile := internal.CreateFileName("visits", flags)
				logger.Info("Writing to file", zap.String("file", visitsFile))
				errLocations := internal.WriteToFile(Folder, visitsFile, loopedVisits.GetAllData())
				if errLocations != nil {
					return errLocations
				}
				logger.Info("File written", zap.String("file", visitsFile))

			default:
				logger.DPanic("We should never end up here!")
			}

			logger.Info("Process complete")
			logger.Info("Process took", zap.Duration("duration", time.Since(startTime)))
			return nil
		}
		logger.Error("Data failed to process - unknown issue.")
		return fmt.Errorf("we ran into an unknown issue when trying to collect all data")
	},
	ValidArgs: []string{"leads", "custom-feeds", "visits"},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().SortFlags = false

	Cwd, _ = os.Getwd()

	getCmd.Flags().StringVarP(&Folder, "folder", "f", "", "NEEDS IMPROVEMENT: Folder where data should be written")
	getCmd.Flags().MarkHidden("folder") // Marking this as hidden, so that it's not used
	getCmd.Flags().StringVarP(&startDate, "start-date", "s", "today", "Start of the time period to return data. Use YYYY-MM-DD or today")
	getCmd.Flags().StringVarP(&endDate, "end-date", "e", "today", "End of the time period to return data. Use YYYY-MM-DD or today")
	getCmd.Flags().IntVarP(&pageSize, "page-size", "z", 100, "Number of results to return per page, 1-100")
	getCmd.Flags().IntVarP(&pageNumber, "page-number", "n", 1, "Page to retrieve")
	getCmd.Flags().BoolVarP(&all, "get-all", "a", false, "Get all data for this endpoint - loop from start to last page")
}
