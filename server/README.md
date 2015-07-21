## Signup

curl -d '{"email": "jamie@tidepool.org","name":"Jamie Bate", "password": "admin"}' -H "Content-Type:application/json" http://localhost:8090/signup

```
{
  "id": "9339251d-e8a7-4194-89c5-a1cb2edc0519",
  "name": "Jamie Bate",
  "email": "jamie@tidepool.org"
}
```

## Login

curl -u jamie@tidepool.org:admin -i -H 'Content-Type: application/json' -d '' http://localhost:8090/loginHTTP/1.1 200 OK

```
x-fantail-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IjkzMzkyNTFkLWU4YTctNDE5NC04OWM1LWExY2IyZWRjMDUxOSIsImV4cCI6MTQzNjA4OTQ3N30.iVkSBQUic-kI5fC7bPogYFjKjnihRfmYRzEe_F6vDkk
X-Powered-By: go-json-rest
Date: Thu, 02 Jul 2015 09:44:37 GMT
Content-Length: 0
Content-Type: text/plain; charset=utf-8
```

## POST smbgs

curl -i -H "x-fantail-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IjkzMzkyNTFkLWU4YTctNDE5NC04OWM1LWExY2IyZWRjMDUxOSIsImV4cCI6MTQzNjA4OTgzOH0.waKAYOHeHysXMGD6_4t9duq1yLsVT7PKJ-mRu3ME1os" -H 'Content-Type: application/json' -d '{"value": 4.5, "time": "2015-05-16T10:42:57.539Z" } {"value": 9.5, "time": "2015-05-16T10:42:57.539Z" }' http://127.0.0.1:8090/data/213/smbgs

```
HTTP/1.1 201 Created
Content-Type: application/json
Vary: Accept-Encoding
X-Powered-By: go-json-rest
Date: Thu, 02 Jul 2015 09:55:08 GMT
Content-Length: 353

[{"id":"914f4d3d-7c6a-4cf4-aab1-901bcfab6cdc","type":"smbg","time":"2015-05-16T10:42:57.539Z","createdAt":"2015-07-02T09:55:08Z","datumVersion":0,"schemaVersion":0,"value":4.5},{"id":"1823d8fe-cf60-4a26-be69-cb6090491e3b","type":"smbg","time":"2015-05-16T10:42:57.539Z","createdAt":"2015-07-02T09:55:08Z","datumVersion":0,"schemaVersion":0,"value":9.5}]
```


## GET smbgs

curl -i -H "x-fantail-token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJRCI6IjkzMzkyNTFkLWU4YTctNDE5NC04OWM1LWExY2IyZWRjMDUxOSIsImV4cCI6MTQzNjA4OTgzOH0.waKAYOHeHysXMGD6_4t9duq1yLsVT7PKJ-mRu3ME1os" -H 'Content-Type: application/json'  http://127.0.0.1:8090/data/213/smbgs

```
HTTP/1.1 200 OK
Content-Type: application/json
Vary: Accept-Encoding
X-Powered-By: go-json-rest
Date: Thu, 02 Jul 2015 09:58:38 GMT
Transfer-Encoding: chunked

[{"id":"914f4d3d-7c6a-4cf4-aab1-901bcfab6cdc","type":"smbg","time":"2015-05-16T10:42:57.539Z","createdAt":"2015-07-02T09:55:08Z","datumVersion":0,"schemaVersion":0,"value":4.5},{"id":"1823d8fe-cf60-4a26-be69-cb6090491e3b","type":"smbg","time":"2015-05-16T10:42:57.539Z","createdAt":"2015-07-02T09:55:08Z","datumVersion":0,"schemaVersion":0,"value":9.5}]
```