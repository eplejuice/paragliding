package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/marni/goigc"
	"gopkg.in/mgo.v2/bson"
)

func handleRouter(w http.ResponseWriter, r *http.Request) {
	// This handles the GET /api
	regHandleParaglidingRedirect, err := regexp.Compile("^/paragliding/?$")
	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}

	regHandleParaglidingAPI, err := regexp.Compile("^/paragliding/api/?$")
	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}

	// This handles the POST/GET /api/igc
	regHandleParaglidingAPITrack, err := regexp.Compile("^/paragliding/api/track/?$")
	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}

	// This handles the	GET /api/igc/id
	regHandleParaglidingAPITrackID, err := regexp.Compile("^/paragliding/api/track/[a-zA-Z0-9]+/?$")
	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}

	// This handles the GET /api/igc/id/field
	regHandleParaglidingAPITrackIDField, err := regexp.Compile("^/paragliding/api/track/[a-zA-Z0-9]+/(pilot|glider|glider_id|track_length|H_date|track_src_url)$")
	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}

	regHandleParaglidingAPITickerLatest, err := regexp.Compile("^/paragliding/api/ticker/latest/?$")

	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}
	regHandleParaglidingAPITicker, err := regexp.Compile("^/paragliding/api/ticker/?$")

	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}
	regHandleParaglidingAPITickerTimestamp, err := regexp.Compile("^/paragliding/api/ticker/[0-9]+/?$")

	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}
	regHandlePOSTParaglidingAPIWebhookNew, err := regexp.Compile("^/paragliding/api/webhook/new_track/?$")

	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}
	regHandleParaglidingAPIWebhookNewID, err := regexp.Compile("^/paragliding/api/webhook/new_track/[0-9]+/?$")

	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}

	regHandleAdminApiTracksCount, err := regexp.Compile("^/admin/api/tracks_count/?$")

	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}

	regHandleAdminApiTracks, err := regexp.Compile("^/admin/api/tracks/?$")

	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}

	// This is a switch that always runs routes the http request to the right handlefunc
	// Otherwise the dafault gives the user a httpBadRequest response
	switch {
	case regHandleParaglidingRedirect.MatchString(r.URL.Path):
		handleParaglidingRedirect(w, r)

	case regHandleParaglidingAPI.MatchString(r.URL.Path):
		handleParaglidingAPI(w, r)

	case regHandleParaglidingAPITrack.MatchString(r.URL.Path):
		handleParaglidingAPITrack(w, r)

	case regHandleParaglidingAPITrackID.MatchString(r.URL.Path):
		handleParaglidingAPITrackID(w, r)

	case regHandleParaglidingAPITrackIDField.MatchString(r.URL.Path):
		handleParaglidingAPITrackIDField(w, r)

	case regHandleParaglidingAPITickerLatest.MatchString(r.URL.Path):
		handleParaglidingAPITickerLatest(w, r)

	case regHandleParaglidingAPITicker.MatchString(r.URL.Path):
		handleParaglidingAPITicker(w, r)

	case regHandleParaglidingAPITickerTimestamp.MatchString(r.URL.Path):
		handleParaglidingAPITickerTimestamp(w, r)

	case regHandlePOSTParaglidingAPIWebhookNew.MatchString(r.URL.Path):
		handlePOSTParaglidingAPIWebhookNew(w, r)

	case regHandleParaglidingAPIWebhookNewID.MatchString(r.URL.Path):
		handleParaglidingAPIWebhookNewID(w, r)

	case regHandleAdminApiTracksCount.MatchString(r.URL.Path):
		handleGetAdminApiTracksCount(w, r)

	case regHandleAdminApiTracks.MatchString(r.URL.Path):
		handleDeleteAdminApiTracks(w, r)
	default:
		fmt.Println("DEFAULT")
		handleError(w, r, nil, http.StatusBadRequest)
	}
}

// This function handles all the errors and writes them as a reponse to the user
// with the right error code based on the parameter recieved
func handleError(w http.ResponseWriter, r *http.Request, err error, status int) {
	if err != nil {
		http.Error(w, fmt.Sprintf("%s/t%s", http.StatusText(status), err), status)
	} else {
		http.Error(w, fmt.Sprintf(http.StatusText(status)), status)
	}
}

// Redirects a user from /paragliding/ to /paragliding/api
func handleParaglidingRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/paragliding/api", 301)
}

// Returns metadata about the program
func handleParaglidingAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
	} else {
		type metaData struct {
			Uptime  string
			Info    string
			Version string
		}
		// Using a struct to easily encode to a json
		metaInfo := metaData{
			Uptime:  calcTime(startTime),
			Info:    "Service for Paragliding track",
			Version: "v1",
		}

		// Using Marshal instead of Endoce, because i believe Marshal is used to encode strings
		// and the struct mainly consist of strings.
		metaResp, _ := json.Marshal(metaInfo)
		// Sets the header to json, and returns a json object as the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(metaResp)
	}
}

