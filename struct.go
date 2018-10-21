package main

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type IgcFiles struct {
	Address  string
	Database string
	Username string
	Password string
}

type Track struct {
	ID          bson.ObjectId `bson:"_id"`
	Timestamp   int64         `bson:"timestamp" json:"timestamp"`
	Url         string        `bson:"track_src_url" json:"track_src_url"`
	HDate       time.Time     `bson:"H_date" json:"H_date"`
	Pilot       string        `bson:"pilot" json:"pilot"`
	Glider      string        `bson:"glider" json:"glider"`
	GliderID    string        `bson:"glider_id" json:"glider_id"`
	TrackLenght float64       `bson:"track_lenght" json:"track_lenght"`
}
