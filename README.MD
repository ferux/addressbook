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

### MongoDB requirements
MongoDB may or may not have Authentication and Authorization. 

### Go server run parameters
Server supports the following parameters:
<table>
    <tr>
        <th>Parameter</th>
        <th>Required</th>
        <th>Type</th>
        <th>Default</th>
        <th>Description</th>
    </tr>
    <tr>
        <td>-dbaddr</td>
        <td>false</td>
        <td>string</td>
        <td>127.0.0.1:27017</td>
        <td>Database address which should be in format "Address:Port"</td>
    </tr>
    <tr>
        <td>-dbname</td>
        <td>true</td>
        <td>string</td>
        <td></td>
        <td>Name of database.</td>
    </tr>
    <tr>
        <td>-dbcoll</td>
        <td>true</td>
        <td>string</td>
        <td></td>
        <td>Name of collection</td>
    </tr>
    <tr>
        <td>-dbuser</td>
        <td>false</td>
        <td>string</td>
        <td></td>
        <td>In case database has an auth mechanismm you must specify username to access the database</td>
    </tr>
    <tr>
        <td>-dbpass</td>
        <td>false</td>
        <td>string</td>
        <td></td>
        <td>In case database has an auth mechanismm you must specify password to access the database</td>
    </tr>
    <tr>
        <td>-listen</td>
        <td>false</td>
        <td>string</td>
        <td>:8080</td>
        <td>This parameter makes server to accept connections at custom location</td>
    </tr>
    <tr>
        <td>-debug</td>
        <td>false</td>
        <td>bool</td>
        <td></td>
        <td>Boolean flag which puts application into debug mode</td>
    </tr>
    <tr>
        <td>-retry</td>
        <td>false</td>
        <td>integer</td>
        <td>3</td>
        <td>Number of tries before exit if the connection to database fails</td>
    </tr>
    <tr>
        <td>-timeout</td>
        <td>false</td>
        <td>integer</td>
        <td>3</td>
        <td>Connection timeout when trying to reach database server. The number reflects the number of seconds</td>
    </tr>
</table>

### Server run
The simplest way to run server is to install MongoDB on localhost and make it listen to default port.  
After this run the following command:
```bash
./addressBook -dbname "Database Name" -dbcoll "Collection Name"
```

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
<table>
    <tr>
        <th>Route</th>
        <th>Method</th>
        <th>Body</th>
        <th>Description</th>
        <th>On Success</th>
        <th>On Error</th>
    </tr>
    <tr>
        <td>/api/v1/addressbook/</td>
        <td>GET</td>
        <td></td>
        <td>Retrieves the full list of records in JSON format</td>
        <td>[ {User}, ...]</td>
        <td>{error: "Message"}</td>
    </tr>
    <tr>
        <td>/api/v1/addressbook/user</td>
        <td>POST</td>
        <td>{User}</td>
        <td>Creates a new user. ID field will be ignored.</td>
        <td>{id: LastInsertedID}</td>
        <td>{error: "Message"}</td>
    </tr>
    <tr>
        <td>/api/v1/addressbook/user/{id}</td>
        <td>GET</td>
        <td></td>
        <td>Gets information about selected user</td>
        <td>{User}</td>
        <td>{error: "Message"}</td>
    </tr>
    <tr>
        <td>/api/v1/addressbook/user/{id}</td>
        <td>PUT</td>
        <td>{UserNew}</td>
        <td>Updates selected user. All fields should be specified except ID</td>
        <td>{UserNew}</td>
        <td>{error: "Message"}</td>
    </tr>
    <tr>
        <td>/api/v1/addressbook/user/{id}</td>
        <td>DELETE</td>
        <td></td>
        <td>Deletes selected user</td>
        <td>It returns STATUS 200 OK</td>
        <td>{error: "Message}</td>
    </tr>
        <tr>
        <td>/api/v1/addressbook/export</td>
        <td>GET</td>
        <td></td>
        <td>Provides export Addressbook to CSV file</td>
        <td>file:import.csv</td>
        <td>{error: "Message}</td>
    </tr>
        <tr>
        <td>/api/v1/addressbook/import</td>
        <td>PUT</td>
        <td>CSV file</td>
        <td>Imports records into Addressbook. This method's headers should be specified. See below for more info.
        </td>
        <td>It returns STATUS 200 OK</td>
        <td>{error: "Message}</td>
    </tr>
</table>

### Import file requirements
When uploading csv file to the server it is necessary to specify  
```Content-type: "text/csv".```  
  
Also there is optional header for data manipulation:  
```Append-type: clear | append | upsert```  
<table>
    <tr><th>Value</th><th>Description</th></tr>
    <tr>
        <td>clear</td>
        <td>Clears all records before importing rows from csv file</td>
    </tr>
    <tr>
        <td>append</td>
        <td>Inserts only new rows. The old ones left unchanged</td>
    </tr>
    <tr>
        <td>upsert</td>
        <td>Upserts all rows into database. It there are rows with the same ID, the imported one will overwrite the old one</td>
    </tr>
    <tr>
        <td>(any other value)</td>
        <td>In other ways it acts like you send "upsert" parameter</td>
    </tr>
</table>