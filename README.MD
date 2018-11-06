[![TODO board](https://imdone.io/api/1.0/projects/5ab993210647bc75fd8324d1/badge)](https://imdone.io/app#/board/ferux/addressbook)

# AddressBook

This is project for coding challenge.
The aim of the project is to show my skills in designing and implementing simple Go microservice server which should expose REST API.

## Installation

Before compile it's necessary to import additional packages.

### Downloading and compiling source code

```bash
mkdir addressbook
git clone https://github.com/ferux/addressbook
cd addressbook
go get .
go build -o addressBook .
```

To successfully run this server MongoDB should be installed.

### Server run

TODO: Rework

## Usage

After the server starts it will accept connections on the following address (by default):

```127.0.0.1:8080```

To simplify the table below, the User struct contains the following fields:

```JSON

User = {
    "id": "ID",
    "first_name": "FistName",
    "last_name": "LastName",
    "email": "Email",
    "phone": "Phone",
}

```

The following table describes available API requests that the server can process:

| Route                  | Method | Body      | Description                                                     | On Success           | On Error           |
|------------------------|--------|-----------|-----------------------------------------------------------------|----------------------|--------------------|
| /api/v1/book/          | GET    |           | Retrieves the full list of records in JSON format               | [ {User}, ...]       | {error: "Message"} |
| /api/v1/book/user      | POST   | {User}    | Creates a new user. ID field will be ignored.                   | {id: LastInsertedID} | {error: "Message"} |
| /api/v1/book/user/{id} | GET    |           | Gets information about selected user                            | {User}               | {error: "Message"} |
| /api/v1/book/user/{id} | PUT    | {UserNew} | Updates selected user. All fields should be specified except ID | {UserNew}            | {error: "Message"} |
| /api/v1/book/user/{id} | DELETE |           | Deletes selected user                                           | Status 200 OK        | {error: "Message"} |
| /api/v1/book/export    | GET    | CSV File  | Provides export Addressbook to CSV file                         | file:import.csv      | {error: "Message"} |

### Import file requirements

When uploading csv file to the server it is necessary to specify  

```Content-type: "text/csv".```  

Also there is optional header for data manipulation:  

```Append-type: clear | append | upsert```  

| Value           | Description                                                                                                     |
|-----------------|-----------------------------------------------------------------------------------------------------------------|
| clear           | Clears all records before importing rows from csv file                                                          |
| append          | Inserts only new rows. The old ones left unchanged                                                              |
| upsert          | Upserts all rows into database. It there are rows with the same ID, the imported one will overwrite the old one |
| any other value | In other ways it acts like you send <upsert> parameter                                                          |

## TODO

- [ ] Move from `mux` to `echo`
- [ ] Use `mgo.Session.Copy()` for each call
- [ ] Better error handling in api
