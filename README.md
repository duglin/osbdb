[![Build Status](https://travis-ci.org/duglin/osbdb.svg?branch=master)](https://travis-ci.org/duglin/osbdb "Travis")

# Open Service Broker API Database

This repo contains a light-weight implementation of the Open Service
Broker API with an even lighter-weight DB as the service it supports.
Both of these run within a single container and everything is kept in memory.
This means this was not really for real production use, but rather for
use in demos where you want to demonstrate talking to an OSB API Broker
without having to worry about how to manage/provision the infrastructure
needed as you create Instances and Bindings.

The DockerHub image `duglin/osbdb` should contain the latest verison
of this repo, which means its ready to be used directly from there without
cloning or building anything from this repo. It listens on port 80 for
OSB API requests.

## Command Line Options

See the help text by specifiying `--help` on the command line.

Here's the latest:
```
Usage of broker:
  -a	Turn off all auth checking
  -h string
    	Host/port string to use for DBs 
  -i string
    	IP/interface to listen on (default "0.0.0.0")
  -p int
    	Listen port (default 80)
  -u string
    	Username for broker/DB admin (default "user")
  -v int
    	Verbosity level (default 3)
  -w string
    	Password for broker/DB admin (default "passw0rd")
```

## Talking to the Service Broker

By default the username and password for talking to the broker are
`user` and `passw0rd`. However, if you start the server with the `-a`
flag then all authentication is turned off and any value (or no value at all)
should work.

## Talking to the Database

The Database is just a simple key/value store.

Once a new Instance and Binding are created, the credentials in the
Binding will contain the following bits of information:
- `url` : The URL at which the DB is listening (via http).
- `user` : The username to use when talking to the DB.
- `password` : The password to use when talking to the DB.

Unless you turn off all authentication via the `-a` flag when starting the
server, the DB will accept the user/password as Basic Auth credentials
on all incoming requests.

You can talk directly to the Database using any HTTP client (even `curl`).
The basic form of the commands are:

Set a value:
```
PUT /db/5/keyName HTTP/1.1
...

data
```

For Database ID `5`, this will set key `keyName` to the value in the HTTP
body. It stores the data as a byte array so it can be any arbitrary data.

Get a value:
```
GET /db/5/keyName HTTP/1.1
```

For Database ID `5`, this will return key `keyName`'s value in the
HTTP response's body.

There are other options but those are the key ones.

There's a golang client library you can use in the `dbclient` dir/package
of this repo. See `broker_test.go` for sample code on how to use it.
