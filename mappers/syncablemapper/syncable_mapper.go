package syncablemapper

import (
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/transaction/transactionpb"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func FromProxy(syncableType types.SyncableType, sequence shared.HeightSequence, data proto.Message) (*syncable.Model, errors.ApplicationError) {
	var bytes string
	var err error
	marshaler := jsonpb.Marshaler{}

	switch syncableType {
	case types.SyncableTypeBlock:
		res := data.(*blockpb.GetByHeightResponse)
		bytes, err = marshaler.MarshalToString(res)
	default:
		return nil, errors.NewErrorFromMessage(fmt.Sprintf("syncable type %s not found", syncableType), errors.ProxyRequestError)
	}

	if err != nil {
		return nil, errors.NewErrorFromMessage(fmt.Sprintf("syncable type %s could not be marshaled to JSON", syncableType), errors.ProxyUnmarshalError)
	}

	d := types.Jsonb{RawMessage: []byte(bytes)}

	e := &syncable.Model{
		SequenceId: types.SequenceId(sequence.Height),
		SequenceType: types.SequenceTypeHeight,
		Data: d,
		Type: syncableType,
	}

	if !e.Valid() {
		return nil, errors.NewErrorFromMessage("syncable not valid", errors.NotValid)
	}
	return e, nil
}

func UnmarshalBlockData(data types.Jsonb) (*blockpb.GetByHeightResponse, errors.ApplicationError) {
	res := &blockpb.GetByHeightResponse{}
	err := jsonpb.UnmarshalString(string(data.RawMessage), res)
	if err != nil {
		return nil, errors.NewError("could not unmarshal grpc block response", errors.ProxyUnmarshalError, err)
	}

	return res, nil
}

func UnmarshalTransactionsData(data types.Jsonb) (*transactionpb.GetByHeightResponse, errors.ApplicationError) {
	res := &transactionpb.GetByHeightResponse{}
	err := jsonpb.UnmarshalString(string(data.RawMessage), res)
	if err != nil {
		return nil, errors.NewError("could not unmarshal grpc transaction response", errors.ProxyUnmarshalError, err)
	}

	return res, nil
}