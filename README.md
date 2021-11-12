`flow-beam` is an HTTP proxy service that fetches Cadence events from the Flow blockchain using
the [flow-go-sdk](https://github.com/onflow/flow-go-sdk). It can fetch events from any block height that has an access
node defined in the supplied JSON file. It was originally built to provide consistent access to historical Flow Cadence
events to services that were not written in Go.

---

### Instructions

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

Clone this repository, and add the `access-nodes.json` file to the root of the project. **If you use a different
filename, you MUST modify the Dockerfile to match your JSON filename.** The default `docker-compose.yml` also expects a
valid `.env` file with configuration values at the root of the project. Here is a sample `.env` file:

```dotenv
LISTEN_PORT=8081
ACCESS_NODES=access-nodes.json
APP_LOG_LEVEL=debug
```

You can copy the provided `.env.example` file and modify it if necessary.

You can now build the docker image and launch it to start the `flow-beam` service

```shell
docker-compose build
docker-compose up
```

You can access the HTTP server using the `LISTEN_PORT` that was specified in the `.env` file. If you kept the default
port of `8081`, then you should be able to make a GET request to the following URL:

`http://localhost:8081/events?eventType=[EVENTTYPE]&start=[START]&end=[END]`

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