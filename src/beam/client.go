package beam

import (
	"context"
	"github.com/onflow/flow-go/engine/common/rpc/convert"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flow/protobuf/go/flow/access"
	legacyaccess "github.com/onflow/flow/protobuf/go/flow/legacy/access"
	legacyentities "github.com/onflow/flow/protobuf/go/flow/legacy/entities"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Block struct {
	Id            flow.Identifier
	Height        uint64
	Timestamp     int64
	CollectionIds []flow.Identifier
}

type Collection struct {
	Id             flow.Identifier
	TransactionIds []flow.Identifier
}

type TransactionEvent struct {
	TransactionIndex uint32
	EventIndex       uint32
	Type             string
	Payload          []byte
}

type TransactionResult struct {
	Status       int32
	StatusCode   uint32
	ErrorMessage string
	Events       []TransactionEvent
}

type AccessClient interface {
	GetEventsForHeightRange(
		ctx context.Context, eventType string, startHeight, endHeight uint64,
	) ([]flow.BlockEvents, error)

	GetBlockByHeight(
		ctx context.Context,
		height uint64,
	) (*Block, error)

	GetCollectionByID(
		ctx context.Context,
		collectionId []byte,
	) (*Collection, error)

	GetTransactionResult(
		ctx context.Context,
		transactionId []byte,
	) (*TransactionResult, error)

	Close() error
}

type Client struct {
	rpcClient access.AccessAPIClient
	close     func() error
}

func NewClient(addr string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	grpcClient := access.NewAccessAPIClient(conn)

	return &Client{
		rpcClient: grpcClient,
		close:     func() error { return conn.Close() },
	}, nil
}

func (c *Client) GetEventsForHeightRange(
	ctx context.Context,
	eventType string,
	startHeight,
	endHeight uint64,
) ([]flow.BlockEvents, error) {
	req := access.GetEventsForHeightRangeRequest{
		Type:        eventType,
		StartHeight: startHeight,
		EndHeight:   endHeight,
	}

	res, err := c.rpcClient.GetEventsForHeightRange(ctx, &req, grpc.MaxCallRecvMsgSize(1024*1024*50))
	if err != nil {
		return nil, err
	}

	results := res.GetResults()

	blockResults := make([]flow.BlockEvents, len(results))

	for i, result := range results {
		blockResults[i] = flow.BlockEvents{
			BlockID:        convert.MessageToIdentifier(result.BlockId),
			BlockHeight:    result.GetBlockHeight(),
			BlockTimestamp: result.GetBlockTimestamp().AsTime(),
			Events:         convert.MessagesToEvents(result.Events),
		}
	}

	return blockResults, nil
}

func (c *Client) GetBlockByHeight(
	ctx context.Context,
	height uint64,
) (*Block, error) {
	getBlockByHeightRequest := access.GetBlockByHeightRequest{
		Height: height,
	}

	getBlockByHeightResponse, err := c.rpcClient.GetBlockByHeight(ctx, &getBlockByHeightRequest)

	if err != nil {
		return nil, err
	}

	collectionIds := make([]flow.Identifier, 0)

	for _, coll := range getBlockByHeightResponse.Block.CollectionGuarantees {
		collectionIds = append(collectionIds, convert.MessageToIdentifier(coll.CollectionId))
	}

	return &Block{
		Id:            convert.MessageToIdentifier(getBlockByHeightResponse.Block.Id),
		Height:        getBlockByHeightResponse.Block.Height,
		Timestamp:     getBlockByHeightResponse.Block.Timestamp.Seconds,
		CollectionIds: collectionIds,
	}, nil
}

func (c *Client) GetCollectionByID(
	ctx context.Context,
	collectionId []byte,
) (*Collection, error) {
	getCollectionRequest := access.GetCollectionByIDRequest{
		Id: collectionId,
	}

	getCollectionResponse, err := c.rpcClient.GetCollectionByID(ctx, &getCollectionRequest)

	if err != nil {
		return nil, err
	}

	transactionIds := make([]flow.Identifier, 0)

	for _, transactionId := range getCollectionResponse.Collection.TransactionIds {
		transactionIds = append(transactionIds, convert.MessageToIdentifier(transactionId))
	}

	return &Collection{
		Id:             convert.MessageToIdentifier(getCollectionResponse.Collection.Id),
		TransactionIds: transactionIds,
	}, nil
}

func (c *Client) GetTransactionResult(
	ctx context.Context,
	transactionId []byte,
) (*TransactionResult, error) {
	getTransactionRequest := access.GetTransactionRequest{
		Id: transactionId,
	}

	getTransactionResultResponse, err := c.rpcClient.GetTransactionResult(ctx, &getTransactionRequest, grpc.MaxCallRecvMsgSize(1024*1024*50))

	if err != nil {
		return nil, err
	}

	events := make([]TransactionEvent, 0)

	for _, event := range getTransactionResultResponse.Events {
		events = append(events, TransactionEvent{
			TransactionIndex: event.TransactionIndex,
			EventIndex:       event.EventIndex,
			Type:             event.Type,
			Payload:          event.Payload,
		})
	}

	return &TransactionResult{
		Status:       int32(getTransactionResultResponse.Status),
		StatusCode:   getTransactionResultResponse.StatusCode,
		ErrorMessage: getTransactionResultResponse.ErrorMessage,
		Events:       events,
	}, nil
}

func (c *Client) Close() error {
	return c.close()
}

type LegacyClient struct {
	rpcClient legacyaccess.AccessAPIClient
	close     func() error
}

func NewLegacyClient(addr string, opts ...grpc.DialOption) (*LegacyClient, error) {
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	grpcClient := legacyaccess.NewAccessAPIClient(conn)

	return &LegacyClient{
		rpcClient: grpcClient,
		close:     func() error { return conn.Close() },
	}, nil
}

func (c *LegacyClient) GetEventsForHeightRange(
	ctx context.Context,
	eventType string,
	startHeight,
	endHeight uint64,
) ([]flow.BlockEvents, error) {
	req := legacyaccess.GetEventsForHeightRangeRequest{
		Type:        eventType,
		StartHeight: startHeight,
		EndHeight:   endHeight,
	}

	res, err := c.rpcClient.GetEventsForHeightRange(ctx, &req, grpc.MaxCallRecvMsgSize(1024*1024*50))
	if err != nil {
		return nil, err
	}

	results := res.GetResults()

	blockResults := make([]flow.BlockEvents, len(results))

	for i, result := range results {
		blockResults[i] = flow.BlockEvents{
			BlockID:     convert.MessageToIdentifier(result.BlockId),
			BlockHeight: result.GetBlockHeight(),
			Events:      legacyMessagesToEvents(result.Events),
		}
	}

	return blockResults, nil
}

func (c *LegacyClient) GetBlockByHeight(
	ctx context.Context,
	height uint64,
) (*Block, error) {
	getBlockByHeightRequest := legacyaccess.GetBlockByHeightRequest{
		Height: height,
	}

	log.Debug().Msg("legacy get block by height")

	getBlockByHeightResponse, err := c.rpcClient.GetBlockByHeight(ctx, &getBlockByHeightRequest)

	if err != nil {
		return nil, err
	}

	collectionIds := make([]flow.Identifier, 0)

	for _, coll := range getBlockByHeightResponse.Block.CollectionGuarantees {
		collectionIds = append(collectionIds, convert.MessageToIdentifier(coll.CollectionId))
	}

	return &Block{
		Id:            convert.MessageToIdentifier(getBlockByHeightResponse.Block.Id),
		Height:        getBlockByHeightResponse.Block.Height,
		Timestamp:     getBlockByHeightResponse.Block.Timestamp.Seconds,
		CollectionIds: collectionIds,
	}, nil
}

func (c *LegacyClient) GetCollectionByID(
	ctx context.Context,
	collectionId []byte,
) (*Collection, error) {
	getCollectionRequest := legacyaccess.GetCollectionByIDRequest{
		Id: collectionId,
	}

	getCollectionResponse, err := c.rpcClient.GetCollectionByID(ctx, &getCollectionRequest)

	if err != nil {
		return nil, err
	}

	transactionIds := make([]flow.Identifier, 0)

	for _, transactionId := range getCollectionResponse.Collection.TransactionIds {
		transactionIds = append(transactionIds, convert.MessageToIdentifier(transactionId))
	}

	return &Collection{
		Id:             convert.MessageToIdentifier(getCollectionResponse.Collection.Id),
		TransactionIds: transactionIds,
	}, nil
}

func (c *LegacyClient) GetTransactionResult(
	ctx context.Context,
	transactionId []byte,
) (*TransactionResult, error) {
	getTransactionRequest := legacyaccess.GetTransactionRequest{
		Id: transactionId,
	}

	getTransactionResultResponse, err := c.rpcClient.GetTransactionResult(ctx, &getTransactionRequest, grpc.MaxCallRecvMsgSize(1024*1024*50))

	if err != nil {
		return nil, err
	}

	events := make([]TransactionEvent, 0)

	for _, event := range getTransactionResultResponse.Events {
		events = append(events, TransactionEvent{
			TransactionIndex: event.TransactionIndex,
			EventIndex:       event.EventIndex,
			Type:             event.Type,
			Payload:          event.Payload,
		})
	}

	return &TransactionResult{
		Status:       int32(getTransactionResultResponse.Status),
		StatusCode:   getTransactionResultResponse.StatusCode,
		ErrorMessage: getTransactionResultResponse.ErrorMessage,
		Events:       events,
	}, nil
}

func legacyMessagesToEvents(m []*legacyentities.Event) []flow.Event {
	events := make([]flow.Event, len(m))

	for i, event := range m {
		events[i] = legacyMessageToEvent(event)
	}

	return events
}

func legacyMessageToEvent(m *legacyentities.Event) flow.Event {
	return flow.Event{
		Type:             flow.EventType(m.GetType()),
		TransactionID:    flow.HashToID(m.GetTransactionId()),
		TransactionIndex: m.GetTransactionIndex(),
		EventIndex:       m.GetEventIndex(),
		Payload:          m.GetPayload(),
	}
}

func (c *LegacyClient) Close() error {
	return c.close()
}
