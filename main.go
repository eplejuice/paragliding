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
	IGF.Connect()

	// Sends every request to the router function with Regex.
	http.HandleFunc("/", handleRouter)

	//Listens to the Url given by heroku
	if err := http.ListenAndServe(":"+os.Getenv("MONGO_PORT"), nil); err != nil {
		// If the Url is wrong the program shuts down immediately.
		log.Fatal(err)
	}

}

func (m *IgcFiles) Connect() {
	fmt.Println("Connecting to database")
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

func (m *IgcFiles) Insert(track Track) error {
	fmt.Println("Trying to insert into the db")
	err := db.C(COLLECTION).Insert(&track)
	return err
}

func (m *IgcFiles) FindAll() ([]Track, error) {
	fmt.Println("Trying to find all")
	var tracks []Track
	err := db.C(COLLECTION).Find(nil).All(&tracks)
	return tracks, err
}

func (m *IgcFiles) FindOne(id string) (Track, error) {
	fmt.Println("Trying to find one by id")
	var track Track
	err := db.C(COLLECTION).FindId(bson.ObjectIdHex(id)).One(&track)
	return track, err
}

func (m *IgcFiles) FindLatest() (Track, error) {
	var track Track
	err := db.C(COLLECTION).Find(bson.M{"$natural": -1}).One(&track)
	return track, err
}
