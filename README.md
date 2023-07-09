# Go API Template

The purpose of this template is to provide an easy way to start a new project. It includes token based http authentication, token authenticated websockets, and a support ticket system. 

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

```
go1.19.3
```

### Installing

Clone the repository
```
git clone git@github.com:carlos-nunez/go-api-template.git
```

Install dependencies
```
go get
```

Create a .env files with the following environment variables. Use "api" to use the user WS token as authentication, or anything else to check the servers database for the provided connection token.
```
MONGO_URI="" // your MONGO connection URI
PRODUCT="" // the name of your product, which will be your database name
SIGNING_SECRET="" // token auth signing secret
DEPLOY_URL = "" // the url of your deploy server if using this in conjunction with the deploy template
DEPLOY_KEY = "" // your deployment key, matching the one on your deploy server
uuid="" // a UUID to identify this environment

```

## Usage

Run the application using

```
go server.go
```

Build the docker image using
```
docker build -t name:tag .
```

Now use Postman to setup everything you need.

Create a user.
```
POST:
http://localhost:5000/api/users

Payload:
{"email": "testuser@test.com", "password": "1234", "full_name": "Test User"}
```
Save the "token" from the response, and add it as a bearer token on Postman.


Make a post request to make a new ws server. Use the same uuid as the one in your .env file.
```
POST: http://localhost:5000/api/servers

Payload:
{
    "name": "test",
    "uuid": "test",
    "memory": "512",
    "cpu": "512"
}
```

Save the response token to connect to the server. Open up home.html or any websocket client Change line 31, the const token = to your token. In the following examples, change {yourservertoken} to your actual token without the curly braces.
```
const token = "{yourservertoken}"
```

The url to connect to the server is in the following format. Change to wss when available.
```
ws://" + document.location.host + "/ws?token={yourservertoken}
```

Go to localhost:5000

Join a room using the following message
```
{"type": "subscribe", "roomID": "room1", "token": "{yourservertoken}"}
```
You should now be able to send messages. Try it out with two tabs!