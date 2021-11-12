package beam

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/client/convert"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow/protobuf/go/flow/access"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"io/ioutil"
	"os"
)

type Event struct {
	Type             flow.EventType
	TransactionId    flow.Identifier
	TransactionIndex uint32
	EventIndex       uint32
	Payload          []byte
}

type EventsBlock struct {
	Id        flow.Identifier
	Height    uint64
	Timestamp int64
	Events    []Event
}

type GetEventsResponse struct {
	Blocks   []EventsBlock
	ApiCalls uint32
}

type GetLatestBlockHeightResponse struct {
	LatestBlockHeight uint64
	ApiCalls          uint32
}

type ExecuteScriptResponse struct {
	Result   []byte
	ApiCalls uint32
}

type AccessNodeInfo struct {
	StartHeight uint64
	EndHeight   uint64
	Address     string
	IsLegacy    bool
}

func GetAccessNodes() []AccessNodeInfo {
	var accessNodes []AccessNodeInfo

	accessNodesFile := os.Getenv("ACCESS_NODES")

	if len(accessNodesFile) == 0 {
		panic("ACCESS_NODES environment variable must be path to access nodes JSON file")
	}

	file, err := os.Open(accessNodesFile)

	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(file)

	err = json.Unmarshal(b, &accessNodes)

	if err != nil {
		panic(err)
	}

	return accessNodes
}

func GetEvents(requestEventType string, startHeight uint64, endHeight uint64) (GetEventsResponse, error) {
	ctx := context.Background()

	accessNodes := GetAccessNodes()

	var accessNode *AccessNodeInfo = nil

	var startRequestHeight = startHeight
	apiCalls := uint32(0)

	blocks := make([]EventsBlock, 0)

	for startRequestHeight <= endHeight {
		// find access node for start height
		for _, node := range accessNodes {
			if node.StartHeight <= startRequestHeight && (node.EndHeight == 0 || node.EndHeight >= startRequestHeight) {
				accessNode = &node
			}
		}

		if accessNode == nil {
			panic("No access node for height range")
		}

		var endRequestHeight = endHeight

		if accessNode.EndHeight > 0 && endRequestHeight > accessNode.EndHeight {
			// this request spans multiple access nodes
			endRequestHeight = accessNode.EndHeight
		}

		log.Debug().Msg(fmt.Sprintf("Requesting %s events for %d - %d", requestEventType, startRequestHeight, endRequestHeight))

		apiCalls += 1
		getEventsResponse, err := func(startRequestHeight uint64, endRequestHeight uint64) (*[]flow.BlockEvents, error) {
			var c AccessClient

			if accessNode.IsLegacy {
				c = newLegacyClient(accessNode.Address)
			} else {
				c = newClient(accessNode.Address)
			}

			defer c.Close()

			blockEvents, err := c.GetEventsForHeightRange(ctx, requestEventType, startRequestHeight, endRequestHeight)

			if err != nil {
				log.Error().Msg(fmt.Sprintf("Error getting %s events for %d - %d: %s", requestEventType, startRequestHeight, endRequestHeight, err))
				return nil, err
			}

			return &blockEvents, nil
		}(startRequestHeight, endRequestHeight)

		if err != nil {
			return GetEventsResponse{
				Blocks:   blocks,
				ApiCalls: apiCalls,
			}, err
		} else {
			for _, block := range *getEventsResponse {
				events := make([]Event, 0)

				for _, event := range block.Events {
					events = append(events, Event{
						Type:             event.Type,
						TransactionId:    event.TransactionID,
						TransactionIndex: event.TransactionIndex,
						EventIndex:       event.EventIndex,
						Payload:          event.Payload,
					})
				}

				blocks = append(blocks, EventsBlock{
					Id:        block.BlockID,
					Height:    block.BlockHeight,
					Timestamp: block.BlockTimestamp.Unix(),
					Events:    events,
				})
			}
		}

		startRequestHeight = endRequestHeight + 1
	}

	return GetEventsResponse{
		Blocks:   blocks,
		ApiCalls: apiCalls,
	}, nil
}

func GetLatestBlockHeight() (GetLatestBlockHeightResponse, error) {
	ctx := context.Background()

	// find current access node
	var accessNode *AccessNodeInfo = nil

	accessNodes := GetAccessNodes()

	for _, node := range accessNodes {
		if accessNode == nil || accessNode.StartHeight < node.StartHeight {
			accessNode = &node
		}
	}

	if accessNode == nil {
		panic("No access node for height range")
	}

	latestClient, err := NewClient(accessNode.Address, grpc.WithInsecure())

	if err != nil {
		return GetLatestBlockHeightResponse{
			LatestBlockHeight: 0,
			ApiCalls:          0,
		}, err
	}

	defer latestClient.close()

	result, err := latestClient.rpcClient.GetLatestBlockHeader(ctx, &access.GetLatestBlockHeaderRequest{IsSealed: true})

	if err != nil {
		return GetLatestBlockHeightResponse{
			LatestBlockHeight: 0,
			ApiCalls:          1,
		}, err
	}

	return GetLatestBlockHeightResponse{
		LatestBlockHeight: result.GetBlock().Height,
		ApiCalls: 1,
	}, nil
}

func ExecuteScript(script string, arguments []cadence.Value) (ExecuteScriptResponse, error) {
	ctx := context.Background()

	// find current access node
	var accessNode *AccessNodeInfo = nil

	accessNodes := GetAccessNodes()

	for _, node := range accessNodes {
		if accessNode == nil || accessNode.StartHeight < node.StartHeight {
			accessNode = &node
		}
	}

	if accessNode == nil {
		panic("No access node for height range")
	}

	latestClient, err := NewClient(accessNode.Address, grpc.WithInsecure())

	if err != nil {
		return ExecuteScriptResponse{
			ApiCalls: 0,
		}, err
	}

	defer latestClient.close()

	args, err := convert.CadenceValuesToMessages(arguments)

	if err != nil {
		return ExecuteScriptResponse{
			ApiCalls: 0,
		}, err
	}

	result, err := latestClient.rpcClient.ExecuteScriptAtLatestBlock(ctx, &access.ExecuteScriptAtLatestBlockRequest{
		Script:    []byte(script),
		Arguments: args,
	})

	if err != nil {
		return ExecuteScriptResponse{
			ApiCalls: 1,
		}, err
	}

	return ExecuteScriptResponse{
		ApiCalls: 1,
		Result:   result.Value,
	}, nil
}

func newClient(host string) AccessClient {
	c, err := NewClient(host, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	return c
}

func newLegacyClient(host string) AccessClient {
	c, err := NewLegacyClient(host, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	return c
}
