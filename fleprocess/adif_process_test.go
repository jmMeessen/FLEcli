package fleprocess

import (
	"fmt"
	"testing"
)

func Test_validateDataforAdif(t *testing.T) {
	type args struct {
		loadedLogFile []LogLine
		isWWFFcli     bool
		isSOTAcli     bool
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			"Happy Case (no sota or wwff)",
			args{isWWFFcli: false, isSOTAcli: false, loadedLogFile: []LogLine{
				{Date: "date", MyCall: "myCall", MySOTA: "mySota", Mode: "mode", Band: "band", Time: "time", Call: "call"},
				{Date: "date", MyCall: "myCall", MySOTA: "mySota", Mode: "mode", Band: "band", Time: "time", Call: "call"},
				{Date: "date", MyCall: "myCall", MySOTA: "mySota", Mode: "mode", Band: "band", Time: "time", Call: "call"}},
			},
			nil,
		},
		{
			"No data",
			args{isWWFFcli: false, isSOTAcli: false, loadedLogFile: []LogLine{}},
			fmt.Errorf("No QSO found"),
		},
		{
			"Missing Date",
			args{isWWFFcli: false, isSOTAcli: false, loadedLogFile: []LogLine{
				{Date: "date", MyCall: "myCall", MySOTA: "mySota", Mode: "mode", Band: "band", Time: "12:01", Call: "call"},
				{Date: "", MyCall: "myCall", MySOTA: "mySota", Mode: "mode", Band: "band", Time: "12:02", Call: "call"},
				{Date: "", MyCall: "myCall", MySOTA: "mySota", Mode: "mode", Band: "band", Time: "12:03", Call: "call"}},
			},
			fmt.Errorf("missing date for log entry at 12:02 (#2), missing date for log entry at 12:03 (#3)"),
		},
		{
			"Missing MyCall",
			args{isWWFFcli: true, isSOTAcli: true, loadedLogFile: []LogLine{
				{Date: "date", MyCall: "", MySOTA: "mySota", Mode: "mode", Band: "band", Time: "12:01", Call: "call"},
				{Date: "date", MyCall: "", MySOTA: "mySota", Mode: "mode", Band: "band", Time: "12:02", Call: "call"},
				{Date: "date", MyCall: "", MySOTA: "mySota", Mode: "mode", Band: "band", Time: "12:03", Call: "call"}},
			},
			fmt.Errorf("Missing MyCall"),
		},
		{
			"Missing MySota",
			args{isWWFFcli: false, isSOTAcli: true, loadedLogFile: []LogLine{
				{Date: "date", MyCall: "myCall", MySOTA: "", Mode: "mode", Band: "band", Time: "time", Call: "call"},
				{Date: "date", MyCall: "myCall", MySOTA: "", Mode: "mode", Band: "band", Time: "time", Call: "call"},
				{Date: "date", MyCall: "myCall", MySOTA: "", Mode: "mode", Band: "band", Time: "time", Call: "call"}},
			},
			fmt.Errorf("Missing MY-SOTA reference"),
		},
		{
			"Misc. missing data (Band, Time, Mode, Call)",
			args{isWWFFcli: false, isSOTAcli: false, loadedLogFile: []LogLine{
				{Date: "date", MyCall: "myCall", MySOTA: "mySota", Mode: "mode", Band: "", Time: "", Call: "call"},
				{Date: "date", MyCall: "myCall", MySOTA: "mySota", Mode: "", Band: "band", Time: "12:02", Call: "call"},
				{Date: "date", MyCall: "myCall", MySOTA: "mySota", Mode: "mode", Band: "band", Time: "12:03", Call: ""}},
			},
			fmt.Errorf("missing band for log entry #1, missing QSO time for log entry #1, missing mode for log entry at 12:02 (#2), missing call for log entry at 12:03 (#3)"),
		},
		{
			"Missing MY-WWFF",
			args{isWWFFcli: true, isSOTAcli: false, loadedLogFile: []LogLine{
				{Date: "date", MyCall: "myCall", MySOTA: "mySota", MyWWFF: "", Mode: "mode", Band: "band", Time: "time", Call: "call"},
				{Date: "date", MyCall: "myCall", MySOTA: "mySota", MyWWFF: "", Mode: "mode", Band: "band", Time: "time", Call: "call"},
				{Date: "date", MyCall: "myCall", MySOTA: "mySota", MyWWFF: "", Mode: "mode", Band: "band", Time: "time", Call: "call"}},
			},
			fmt.Errorf("Missing MY-WWFF reference"),
		},
		{
			"Missing MY-WWFF",
			args{isWWFFcli: true, isSOTAcli: false, loadedLogFile: []LogLine{
				{Date: "date", MyCall: "myCall", MyWWFF: "myWwff", Operator: "", Mode: "mode", Band: "band", Time: "time", Call: "call"},
				{Date: "date", MyCall: "myCall", MyWWFF: "myWwff", Operator: "", Mode: "mode", Band: "band", Time: "time", Call: "call"},
				{Date: "date", MyCall: "myCall", MyWWFF: "myWwff", Operator: "", Mode: "mode", Band: "band", Time: "time", Call: "call"}},
			},
			fmt.Errorf("Missing Operator call sign"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateDataforAdif(tt.args.loadedLogFile, tt.args.isWWFFcli, tt.args.isSOTAcli)

			//Test the error message, if any
			if got != nil && tt.want != nil {
				if got.Error() != tt.want.Error() {
					t.Errorf("validateDataforAdif() = %v, want %v", got, tt.want)
				}
			} else {
				if !(got == nil && tt.want == nil) {
					t.Errorf("validateDataforAdif() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
