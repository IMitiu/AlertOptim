package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type RawAlert []string

type Alert struct {
	RawAlert

	Name        string
	Description string
	Type        string
	WarningLev  float64
	CriticalLev float64
	Sustain     int
	Action      string
	Links       string
	Query       string
}

type AlertInfo struct {
	Header       []string
	AlertsZUpper []Alert
	AlertsZLower []Alert
	AlertsLUpper []Alert
	AlertsLLower []Alert
	AlertsOther  []Alert
}

func NewSimple(source string) (*AlertInfo, error) {
	data, errRead := readFile(source)
	if errRead != nil {
		return nil, errRead
	}

	header, posAlerts := isolateHeader(data)
	alertData := isolateAlertData(data, posAlerts)

	res := mapAlerts(extractAlerts(alertData))
	res.Header = header

	return &res, nil
}

func (a *AlertInfo) Spool(w io.Writer) {
	s := func(data []Alert, w io.Writer) {
		for _, a := range data {
			w.Write([]byte(strings.Join(a.RawAlert, "")))
		}
	}

	w.Write([]byte(strings.Join(a.Header, "")))
	s(a.AlertsZUpper, w)
	s(a.AlertsZLower, w)
	s(a.AlertsLUpper, w)
	s(a.AlertsLLower, w)
	s(a.AlertsOther, w)
}

func mapAlerts(alerts []RawAlert) AlertInfo {
	var res AlertInfo

	for _, a := range alerts {
		alert := extractAlert(a)

		if alert.Type == "ZONE" && (alert.CriticalLev > alert.WarningLev) {
			res.AlertsZUpper = append(res.AlertsZUpper, alert)
			continue
		}

		if alert.Type == "ZONE" && (alert.CriticalLev <= alert.WarningLev) {
			res.AlertsZLower = append(res.AlertsZLower, alert)
			continue
		}

		if alert.Type == "LEGACY" && (alert.CriticalLev > alert.WarningLev) {
			res.AlertsLUpper = append(res.AlertsLUpper, alert)
			continue
		}

		if alert.Type == "LEGACY" && (alert.CriticalLev <= alert.WarningLev) {
			res.AlertsLLower = append(res.AlertsLLower, alert)
			continue
		}

		res.AlertsOther = append(res.AlertsOther, alert)
	}

	return res
}

