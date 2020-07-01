package cmd

/*
Copyright © 2020 Jean-Marc Meessen, ON4KJM <on4kjm@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	//"fmt"
)

//TODO: validate a record for minimal values

// LogLine is used to store all the data of a single log line
type LogLine struct {
	Date           string
	MyCall         string
	Operator       string
	MyWWFF         string
	MySOTA         string
	QslMsg         string
	Nickname       string
	Mode           string
	ModeType       string
	Band           string
	BandLowerLimit float64
	BandUpperLimit float64
	Frequency      string
	Time           string
	ActualTime     string //time actually recorded in FLE
	Call           string
	Comment        string
	QSLmsg         string
	OMname         string
	GridLoc        string
	RSTsent        string
	RSTrcvd        string
	WWFF           string
	SOTA           string
}

var regexpIsFullTime = regexp.MustCompile("^[0-2]{1}[0-9]{3}$")
var regexpIsTimePart = regexp.MustCompile("^[0-5]{1}[0-9]{1}$|^[1-9]{1}$")
var regexpIsOMname = regexp.MustCompile("^@")
var regexpIsGridLoc = regexp.MustCompile("^#")
var regexpIsRst = regexp.MustCompile("^[\\d]{1,3}$")
var regexpIsFreq = regexp.MustCompile("^[\\d]+\\.[\\d]+$")

// ParseLine cuts a FLE line into useful bits
func ParseLine(inputStr string, previousLine LogLine) (logLine LogLine, errorMsg string) {
	//TODO: input null protection?

	//Flag telling that we are processing data to the right of the callsign
	isRightOfCall := false

	//Flag used to know if we are parsing the Sent RST (first) or received RST (second)
	haveSentRST := false

	//TODO: Make something more intelligent
	//TODO: What happens if we have partial lines
	previousLine.Call = ""
	previousLine.RSTsent = ""
	previousLine.RSTrcvd = ""
	previousLine.SOTA = ""
	previousLine.WWFF = ""
	previousLine.OMname = ""
	previousLine.GridLoc = ""
	previousLine.Comment = ""
	previousLine.ActualTime = ""
	logLine = previousLine

	//TODO: what happens when we have <> or when there are multiple comments
	//TODO: Refactor this! it is ugly
	comment, inputStr := getBraketedData(inputStr, COMMENT)
	if comment != "" {
		logLine.Comment = comment
	}

	QSLmsg, inputStr := getBraketedData(inputStr, QSL)
	if QSLmsg != "" {
		logLine.QSLmsg = QSLmsg
	}

	elements := strings.Fields(inputStr)

	for _, element := range elements {

		//  Is it a mode?
		if lookupMode(strings.ToUpper(element)) {
			logLine.Mode = strings.ToUpper(element)
			//TODO: improve this: what if the band is at the end of the line
			// Set the default RST depending of the mode
			if (logLine.RSTsent == "") || (logLine.RSTrcvd == "") {
				// get default RST and Mode category
				modeType, defaultReport := getDefaultReport(logLine.Mode)
				logLine.ModeType = modeType
				logLine.RSTsent = defaultReport
				logLine.RSTrcvd = defaultReport

			} else {
				errorMsg = errorMsg + "Double definitiion of RST"
			}
			continue
		}

		// Is it a band?
		isBandElement, bandLowerLimit, bandUpperLimit := IsBand(element)
		if isBandElement {
			logLine.Band = strings.ToLower(element)
			logLine.BandLowerLimit = bandLowerLimit
			logLine.BandUpperLimit = bandUpperLimit
			continue
		}

		// Is it a Frequency?
		if regexpIsFreq.MatchString(element) {
			var qrg float64
			qrg, _ = strconv.ParseFloat(element, 32)
			if (logLine.BandLowerLimit != 0.0) && (logLine.BandUpperLimit != 0.0) {
				if (qrg >= logLine.BandLowerLimit) && (qrg <= logLine.BandUpperLimit) {
					logLine.Frequency = fmt.Sprintf("%.3f", qrg)
				} else {
					logLine.Frequency = ""
					errorMsg = errorMsg + " Frequency " + element + " is invalid for " + logLine.Band + " band"
				}
			} else {
				errorMsg = errorMsg + " Unable to load frequency " + element + ": no band defined."
			}
			continue
		}

		// Is it a call sign ?
		if validCallRegexp.MatchString(strings.ToUpper(element)) {
			callErrorMsg := ""
			logLine.Call, callErrorMsg = ValidateCall(element)
			errorMsg = errorMsg + callErrorMsg
			isRightOfCall = true
			continue
		}

		// Is it a "full" time ?
		if isRightOfCall == false {
			if regexpIsFullTime.MatchString(element) {
				logLine.Time = element
				logLine.ActualTime = element
				continue
			}

			// Is it a partial time ?
			if regexpIsTimePart.MatchString(element) {
				if logLine.Time == "" {
					logLine.Time = element
					logLine.ActualTime = element
				} else {
					goodPart := logLine.Time[:len(logLine.Time)-len(element)]
					logLine.Time = goodPart + element
					logLine.ActualTime = goodPart + element
				}
				continue
			}
		}

		// Is it the OM's name (starting with "@")
		if regexpIsOMname.MatchString(element) {
			logLine.OMname = strings.TrimLeft(element, "@")
			continue
		}

		// Is it the Grid Locator (starting with "#")
		if regexpIsGridLoc.MatchString(element) {
			logLine.GridLoc = strings.TrimLeft(element, "#")
			continue
		}

		if isRightOfCall {
			//This is probably a RST
			if regexpIsRst.MatchString(element) {
				workRST := ""
				switch len(element) {
				case 1:
					if logLine.ModeType == "CW" {
						workRST = "5" + element + "9"
					} else {
						if logLine.ModeType == "PHONE" {
							workRST = "5" + element
						}
					}
				case 2:
					if logLine.ModeType == "CW" {
						workRST = element + "9"
					} else {
						if logLine.ModeType == "PHONE" {
							workRST = element
						}
					}
				case 3:
					if logLine.ModeType == "CW" {
						workRST = element
					} else {
						workRST = "*" + element
						errorMsg = errorMsg + "Invalid report (" + element + ") for " + logLine.ModeType + " mode "
					}
				}
				if haveSentRST {
					logLine.RSTrcvd = workRST
				} else {
					logLine.RSTsent = workRST
					haveSentRST = true
				}
				continue
			}

			// Is it a WWFF to WWFF reference?
			workRef, wwffErr := ValidateWwff(element)
			if wwffErr == "" {
				logLine.WWFF = workRef
				continue
			}

			// Is it a WWFF to WWFF reference?
			workRef, sotaErr := ValidateSota(element)
			if sotaErr == "" {
				logLine.SOTA = workRef
				continue
			}
		}

		//If we come here, we could not make sense of what we found
		errorMsg = errorMsg + "Unable to parse " + element + " "

	}

	//If no report is present, let's fill it with mode default
	if logLine.RSTsent == "" {
		_, logLine.RSTsent = getDefaultReport(logLine.Mode)
	}
	if logLine.RSTrcvd == "" {
		_, logLine.RSTrcvd = getDefaultReport(logLine.Mode)
	}

	//For debug purposes
	//fmt.Println("\n", SprintLogRecord(logLine))

	return logLine, errorMsg
}

func lookupMode(lookup string) bool {
	switch lookup {
	case
		"CW",
		"SSB",
		"AM",
		"FM",
		"RTTY",
		"FT8",
		"PSK",
		"JT65",
		"JT9",
		"FT4",
		"JS8",
		"ARDOP",
		"ATV",
		"C4FM",
		"CHIP",
		"CLO",
		"CONTESTI",
		"DIGITALVOICE",
		"DOMINO",
		"DSTAR",
		"FAX",
		"FSK441",
		"HELL",
		"ISCAT",
		"JT4",
		"JT6M",
		"JT44",
		"MFSK",
		"MSK144",
		"MT63",
		"OLIVIA",
		"OPERA",
		"PAC",
		"PAX",
		"PKT",
		"PSK2K",
		"Q15",
		"QRA64",
		"ROS",
		"RTTYM",
		"SSTV",
		"T10",
		"THOR",
		"THRB",
		"TOR",
		"V4",
		"VOI",
		"WINMOR",
		"WSPR":
		return true
	}
	return false
}
