package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var startTime = time.Now()
var db *mgo.Database

const (
	COLLECTION = "tracks"
	WEBHOOKS   = "webhooks"
)

var IGF = IgcFiles{
	Address:  os.Getenv("MONGO_ADDRESS"),
	Database: os.Getenv("MONGO_DATABASE"),
	Username: os.Getenv("MONGO_USER"),
	Password: os.Getenv("MONGO_PASSWORD"),
}

func main() {
	fmt.Println("Starting main")

	// Connects to the databse
	fmt.Println("Connecting to database")
	IGF.Connect()
	fmt.Println("Connection success")
	// Sends every request to the router function with Regex.
	http.HandleFunc("/", handleRouter)

	port := os.Getenv("PORT")
	//Listens to the Url given by heroku
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		// If the Url is wrong the program shuts down immediately.
		fmt.Printf("Listen and serve failed %s", err)
		log.Fatal(err)
	}

}

func (m *IgcFiles) Connect() {
	session := &mgo.DialInfo{
		Addrs:    []string{m.Address},
		Timeout:  60 * time.Second,
		Database: m.Database,
		Username: m.Username,
		Password: m.Password,
	}

	connection, err := mgo.DialWithInfo(session)
	if err != nil {
		log.Fatal(err)
	}
	db = connection.DB(m.Database)
}

// This function is for inserting a document into the databse
func (m *IgcFiles) Insert(track Track) error {
	fmt.Println("Trying to insert into the db")
	// Inserts the track parameter into the right collection in the database.
	// returns error if failed
	err := db.C(COLLECTION).Insert(&track)
	return err
}

// This function returns all documents from a collection in the database
func (m *IgcFiles) FindAll() ([]Track, error) {
	fmt.Println("Trying to find all")
	var tracks []Track
	// Using the nil parameter in find gets all tracks
	err := db.C(COLLECTION).Find(nil).All(&tracks)
	return tracks, err
}

// This function finds one document in the collection based in the id parameter
func (m *IgcFiles) FindOne(id string) (Track, error) {
	fmt.Println("Trying to find one by id")
	var track Track
	// Using bson.ObjectIdHex to convert the ID send to a hex,
	// then compares it to the hexadesimal IDs generated by mongodb
	err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&track)
	return track, err
}

// This function returns the latest inserted document in the database
func (m *IgcFiles) FindLatest() (Track, error) {
	var track Track
	// Returns the first object of all documents sorted by "_id"
	err := db.C(COLLECTION).Find(nil).Sort("-_id").One(&track)
	return track, err
}

// This function returns an int with the count of how many documents
// are in a collection in the database
func (m *IgcFiles) FindCount() (int, error) {
	trackCount, err := db.C(COLLECTION).Count()
	return trackCount, err
}

// This function deletes all documents in a collection
func (m *IgcFiles) DeleteAll() (*mgo.ChangeInfo, error) {
	rem, err := db.C(COLLECTION).RemoveAll(nil)
	return rem, err
}

func (m *IgcFiles) FindOldest() ([]Track, error) {
	var tracks []Track
	// Gets the first 5 object when documents are sorted reverse order by giving "-" to _id
	err := db.C(COLLECTION).Find(nil).Sort("timestamp").Limit(5).All(&tracks)
	return tracks, err
}

func (m *IgcFiles) FindOldestById(id int) ([]Track, error) {
	var tracks []Track
	var startPoint Track
	err := db.C(COLLECTION).Find(bson.M{"timestamp": id}).One(&startPoint)
	err = db.C(COLLECTION).Find(bson.M{"timestamp": bson.M{"$gt": startPoint.Timestamp}}).Limit(5).All(&tracks)
	return tracks, err
}

func (m *IgcFiles) NewWebHook(webhook Webhooks) error {
	fmt.Println("Trying to insert new webhook into the db")
	// Inserts the webhook into the right collection in the database.
	// returns error if failed
	err := db.C(WEBHOOKS).Insert(&webhook)
	return err
}

func (m *IgcFiles) getAllWebhooks() ([]Webhooks, error) {
	fmt.Println("Trying to find all")
	var webhook []Webhooks
	// Using the nil parameter in find gets all tracks
	err := db.C(WEBHOOKS).Find(nil).All(&webhook)
	return webhook, err
}

// Returns the oldest webhook in the collection
func (m *IgcFiles) FindOldestByIdWebhook(id int) ([]Track, error) {
	var tracks []Track
	var startPoint Track
	// Using bson to match the id
	err := db.C(COLLECTION).Find(bson.M{"timestamp": id}).One(&startPoint)
	// using bson with the parameter $gt to get all with timestamp greater than given
	err = db.C(COLLECTION).Find(bson.M{"timestamp": bson.M{"$gt": startPoint.Timestamp}}).All(&tracks)
	return tracks, err
}

// This function finds one document in the collection based in the id parameter
func (m *IgcFiles) FindOneWebhook(id string) (Webhooks, error) {
	fmt.Println("Trying to find one webhook by id")
	var webhook Webhooks
	// Using bson.ObjectIdHex to convert the ID send to a hex,
	// then compares it to the hexadesimal IDs generated by mongodb
	err := db.C(WEBHOOKS).FindId(bson.ObjectIdHex(id)).One(&webhook)
	return webhook, err
}

//returns the values from one webhook of given ID, then deletes it
func (m *IgcFiles) DeleteOneHook(id string) (Webhooks, error) {
	fmt.Println("Trying to delete one webhook by id")
	var webhook Webhooks
	// Using bson.ObjectIdHex to convert the ID send to a hex,
	// then compares it to the hexadesimal IDs generated by mongodb
	err := db.C(WEBHOOKS).FindId(bson.ObjectIdHex(id)).One(&webhook)
	err = db.C(WEBHOOKS).RemoveId(bson.ObjectIdHex(id))
	return webhook, err
}