func readFile(path string) ([]string, error) {
	f, errOpen := os.Open(path)
	if errOpen != nil {
		return nil, errOpen
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var res []string

	for scanner.Scan() {
		res = append(res, scanner.Text()+"\n")
	}
	if errScan := scanner.Err(); errScan != nil {
		return nil, errScan
	}

	return res, nil
}

func isolateHeader(data []string) ([]string, int) {
	var res []string
	var posAlertInfo int

	for i, line := range data {
		res = append(res, line)

		if strings.Contains(line, "alerts:") {
			posAlertInfo = i
			break
		}
	}

	return res, posAlertInfo
}

func isolateAlertData(data []string, posAlert int) []string {
	var res []string

	for i := posAlert + 1; i < len(data); i++ {
		res = append(res, data[i])
	}

	return res
}

func extractAlerts(data []string) []RawAlert {
	if len(data) == 0 {
		return nil
	}

	var res []RawAlert

	i := 0
	alert := []string{}

	for i < len(data) {
		if strings.Contains(data[i], "- alert:") && len(alert) > 0 {
			res = append(res, alert)
			alert = []string{}
		}

		alert = append(alert, data[i])

		if i == len(data)-1 {
			res = append(res, alert)

			break
		}

		i++
	}

	return res
}

// TODO: could refactor by placing logic in a generic node extractor.
func extractAlert(r RawAlert) Alert {
	var a Alert
	raw := []string{}
	var tabulationIndex int
	var alertTabIndex int

	for i := 0; i < len(r); i++ {
		vals := strings.Split(r[i], ":")

		switch strings.Trim(vals[0], " ") {
		case "- alert":
			{
				alertTabIndex = startPos(r[i])
				tabulationIndex = startPos(r[i+1]) // for i==1
			}

		case "name":
			{
				if startPos(r[i]) == tabulationIndex {
					a.Name = strings.Title(vals[1])
				}
			}

		case "type":
			{
				if startPos(r[i]) == tabulationIndex {
					a.Type = strings.Trim(vals[1], " \n")
				}
			}

		case "description":
			{
				if startPos(r[i]) == tabulationIndex {
					item := []string{}
					posToken := strings.Index(r[i], "description")
					item = append(item, r[i])

					for (startPos(r[i+1]) > posToken) || len(r[i+1]) == 1 {
						item = append(item, r[i+1])
						i++
					}

					a.Description = strings.Join(item, "")
				}
			}

		case "query":
			{
				if startPos(r[i]) == tabulationIndex {
					item := []string{}
					posToken := strings.Index(r[i], "query")
					item = append(item, r[i])

					for (startPos(r[i+1]) > posToken) || len(r[i+1]) == 1 {
						item = append(item, r[i+1])
						i++
					}

					a.Query = strings.Join(item, "")
				}
			}

		case "links":
			{
				if startPos(r[i]) == tabulationIndex {
					item := []string{}
					posToken := strings.Index(r[i], "links")
					item = append(item, r[i])

					for (startPos(r[i+1]) > posToken) || len(r[i+1]) == 1 {
						item = append(item, r[i+1])
						i++
					}

					a.Links = strings.Join(item, "")
				}
			}

		case "actions":
			{
				if startPos(r[i]) == tabulationIndex {
					item := []string{}
					posToken := strings.Index(r[i], "actions")
					item = append(item, r[i])

					for (startPos(r[i+1]) > posToken) || len(r[i+1]) == 1 {
						item = append(item, r[i+1])
						i++
					}

					a.Action = strings.Join(item, "")
				}
			}

		case "warn":
			{
				if startPos(r[i]) == tabulationIndex {
					var errWarn error
					a.WarningLev, errWarn = strconv.ParseFloat(strings.Trim(vals[1], "  \n"), 64)
					if errWarn != nil {
						log.Println(errWarn)
					}
				}
			}

		case "critical":
			{
				if startPos(r[i]) == tabulationIndex {
					var errCri error
					a.CriticalLev, errCri = strconv.ParseFloat(strings.Trim(vals[1], "  \n"), 64)
					if errCri != nil {
						log.Println(errCri)
					}
				}
			}

		case "sustainPeriod":
			{
				if startPos(r[i]) == tabulationIndex {
					var errSus error
					a.Sustain, errSus = strconv.Atoi(strings.Trim(vals[1], "  \n"))
					if errSus != nil {
						log.Println(errSus)
					}
				}
			}

		default:
			{
				raw = append(raw, r[i])
			}
		}
	}

	prefix := strings.Repeat(" ", tabulationIndex)
	alertTabs := strings.Repeat(" ", alertTabIndex)

	alertPrime := []string{
		alertTabs + "- alert:" + "\n",
		prefix + "name:" + a.Name,
		a.Description,
		prefix + "type:" + " " + a.Type + "\n",
		prefix + "warn:" + " " + strconv.FormatFloat(a.WarningLev, 'f', -1, 64) + "\n",
		prefix + "critical:" + " " + strconv.FormatFloat(a.CriticalLev, 'f', -1, 64) + "\n",
		prefix + "sustainPeriod:" + " " + strconv.Itoa(a.Sustain) + "\n",
		a.Action,
		a.Links,
		a.Query,
	}

	a.RawAlert = append(a.RawAlert, alertPrime...)
	a.RawAlert = append(a.RawAlert, raw...)
	a.RawAlert = append(a.RawAlert, "\n")

	return a
}

func startPos(s string) int {
	if len(s) == 0 {
		return -1
	}

	i := 0
	for i < len(s)-1 {
		if s[i:i+1] != " " {
			break
		}

		i++
	}

	return i
}
