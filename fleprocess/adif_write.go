package fleprocess

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
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// OutputAdif generates and writes data in ADIF format
func OutputAdif(outputFile string, fullLog []LogLine, isWWFF bool, isSOTA bool) {

	//convert the log data to an in-memory ADIF file
	adifData := buildAdif(fullLog, isWWFF, isSOTA)

	//write to a file
	writeFile(outputFile, adifData)
}

// buildAdif creates the adif file in memory ready to be printed
func buildAdif(fullLog []LogLine, isWWFF bool, isSOTA bool) (adifList []string) {
	//Print the fixed header
	adifList = append(adifList, "ADIF Export for Fast Log Entry by DF3CB")
	adifList = append(adifList, "<PROGRAMID:3>FLE")
	adifList = append(adifList, "<ADIF_VER:5>3.1.0")
	adifList = append(adifList, "<EOH>")

	for _, logLine := range fullLog {
		var adifLine strings.Builder
		adifLine.WriteString(adifElement("STATION_CALLSIGN", logLine.MyCall))
		adifLine.WriteString(adifElement("CALL", logLine.Call))
		adifLine.WriteString(adifElement("QSO_DATE", adifDate(logLine.Date)))
		adifLine.WriteString(adifElement("TIME_ON", logLine.Time))
		adifLine.WriteString(adifElement("BAND", logLine.Band))
		adifLine.WriteString(adifElement("MODE", logLine.Mode))
		if logLine.Frequency != "" {
			adifLine.WriteString(adifElement("FREQ", logLine.Frequency))
		}
		adifLine.WriteString(adifElement("RST_SENT", logLine.RSTsent))
		adifLine.WriteString(adifElement("RST_RCVD", logLine.RSTrcvd))
		if logLine.Comment != "" {
			adifLine.WriteString(adifElement("COMMENT", logLine.Comment))
		}
		if logLine.OMname != "" {
			adifLine.WriteString(adifElement("NAME", logLine.OMname))
		}
		if logLine.QSLmsg != "" {
			adifLine.WriteString(adifElement("QSLMSG", logLine.QSLmsg))
		}
		if isWWFF {
			adifLine.WriteString(adifElement("MY_SIG", "WWFF"))
			adifLine.WriteString(adifElement("MY_SIG_INFO", logLine.MyWWFF))
		}
		if isSOTA {
			adifLine.WriteString(adifElement("MY_SOTA_REF", logLine.MySOTA))
			if logLine.SOTA != "" {
				adifLine.WriteString(adifElement("SOTA_REF", logLine.SOTA))
			}
		}
		adifLine.WriteString(adifElement("OPERATOR", logLine.Operator))
		if logLine.Nickname != "" {
			adifLine.WriteString(adifElement("APP_EQSL_QTH_NICKNAME", logLine.Nickname))
		}
		adifLine.WriteString("<EOR>")

		adifList = append(adifList, adifLine.String())

	}

	return adifList
}

// writeFile writes the in-memory data (lines) to a file
func writeFile(outputFile string, adifData []string) {

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
	fmt.Printf("\nSuccessfully wrote %d lines to file \"%s\"\n", lineCount, outputFile)
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
	const FLEdateFormat = "2006-01-02"
	date, err := time.Parse(FLEdateFormat, inputDate)
	//error should never happen
	if err != nil {
		panic(err)
	}

	const ADIFdateFormat = "20060102"
	return date.Format(ADIFdateFormat)
}