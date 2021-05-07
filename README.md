# FhaaS - File handling as a Service

**Scenario:** You are an on environment where you have to do many file handling. Activities like move, copy, retrieve content, list folder. And these activities are used by many of your services and each of these services may be written in a different language and you have to implement a class for abstract the file handling in each one of these. You want to update some of the access rules for these apps, so you have to update each class implementation.

## You problems are over!

FhaaS is a JSON based REST web service with implementation to do several typical file operations, like copy, move, delete, in addition to retrieve content, list folders and generate file digest. All of this available at a single HTTP request. Your server just need to have access to the target files/directories. You can even make an operation in a separate thread, follow your requester program and get the result of operation when it ends.

## Features

- Easy to use API: The endpoints are simple and have highly semantical contracts, even in HTTP method used for them.
- Security concern: Presents an interface for checking authorization of operations.
- Written in Go: A well-proved technology for deal with concurrent process and scalability. Never heard about it? [Take a look](https://stackoverflow.blog/2020/11/02/go-golang-learn-fast-programming-languages/).

## Examples
*Some of these may not be implemented yet*
The shown headers are required to every request

### Copying file

```http
POST /copy
Content-Type: application/json
X-FhaaS-Authorization: your-requester-token
X-FhaaS-Async: false

{"file_in": "full/path/to/src.file", "file_out": "full/path/to/dest.file"}
```

Return for success

```json
201 Created
{"message": "File copied successfully", "data": {"file_in": "full/path/to/src.file", "file_out": "full/path/to/dest.file"}}
```
The body sent is returned so you can easily revert an operation if needed

### Moving file
Obviously, move operation can be used to simply rename a file.


```http
PUT /move
Content-Type: application/json
X-FhaaS-Authorization: your-requester-token
X-FhaaS-Async: false

{"file_in": "full/path/to/src.file", "file_out": "full/path/to/dest.file"}
```

Return for failure

```json
200 OK
{"message": "File moved successfully", "data": {"file_in": "full/path/to/src.file", "file_out": "full/path/to/dest.file"}}
```

## Deleting file
An example using `async` mode. With async a `X-FhaaS-SendStatusTo` header can be set to declare to where FhaaS should send the result of an asynchronous operation. It may be omitted according to some options set when executing the service but, if so, you'll have no track of operation. You can send a `X-FhaaS-ResponseAuth` token to authorize the post request of operation status. It will be returned in post status report request.

```http
DELETE /delete
Content-Type: application/json
X-FhaaS-Authorization: your-requester-token
X-FhaaS-ResponseAuth: some-token-to-response
X-FhaaS-Async: true
X-FhaaS-SendStatusTo: http://therequester.app/api/results

{"file_in": "full/path/to/src.file", "file_out": "full/path/to/dest.file"}
```

Return for success of request

```json
202 Accepted
{"message": "File started to be deleted successfully", "data": {"file": "full/path/to/src.file"}}
```

Post for endpoint given to receive the status of operation. The endpoint must be able to deal with body content.
```http
POST /http://therequester.app/api/results
Content-Type: application/json
X-FhaaS-OperationStatus: "200"
X-FhaaS-ResponseAuth: some-token-to-response

{"message": "File deleted successfully", "data": {"file": "full/path/to/file.src"}}
```