package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/marni/goigc"
	"gopkg.in/mgo.v2/bson"
)

// A mutex to lock/unlock the timestamp as a critical sector, to avoid duplicates
var mutex = &sync.Mutex{}

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
	regHandleParaglidingAPIWebhookNewID, err := regexp.Compile("^/paragliding/api/webhook/new_track/[a-z0-9]+/?$")

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
		http.Error(w, fmt.Sprintf("%s, %s", http.StatusText(status), err), status)
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
	// Checks if the method is GET, otherwise returns error
	if r.Method != http.MethodGet {
		status := http.StatusMethodNotAllowed
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
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	}
}

// Returns an array containing the IDs of all stored tracks in the database
func handleGetParaglidingAPITrack(w http.ResponseWriter, r *http.Request) {
	// Calls the fundall function from main which returns all object from the db in a slice
	tracks, err := IGF.FindAll()
	if err != nil {
		handleError(w, r, err, http.StatusBadRequest)
	}
	// makes a new slice to put the IDs in
	var trackks []string
	// Loops throught the slice with the objects from the db,
	// and puts their IDs into the new slice
	for i := 0; i < len(tracks); i++ {
		trackks = append(trackks, (tracks[i].ID.Hex()))
	}
	// Sends the slice to be converted and responded as json
	JsonStringResponse(w, http.StatusOK, trackks)
}

// Lets a user post a new track into the database with a url to an igcfile
func handlePostParaglidingAPITrack(w http.ResponseWriter, r *http.Request) {
	// Locks th critical sector before creating the unique timestamp
	// So only one can be created at a time
	mutex.Lock()
	// Creating a unique monotone ID by converting a timestamp to milliseconds
	Uniq := time.Now().UnixNano() / int64(time.Millisecond)
	// Unlocks the critical sector again after creation
	mutex.Unlock()

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
	//Notifies all registered webhooks that changes has been made
	invokeWebhooks(w, r)

}

