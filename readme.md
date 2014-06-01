# trending-torrents
prototype of a data analyzing platform tracking torrents

# technical details
The platform is experimenal, should be based on:  

  * ~~Golang~~ - backend, spider
  * node.js - backend, spider
  * ~~OrientDB - database, rest API, map/reduce~~
  * CouchDB - database, rest API, map/reduce
  * frontend:
    * angular.js - frontend webapp
    * can.js + d3 + nvd3 - alternatively, if angular is not needed entirely 

## couchdb designs

Import JSON from the repository

```bash
$ curl -v -X PUT -H "Content-Type: application/json" localhost:5984/trendingtorrents/_design/convert -d @couchdb.json
```

query them:

```bash
$ curl -v -X GET localhost:5984/trendingtorrents/_design/convert/_view/getFetches
$ curl -v -X GET localhost:5984/trendingtorrents/_design/convert/_list/listFetches/getFetches
```

# License
[MIT](LICENSE) license for now