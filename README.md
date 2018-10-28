# Assignment 2: IGC track viewer extended

##### Martin Brådalen  |  martbraa@stud.ntnu.no  |  Studnr: 473145

## Assignment description
Develop an online service that will allow users to browse information about IGC files using a NoSQL Database

#### Assignment link: 
    http://prod3.imt.hig.no/teaching/imt2681-2018/wikis/assignment-2
#### Heroku link 
    tbd
    
#### External dependencies
    https://github.com/marni/goigc
    https://github.com/gorilla/mux
    gopkg.in/mgo.v2
    gopkg.in/mgo.v2/bson
 
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
### MongoDB hosten on [mLab](https://mlab.com/home)
