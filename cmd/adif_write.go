package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

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

// outputAdif generates and writes data in ADIF format
func outputAdif(outputFile string, fullLog []LogLine) {

	//convert the log data to an in-memory ADIF file
	adifData := buildAdif(fullLog)

	//write to a file
	writeAdif(outputFile, adifData)
}

// buildAdif creates the adif file in memory ready to be printed
func buildAdif(fullLog []LogLine) (adifList []string) {
	//Print the fixed header
	adifList = append(adifList, "ADIF Export for Fast Log Entry by DF3CB")
	adifList = append(adifList, "<PROGRAMID:3>FLE")
	adifList = append(adifList, "<ADIF_VER:5>3.0.6")
	adifList = append(adifList, "<EOH>")

	for _, logLine := range fullLog {
		adifLine := ""
		adifLine = adifLine + adifElement("STATION_CALLSIGN", logLine.MyCall)
		adifLine = adifLine + adifElement("CALL", logLine.Call)
		adifLine = adifLine + adifElement("QSO_DATE", adifDate(logLine.Date))
		adifLine = adifLine + adifElement("TIME_ON", logLine.Time)
		adifLine = adifLine + adifElement("BAND", logLine.Band)
		adifLine = adifLine + adifElement("MODE", logLine.Mode)
		if logLine.Frequency != "" {
			adifLine = adifLine + adifElement("FREQ", logLine.Frequency)
		}
		adifLine = adifLine + adifElement("RST_SENT", logLine.RSTsent)
		adifLine = adifLine + adifElement("RST_RCVD", logLine.RSTrcvd)
		adifLine = adifLine + adifElement("MY_SIG", "WWFF")
		adifLine = adifLine + adifElement("MY_SIG_INFO", logLine.MyWWFF)
		adifLine = adifLine + adifElement("OPERATOR", logLine.Operator)
		if logLine.Nickname != "" {
			adifLine = adifLine + adifElement("APP_EQSL_QTH_NICKNAME", logLine.Nickname)
		}
		adifLine = adifLine + "<EOR>"

		adifList = append(adifList, adifLine)

	}

	return adifList
}

// writeAdif writes the in-memory adif data to a file
func writeAdif(outputFile string, adifData []string) {

	//TODO: check access rights
	f, err := os.Create(outputFile)
	checkFileError(err)

	defer f.Close()

	w := bufio.NewWriter(f)

	lineCount := 0
	for _, adifLine := range adifData {
		_, err := w.WriteString(adifLine + "\n")
		checkFileError(err)

		w.Flush()
		checkFileError(err)
		lineCount++
	}
	fmt.Printf("\nSuccessfully wrote %d lines to file \"%s\"", lineCount, outputFile)
}

// adifElement generated the ADIF sub-element
func adifElement(elementName, elementValue string) (element string) {
	return fmt.Sprintf("<%s:%d>%s ", strings.ToUpper(elementName), len(elementValue), elementValue)
}

// checkFileError handles file related errors
func checkFileError(e error) {
	if e != nil {
		panic(e)
	}
}

//adifDate converts a date in YYYY-MM-DD format to YYYYMMDD
func adifDate(inputDate string) (outputDate string) {
	const RFC3339FullDate = "2006-01-02"
	date, err := time.Parse(RFC3339FullDate, inputDate)
	//error should never happen
	if err != nil {
		panic(err)
	}
	outputDate = fmt.Sprintf("%04d%02d%02d", date.Year(), date.Month(), date.Day())

	return outputDate
}
