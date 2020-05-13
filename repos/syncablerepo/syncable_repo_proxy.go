package syncablerepo

import (
	"context"
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/mappers/syncablemapper"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
	"github.com/golang/protobuf/proto"
	"sync"
)

const (
	LatestHeight = 0
)

type ProxyRepo interface {
	//Queries
	GetHeadInfo() (*GetHeadResponse, errors.ApplicationError)
	GetByHeight(types.SyncableType, types.Height) (*syncable.Model, errors.ApplicationError)
}

type heightSequencePropsCache struct {
	items map[types.Height]*shared.HeightSequence

	mu sync.Mutex
}

type proxyRepo struct {
	blockClient       blockpb.BlockServiceClient
	transactionClient transactionpb.TransactionServiceClient

	heightSequencePropsCache heightSequencePropsCache
}

func NewProxyRepo(
	blockClient blockpb.BlockServiceClient,
	transactionClient transactionpb.TransactionServiceClient,
) ProxyRepo {
	return &proxyRepo{
		blockClient:       blockClient,
		transactionClient: transactionClient,

		heightSequencePropsCache: heightSequencePropsCache{
			items: map[types.Height]*shared.HeightSequence{},
			mu:    sync.Mutex{},
		},
	}
}

type GetHeadResponse struct {
	Height  types.Height
	Session int64
	Era     int64
}

func (r *proxyRepo) GetHeadInfo() (*GetHeadResponse, errors.ApplicationError) {
	ctx := context.Background()
	res, err := r.blockClient.GetHead(ctx, &blockpb.GetHeadRequest{})
	if err != nil {
		return nil, errors.NewError("error getting block by height", errors.ProxyRequestError, err)
	}
	return &GetHeadResponse{
		Height:  types.Height(res.Height),
		Session: res.Session,
		Era:     res.Era,
	}, nil
}

func (r *proxyRepo) GetByHeight(t types.SyncableType, h types.Height) (*syncable.Model, errors.ApplicationError) {
	ctx := context.Background()
	sequenceProps, err := r.getSequencePropsByHeight(ctx, h)
	if err != nil {
		return nil, err
	}

	data, err := r.getRawDataByHeight(ctx, t, h)
	if err != nil {
		return nil, err
	}
	return syncablemapper.FromProxy(t, *sequenceProps, data)
}

/*************** Private ***************/

func (r *proxyRepo) getSequencePropsByHeight(ctx context.Context, h types.Height) (*shared.HeightSequence, errors.ApplicationError) {
	r.heightSequencePropsCache.mu.Lock()
	defer r.heightSequencePropsCache.mu.Unlock()
	sequenceProps, ok := r.heightSequencePropsCache.items[h]
	if !ok {
		data, err := r.getRawDataByHeight(ctx, types.SyncableTypeBlock, h)
		if err != nil {
			return nil, err
		}

		res := data.(*blockpb.GetByHeightResponse)

		sequenceProps = &shared.HeightSequence{
			ChainUid:       res.GetChain(),
			SpecVersionUid: res.GetSpecVersion(),
			Height:         h,
			Time:           types.NewTimeFromTimestamp(*res.Block.Header.GetTime()),
		}
		r.heightSequencePropsCache.items[h] = sequenceProps
	}

	return sequenceProps, nil
}

func (r *proxyRepo) getRawDataByHeight(ctx context.Context, syncableType types.SyncableType, h types.Height) (proto.Message, errors.ApplicationError) {
	var res proto.Message
	var err error
	switch syncableType {
	case types.SyncableTypeBlock:
		res, err = r.blockClient.GetByHeight(ctx, &blockpb.GetByHeightRequest{Height: h.Int64()})
		if err != nil {
			return nil, errors.NewError("error getting block by height", errors.ProxyRequestError, err)
		}
	default:
		return nil, errors.NewErrorFromMessage(fmt.Sprintf("syncable type %s not found", syncableType), errors.ProxyRequestError)
	}
	return res, nil
}
