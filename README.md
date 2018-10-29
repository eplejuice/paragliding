# Assignment 2: IGC track viewer extended

##### Martin Brådalen  |  martbraa@stud.ntnu.no  |  Studnr: 473145

## Assignment description
Develop an online service that will allow users to browse information about IGC files using a NoSQL Database

#### Assignment link: 
    http://prod3.imt.hig.no/teaching/imt2681-2018/wikis/assignment-2
#### Heroku link 
    https://pure-stream-73485.herokuapp.com/
    
#### External dependencies
    https://github.com/marni/goigc
    https://github.com/gorilla/mux
    gopkg.in/mgo.v2
    gopkg.in/mgo.v2/bson
    
#### Why use [mgo](https://github.com/globalsign/mgo) instead of the [official MongoDB Go driver](https://github.com/mongodb/mongo-go-driver) ?
- In my opinion, easier installation and better documentation.


### Quality
- [x] Golint
- [x] GoVet

## Setup for testing
Make an .env file containing
    
     MONGO_ADDRESS= <address to the mongodb>
     MONGO_USER= <username>
     MONGO_PASSWORD= <password>
     MONGO_DATABASE=paragliding
     MONGO_PORT=8080

## Test and expected results

### Tested using [Postman](https://www.getpostman.com/)
### MongoDB hosted on [mLab](https://mlab.com/home)

### Additional info
The clocktrigger function is deployed on a VM in skyhigh Openstack with floating IP: 10.212.137.65

For the prevention of duplicate IDs, i used mutex to lock and unlock critical sector while posting a new track.

The entire API is not deployed on AWS as a cloud function, mainly because we have yet to gain access to AWS