// A Router wich handles a request to see if its either a GET or a POST request
// and sends it to the right handleFunc
func handleParaglidingAPITrack(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetParaglidingAPITrack(w, r)
	case http.MethodPost:
		handlePostParaglidingAPITrack(w, r)
	default:
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
	}
}

// Returns an array containing the IDs of all stored tracks in the database
func handleGetParaglidingAPITrack(w http.ResponseWriter, r *http.Request) {

	tracks, err := IGF.FindAll()
	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
	}
	var trackks []string
	for i := 0; i < len(tracks); i++ {
		trackks = append(trackks, (tracks[i].ID.Hex()))
	}
	JsonStringResponse(w, http.StatusOK, trackks)
}

// Lets a user post a new track into the database with a url to an igcfile
func handlePostParaglidingAPITrack(w http.ResponseWriter, r *http.Request) {

	// Creating a unique monotone ID by converting a timestamp to milliseconds
	Uniq := time.Now().UnixNano() / int64(time.Millisecond)

	// Struct for handling the recieved url from the json object
	type Tmp struct {
		Url string `bson:"id" json:"url"`
	}
	var tmp Tmp

	// Decodes the json obejct and puts the Url in the struct
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&tmp); err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}

	// Checks if the Url is legit and leads to an igcfile using the marni/goigc library
	s := tmp.Url
	tmpTrack, err := igc.ParseLocation(s)
	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}

	// The struct used to put data into the database
	track := Track{
		ID:          bson.NewObjectId(),
		Timestamp:   Uniq,
		Url:         tmp.Url,
		HDate:       tmpTrack.Header.Date,
		Pilot:       tmpTrack.Pilot,
		Glider:      tmpTrack.GliderType,
		GliderID:    tmpTrack.GliderID,
		TrackLenght: getTrackLenght(tmpTrack),
	}

	// Inserts the object into the database with the Insert function from main.go
	if err := IGF.Insert(track); err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}

	// The struct used to return the ID
	type ReturnId struct {
		ID string `json:"id"`
	}

	// converts the bson object to a string
	AsString := track.ID.Hex()

	// Returns the ID as a json back to the user
	RID := ReturnId{AsString}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(RID)
	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func handleParaglidingAPITrackID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
	} else {
		// Base lets us get the last value of the Url, which in this case is the ID
		tmp := path.Base(r.URL.Path)

		track, err := IGF.FindOne(tmp)
		if err != nil {
			fmt.Println("Findone failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		JsonStringResponse(w, http.StatusOK, track)

	}
}

func handleParaglidingAPITrackIDField(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
	} else {
		fmt.Println("finding field")
		// First use Base to get the last value of the Url which is field.
		field := path.Base(r.URL.Path)
		// Dir returns everything in the URL, but the last value.
		// this way we can use Base again, but this time ID is the last value of the Url
		tmp := path.Dir(r.URL.Path)
		nummer := path.Base(tmp)

		fmt.Println(field)
		fmt.Println(nummer)

		track, err := IGF.FindOne(nummer)
		if err != nil {
			fmt.Println("Findone failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}

		s, err := igc.ParseLocation(track.Url)
		switch field {
		case "H_date":
			text, err := track.HDate.MarshalText()
			if err != nil {
				handleError(w, r, err, http.StatusBadRequest)
				return
			}
			w.Write(text)
		case "pilot":
			w.Write([]byte(track.Pilot))
		case "glider":
			w.Write([]byte(track.Glider))
		case "glider_id":
			w.Write([]byte(track.GliderID))
		case "track_src_url":
			w.Write([]byte(track.Url))
		case "track_length":
			w.Write([]byte(strconv.Itoa(int(getTrackLenght(s)))))
		}
	}
}

func handleParaglidingAPITickerLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
	} else {
		track, err := IGF.FindLatest()
		fmt.Println(track)
		if err != nil {
			fmt.Println("FindLatest failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		text := []byte(strconv.FormatInt(track.Timestamp, 10))
		w.Write(text)
		fmt.Println(track)
	}
}

func handleParaglidingAPITicker(w http.ResponseWriter, r *http.Request) {

}

func handleParaglidingAPITickerTimestamp(w http.ResponseWriter, r *http.Request) {

}

func handlePOSTParaglidingAPIWebhookNew(w http.ResponseWriter, r *http.Request) {

}

func handleParaglidingAPIWebhookNewID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetWebhook(w, r)
	case http.MethodDelete:
		handleDeleteWebhook(w, r)
	default:
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
	}
}

func handleGetWebhook(w http.ResponseWriter, r *http.Request) {

}

func handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {

}

func handleGetAdminApiTracksCount(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		trackCount, err := IGF.FindCount()
		if err != nil {
			fmt.Println("Count all tracks failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		countString := strconv.Itoa(trackCount)
		w.Write([]byte(countString))
	} else {

		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
	}
}

func handleDeleteAdminApiTracks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		changeInfo, err := IGF.DeleteAll()
		if err != nil {
			fmt.Println("Delete all failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		fmt.Println(changeInfo)
		//w.Write([]byte(changeInfo))
	} else {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
	}
}
