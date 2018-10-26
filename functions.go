package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	igc "github.com/marni/goigc"
	"gopkg.in/mgo.v2/bson"
)

// This function is heavily influenced by:
// https://stackoverflow.com/questions/36530251/golang-time-since-with-months-and-years

// with some modifications to write the uptime of the program in the ISO 8601 standard:
// https://en.wikipedia.org/wiki/ISO_8601#Durations

func calcTime(a time.Time) string {

	b := time.Now()

	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year := int(y2 - y1)
	month := int(M2 - M1)
	day := int(d2 - d1)
	hour := int(h2 - h1)
	min := int(m2 - m1)
	sec := int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}
	returnVal := strings.Join([]string{
		"P", strconv.Itoa(year),
		"Y", strconv.Itoa(month),
		"M", strconv.Itoa(day),
		"D", "T", strconv.Itoa(hour),
		"H", strconv.Itoa(min),
		"M", strconv.Itoa(sec),
		"S"},
		"")
	return returnVal
}

func JsonStringResponse(w http.ResponseWriter, statuscode int, text interface{}) {
	r, _ := json.Marshal(text)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statuscode)
	w.Write(r)
}

// This is a funtion to calculate the track_lenght variable based on a set of given coordinates
func getTrackLenght(s igc.Track) float64 {
	totalDistance := 0.0
	// Loops through all given coordinates and adds to the total distance variable
	for i := 0; i < len(s.Points)-1; i++ {
		totalDistance += s.Points[i].Distance(s.Points[i+1])
	}
	return totalDistance
}

func invokeWebhooks(w http.ResponseWriter, r *http.Request) {
	fmt.Println("###################INVOKING WEBHOOKS##########################")
	webhooks, err := IGF.getAllWebhooks()
	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
	}

	track, err := IGF.FindLatest()
	if err != nil {
		fmt.Println("FindLatest failed")
		handleError(w, r, err, http.StatusBadRequest)
		return
	}
	for _, hook := range webhooks {
		start := time.Now()
		type rStruct struct {
			Latest     int64         `json:"t_latest"`
			Tracks     []string      `json:"tracks"`
			Processing time.Duration `json:"processing"`
		}
		trackLatestArr, err := IGF.FindOldestByIdWebhook(int(hook.LatestKnownTrack))
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		var trackks []int64
		for i := 0; i < len(trackLatestArr); i++ {
			trackks = append(trackks, (trackLatestArr[i].Timestamp))
		}
		var trackksID []string
		for i := 0; i < len(trackLatestArr); i++ {
			trackksID = append(trackksID, (trackLatestArr[i].ID.Hex()))
		}

		sliceLength := len(trackks)

		returnSt := rStruct{
			Latest:     track.Timestamp,
			Tracks:     trackksID,
			Processing: time.Since(start) / time.Millisecond,
		}
		if sliceLength > hook.MinTriggerValue {

			payload := "Latest timestamp: " + strconv.Itoa(int(returnSt.Latest)) + ", " + strconv.Itoa(len(trackksID)) + " new tracks are: "
			for _, track := range returnSt.Tracks {
				payload = fmt.Sprintf("%s, %s", payload, track)
			}
			payload = fmt.Sprintf("%s. (processing %dms)", payload, returnSt.Processing)

			type returnContent struct {
				Payload string `json:"content"`
			}

			payloadStruct := returnContent{payload}
			teemo, _ := json.Marshal(payloadStruct)
			_, err = http.Post(hook.WebhookURL, "application/json", bytes.NewBuffer(teemo))
			NowLatest := track.Timestamp
			db.C(WEBHOOKS).Update(bson.M{"_id": hook.ID}, bson.M{"$set": bson.M{"latestKnownTrack": NowLatest}})
		} else {
			fmt.Println("minTriggerValue not high enough")
			return
		}
	}

}
