`flow-beam` is an HTTP proxy service that exposes some functionaly of
the [flow-go-sdk](https://github.com/onflow/flow-go-sdk) through an HTTP API. You can use it to get the latest sealed
block height, fetch Cadence events from any block height that has an access node defined in the supplied JSON file, or
execute an arbitrary Cadence script at the latest block height. It was originally built to provide consistent access to
historical Flow Cadence events to services that were not written in Go.

---

### Docker Quickstart

A docker image is provided if you just want to easily launch a prebuilt image of `flow-beam`. We do our best to keep the
embedded list of access nodes up-to-date, but you can always build your own version using a custom access-nodes.json
file using the [development instructions](#development) below.

The Docker Hub repository can be found at [rayvinflow/flow-beam](https://hub.docker.com/r/rayvinflow/flow-beam). The current stable version uses the `1.x` tag.

Launch docker image:

```shell
docker run rayvinflow/flow-beam:1.x
```

---

### API

You can access the HTTP server using the `LISTEN_PORT` that was specified in the `.env` file. If you kept the default
port of `8080`, then you should be able to make a GET request to the following URL:

`https://localhost:8080/[method]`

The following methods are available:

[Get Latest Block Height](#getting-latest-block-height)

[Query Events](#querying-events)

[Execute Script](#executing-a-script)

### Getting The Latest Block Height

GET `/latest-block-height`

If the request is successful, the server will return a 200 status code with a JSON object in the body matching
the `Response` schema defined below:

```typescript
type Response = {
    "LatestBlockHeight": number // height of the latest sealed block
    "ApiCalls": number // number of API calls made to Flow
}
```

Example Response:

```json
{
  "LatestBlockHeight": 19896000,
  "ApiCalls": 1
}
```

If the request fails, the server will return a 500 status code with a JSON object in the body matching
the `ErrorResponse` schema defined below:

```typescript
type ErrorResponse = {
    "ApiCalls": number // number of API calls made to Flow
    "Error": string // string representation of the error
}
```

Example error response:

```json
{
  "ApiCalls": 1,
  "Error": "rpc error: code = OutOfRange desc = start height 19897000 is greater than the last sealed block height 19896844"
}
```

### Querying Events

GET `/events?eventType=[EVENTTYPE]&start=[START]&end=[END]`

The following querystring parameters are required:

|Parameter|Description|
|:---|:---|
|eventType|Cadence event type of events to fetch|
|start|Starting block height of events to fetch|
|end|Ending block height of events to fetch|

If the request is successful, the server will return a 200 status code with a JSON object in the body matching
the `Response` schema defined below:

```typescript
type Response = {
    "Blocks": Block[] // all blocks returned from this request
    "ApiCalls": number // number of API calls made to Flow
}

type Block = {
    "Id": string
    "Height": number
    "Timestamp": number
    "Events": Event[]
}

type Event = {
    "Type": string
    "TransactionId": string
    "TransactionIndex": number
    "EventIndex": number
    "Payload": string
}
```

Example Response:

```json
{
  "Blocks": [
    {
      "Id": "64cf1ea33398c12a67d2fc9da7d8dd739705b17c003eb65609172118abc05aa8",
      "Height": 19896000,
      "Timestamp": 1635753017,
      "Events": [
        {
          "Type": "A.c1e4f4f4c4257510.TopShotMarketV3.MomentListed",
          "TransactionId": "41e0105d8b69a21ca07da64e7b9fa9ccd42123b51475b10b278b1b7f4445d535",
          "TransactionIndex": 1,
          "EventIndex": 0,
          "Payload": "eyJ0eXBlIjoiRXZlbnQiLCJ2YWx1ZSI6eyJpZCI6IkEuYzFlNGY0ZjRjNDI1NzUxMC5Ub3BTaG90TWFya2V0VjMuTW9tZW50TGlzdGVkIiwiZmllbGRzIjpbeyJuYW1lIjoiaWQiLCJ2YWx1ZSI6eyJ0eXBlIjoiVUludDY0IiwidmFsdWUiOiIxNzczODIyNyJ9fSx7Im5hbWUiOiJwcmljZSIsInZhbHVlIjp7InR5cGUiOiJVRml4NjQiLCJ2YWx1ZSI6Ijg2LjAwMDAwMDAwIn19LHsibmFtZSI6InNlbGxlciIsInZhbHVlIjp7InR5cGUiOiJPcHRpb25hbCIsInZhbHVlIjp7InR5cGUiOiJBZGRyZXNzIiwidmFsdWUiOiIweDFiNGJkZGFlMTQzMGMwYjgifX19XX19Cg=="
        },
        {
          "Type": "A.c1e4f4f4c4257510.TopShotMarketV3.MomentListed",
          "TransactionId": "beccf8e3e0a1193d4412563a7bd8be9cf759a90d5af80cd19417b9bdc268865c",
          "TransactionIndex": 2,
          "EventIndex": 0,
          "Payload": "eyJ0eXBlIjoiRXZlbnQiLCJ2YWx1ZSI6eyJpZCI6IkEuYzFlNGY0ZjRjNDI1NzUxMC5Ub3BTaG90TWFya2V0VjMuTW9tZW50TGlzdGVkIiwiZmllbGRzIjpbeyJuYW1lIjoiaWQiLCJ2YWx1ZSI6eyJ0eXBlIjoiVUludDY0IiwidmFsdWUiOiI0NzQzNzg4In19LHsibmFtZSI6InByaWNlIiwidmFsdWUiOnsidHlwZSI6IlVGaXg2NCIsInZhbHVlIjoiNC4wMDAwMDAwMCJ9fSx7Im5hbWUiOiJzZWxsZXIiLCJ2YWx1ZSI6eyJ0eXBlIjoiT3B0aW9uYWwiLCJ2YWx1ZSI6eyJ0eXBlIjoiQWRkcmVzcyIsInZhbHVlIjoiMHg3MzUwMjA5ZjgzMjc0MmNkIn19fV19fQo="
        }
      ]
    },
    {
      "Id": "f14139db1780fac5ba6fce1d9657025fae965456cd6f5a345c79eb0b85a9996d",
      "Height": 19896001,
      "Timestamp": 1635753019,
      "Events": []
    }
  ],
  "ApiCalls": 1
}
```

The `Payload` property on the `Event` object is a base64 encoded representation of the raw bytes of the Cadence event
payload. The format is defined in
the [JSON-Cadence Data Interchance Format Specs](https://docs.onflow.org/cadence/json-cadence-spec/). You can use
the [flow-json-cadence-decoder](https://github.com/rayvin-flow/flow-json-cadence-decoder) library to decode it to a
plain JSON object.

If the request fails, the server will return a 500 status code with a JSON object in the body matching
the `ErrorResponse` schema defined below:

```typescript
type ErrorResponse = {
    "ApiCalls": number // number of API calls made to Flow
    "Error": string // string representation of the error
}
```

Example error response:

```json
{
  "ApiCalls": 1,
  "Error": "rpc error: code = OutOfRange desc = start height 19897000 is greater than the last sealed block height 19896844"
}
```

### Executing A Script

POST `/execute-script`

Make a POST request to this endpoint to execute a script at the latest block. The body of the request should be a JSON
payload with the following schema:

```typescript
type Request = {
    Script: string
    Arguments: CadenceValue[]
}
```

The `Arguments` property should be an array of values following
the [JSON-Cadence Data Interchange Format](https://docs.onflow.org/cadence/json-cadence-spec/).

Example Request:

```json
{
  "Script": "pub fun main(greeting: String, who: String): String {\nreturn greeting.concat(\" \").concat(who)\n}",
  "Arguments": [
    {
      "type": "String",
      "value": "Hello"
    },
    {
      "type": "String",
      "value": "World"
    }
  ]
}
```

If the request is successful, the server will return a 200 status code with a JSON object in the body matching
the `Response` schema defined below:

```typescript
type Response = {
    "Result": string   // all blocks returned from this request
    "ApiCalls": number // number of API calls made to Flow
}
```

Example Response:

```json
{
  "Result": "eyJ0eXBlIjoiRXZlbnQiLCJ2YWx1ZSI6eyJpZCI6IkEuYzFlNGY0ZjRjNDI1NzUxMC5Ub3BTaG90TWFya2V0VjMuTW9tZW50TGlzdGVkIiwiZmllbGRzIjpbeyJuYW1lIjoiaWQiLCJ2YWx1ZSI6eyJ0eXBlIjoiVUludDY0IiwidmFsdWUiOiI0NzQzNzg4In19LHsibmFtZSI6InByaWNlIiwidmFsdWUiOnsidHlwZSI6IlVGaXg2NCIsInZhbHVlIjoiNC4wMDAwMDAwMCJ9fSx7Im5hbWUiOiJzZWxsZXIiLCJ2YWx1ZSI6eyJ0eXBlIjoiT3B0aW9uYWwiLCJ2YWx1ZSI6eyJ0eXBlIjoiQWRkcmVzcyIsInZhbHVlIjoiMHg3MzUwMjA5ZjgzMjc0MmNkIn19fV19fQo=",
  "ApiCalls": 1
}
```

The `Result` property is a base64 encoded representation of the raw bytes of the Cadence value returned from the script.
The format is defined in
the [JSON-Cadence Data Interchance Format Specs](https://docs.onflow.org/cadence/json-cadence-spec/). You can use
the [flow-json-cadence-decoder](https://github.com/rayvin-flow/flow-json-cadence-decoder) library to decode it to a
plain JSON object.

If the request fails, the server will return a 500 status code with a JSON object in the body matching
the `ErrorResponse` schema defined below:

```typescript
type ErrorResponse = {
    "ApiCalls": number // number of API calls made to Flow
    "Error": string // string representation of the error
}
```

Example error response:

```json
{
  "ApiCalls": 1,
  "Error": "rpc error: code = OutOfRange desc = start height 19897000 is greater than the last sealed block height 19896844"
}
```

---

### Development

Clone this repository, and add the `access-nodes.json` file to the root of the project. **If you use a different
filename, you MUST modify the Dockerfile to match your JSON filename.** The default `docker-compose.yml` also expects a
valid `.env` file with configuration values at the root of the project. Here is a sample `.env` file:

```dotenv
LISTEN_PORT=8080
ACCESS_NODES=access-nodes.json
APP_LOG_LEVEL=debug
```

You can copy the provided `.env.example` file and modify it if necessary.

You can now build the docker image and launch it to start the `flow-beam` service

```shell
docker-compose build
docker-compose up
```

You must supply a valid list of Flow access nodes and their block height ranges in JSON format. We try to maintain an
up-to-date version of that file here:

[access-nodes.json](https://raw.githubusercontent.com/rayvin-flow/gists/main/access-nodes.json)

The JSON file has the following format:

```json
[
  {
    "StartHeight": 6483246,
    "EndHeight": 7601062,
    "Address": "access-001.candidate9.nodes.onflow.org:9000",
    "IsLegacy": true
  },
  {
    "StartHeight": 7601063,
    "EndHeight": 0,
    "Address": "access-001.mainnet1.nodes.onflow.org:9000",
    "IsLegacy": false
  }
]
```

Each entry in the array should have the following properties:

|Property|Type|Description|
|:---|:---|:---|
|StartHeight|number|Starting height of the access node|
|EndHeight|number|Ending height of the access node (use 0 to indicate that this is the current Spork)|
|Address|string|gRPC endpoint for the access node|
|IsLegacy|boolean|`true` if the legacy protobuf schema should be used (this should be used for candidate nodes)|