func handleParaglidingAPITrackID(w http.ResponseWriter, r *http.Request) {
	// checks if the method is actually Get
	if r.Method != http.MethodGet {
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	} else {
		// Base lets us get the last value of the Url, which in this case is the ID
		tmp := path.Base(r.URL.Path)
		if !bson.IsObjectIdHex(tmp) {
			status := http.StatusBadRequest
			http.Error(w, http.StatusText(status), status)
			return
		}

		// Calls the findOne function which returns the object based in the ID in tmp
		track, err := IGF.FindOne(tmp)
		if err != nil {
			fmt.Println("Findone failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		// Converts is to json and responds
		JsonStringResponse(w, http.StatusOK, track)

	}
}

func handleParaglidingAPITrackIDField(w http.ResponseWriter, r *http.Request) {
	// Checks if the method is actually GET
	if r.Method != http.MethodGet {
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	} else {
		fmt.Println("finding field")
		// First use Base to get the last value of the Url which is field.
		field := path.Base(r.URL.Path)
		// Dir returns everything in the URL, but the last value.
		// this way we can use Base again, but this time ID is the last value of the Url
		tmp := path.Dir(r.URL.Path)
		nummer := path.Base(tmp)

		if !bson.IsObjectIdHex(tmp) {
			status := http.StatusBadRequest
			http.Error(w, http.StatusText(status), status)
			return
		}

		fmt.Println(field)
		fmt.Println(nummer)
		// Uses the findOne function, with the nummber var to find the right object.
		track, err := IGF.FindOne(nummer)
		if err != nil {
			fmt.Println("Findone failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		// Checks if the url in the object is an actual URL leading to and IGC file
		s, err := igc.ParseLocation(track.Url)
		// switches on the field parameter, and writes out the right data from the object
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
	// Checks if the method is GET
	if r.Method != http.MethodGet {
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	} else {
		// Calls the FindLatest function from main, returns the latest object into track
		track, err := IGF.FindLatest()
		fmt.Println(track)
		if err != nil {
			fmt.Println("FindLatest failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		// Converts and writes the timestamp from the object as response.
		// the 10 parameter is to convert to desimal, alternatively 16 for hex
		text := []byte(strconv.FormatInt(track.Timestamp, 10))
		w.Write(text)
		fmt.Println(track)
	}
}

//Returns info about the latest 5 added tracks
func handleParaglidingAPITicker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	} else {
		// Starts the timer to calculate processing time
		start := time.Now()
		type rStruct struct {
			T_latest   int64         `json:"t_latest"`
			T_start    int64         `json:"t_start"`
			T_stop     int64         `json:"t_stop"`
			Tracks     []string      `json:"tracks"`
			Processing time.Duration `json:"processing"`
		}
		// finds the latest added track to the database
		trackLatest, err := IGF.FindLatest()
		if err != nil {
			fmt.Println("FindLatest failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		// Finds the 5 oldest objects in the database
		trackStart, err := IGF.FindOldest()
		if err != nil {
			fmt.Println("FindOldest failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		// puts the timestamps of the returnes object into a slice
		var trackks []int64
		for i := 0; i < len(trackStart); i++ {
			trackks = append(trackks, (trackStart[i].Timestamp))
		}

		// puts the IDs of the returned objects into a slice
		var trackksID []string
		for i := 0; i < len(trackStart); i++ {
			trackksID = append(trackksID, (trackStart[i].ID.Hex()))
		}

		// checks how many object where actually returned
		sliceLength := len(trackks)

		// Puts the right values into the struct
		returnSt := rStruct{
			T_latest: trackLatest.Timestamp,
			// T_start is always the first object in the array
			T_start: trackks[0],
			// T_stop wil always be the last object int the array
			// using length -1, bcus the array starts at 0
			T_stop:     trackks[sliceLength-1],
			Tracks:     trackksID,
			Processing: time.Since(start) / time.Millisecond,
		}

		// Returns the content as json
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(returnSt)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return

		}
	}
}

func handleParaglidingAPITickerTimestamp(w http.ResponseWriter, r *http.Request) {
	// checks if the method is actually Get
	if r.Method != http.MethodGet {
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	} else {

		start := time.Now()
		type rStruct struct {
			T_latest   int64         `json:"t_latest"`
			T_start    int64         `json:"t_start"`
			T_stop     int64         `json:"t_stop"`
			Tracks     []string      `json:"tracks"`
			Processing time.Duration `json:"processing"`
		}
		trackLatest, err := IGF.FindLatest()
		if err != nil {
			fmt.Println("FindLatest failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}

		// Base lets us get the last value of the Url
		// which in this case is the timestamp
		tmp := path.Base(r.URL.Path)
		tmpInt, _ := strconv.Atoi(tmp)
		trackLatestArr, err := IGF.FindOldestById(tmpInt)

		// The rest of the code is nearly identical to the function above
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
			T_latest:   trackLatest.Timestamp,
			T_start:    trackks[0],
			T_stop:     trackks[sliceLength-1],
			Tracks:     trackksID,
			Processing: time.Since(start) / time.Millisecond,
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(returnSt)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
	}
}

func handlePOSTParaglidingAPIWebhookNew(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New webhook")
	// checks if the method is actually POST
	if r.Method != http.MethodPost {
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	} else {
		type getParams struct {
			WebHookURL      string `json:"webhookURL"`
			MinTriggerValue int    `json:"minTriggerValue"`
		}

		// Creates a struct to decode the the json object into
		params := getParams{}

		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}

		// Finds the latest inserted object in the database
		latestKnown, err := IGF.FindLatest()
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}

		// Creates the actual webhook and sends it to be inserted
		webhook := Webhooks{
			ID:               bson.NewObjectId(),
			WebhookURL:       params.WebHookURL,
			MinTriggerValue:  params.MinTriggerValue,
			LatestKnownTrack: latestKnown.Timestamp,
		}
		err = IGF.NewWebHook(webhook)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
		}
		w.Write([]byte(webhook.ID.Hex()))
	}
}

// This function redirects the user based on the method when sending the request
func handleParaglidingAPIWebhookNewID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetWebhook(w, r)
	case http.MethodDelete:
		handleDeleteWebhook(w, r)
	default:
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	}
}

// Returns the info about a webhook based on given ID
func handleGetWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	} else {
		// Base lets us get the last value of the Url, which in this case is the ID
		tmp := path.Base(r.URL.Path)

		if !bson.IsObjectIdHex(tmp) {
			status := http.StatusBadRequest
			http.Error(w, http.StatusText(status), status)
			return
		}
		// Calls the findOne function which returns the webhook based in the ID in tmp
		webhook, err := IGF.FindOneWebhook(tmp)
		if err != nil {
			fmt.Println("Findone webhook failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		type returnVal struct {
			WebhookURL      string `json:"webhookURL"`
			MinTriggerValue int    `json:"minTriggerValue"`
		}

		RV := returnVal{
			WebhookURL:      webhook.WebhookURL,
			MinTriggerValue: webhook.MinTriggerValue,
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(RV)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		// Converts is to json and responds

	}
}

// Nearly identical to the function above, exepct it also deletes the object after
// returning the values. Could perhaps be made into one function.
// Ran out of time while thinking about it.
func handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
	} else {
		// Base lets us get the last value of the Url, which in this case is the ID
		tmp := path.Base(r.URL.Path)

		if !bson.IsObjectIdHex(tmp) {
			status := http.StatusBadRequest
			http.Error(w, http.StatusText(status), status)
			return

		}
		// This function retuns and deletes the function
		webhook, err := IGF.DeleteOneHook(tmp)
		if err != nil {
			fmt.Println("Find and delete failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}

		type returnVal struct {
			WebhookURL      string `json:"webhookURL"`
			MinTriggerValue int    `json:"minTriggerValue"`
		}

		RV := returnVal{
			WebhookURL:      webhook.WebhookURL,
			MinTriggerValue: webhook.MinTriggerValue,
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(RV)
		if err != nil {
			handleError(w, r, err, http.StatusBadRequest)
			return
		}

	}
}

// Returns a count of how many tracks exists in the database
func handleGetAdminApiTracksCount(w http.ResponseWriter, r *http.Request) {
	// Checks if the method is GET, otherwise responds with error
	if r.Method == http.MethodGet {
		// calls the fund function which returns number of tracks into trackCount variable
		trackCount, err := IGF.FindCount()
		if err != nil {
			fmt.Println("Count all tracks failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		// converts the int returned to a string, and then converts it to byte to write.
		countString := strconv.Itoa(trackCount)
		w.Write([]byte(countString))
	} else {

		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
	}
}

// Deletes all tracks from the database
func handleDeleteAdminApiTracks(w http.ResponseWriter, r *http.Request) {
	//Checks if the method is delete
	if r.Method == http.MethodDelete {
		// calls the delete all function from main
		changeInfo, err := IGF.DeleteAll()
		// checks if the function executed correctly
		if err != nil {
			fmt.Println("Delete all failed")
			handleError(w, r, err, http.StatusBadRequest)
			return
		}
		// prints the output to console
		fmt.Println(changeInfo)
		//w.Write([]byte(changeInfo))
	} else {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
	}
}
