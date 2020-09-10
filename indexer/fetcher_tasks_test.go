package indexer

import (
	"context"
	"reflect"
	"testing"

	mock "github.com/figment-networks/polkadothub-indexer/mock/indexer"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-proxy/grpc/block/blockpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/chain/chainpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/event/eventpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/height/heightpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/staking/stakingpb"
	"github.com/figment-networks/polkadothub-proxy/grpc/validatorperformance/validatorperformancepb"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
)

func TestFetcher_Run(t *testing.T) {
	errTestClient := errors.New("errTestClient")

	t.Run("returns error if client errors", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ctx := context.Background()

		mockClient := mock.NewMockFetcherClient(ctrl)
		task := NewFetcherTask(mockClient)

		pl := &payload{CurrentHeight: 20}

		mockClient.EXPECT().GetAll(pl.CurrentHeight).Return(nil, errTestClient).Times(1)

		if err := task.Run(ctx, pl); err != errTestClient {
			t.Errorf("want %v; got %v", errTestClient, err)
			return
		}
	})

	expectTimeStamp := ptypes.TimestampNow()
	expectHeightMeta := HeightMeta{
		Height:        20,
		Time:          *types.NewTimeFromTimestamp(*expectTimeStamp),
		SpecVersion:   "v1.0",
		ChainUID:      "chain123",
		Session:       1,
		Era:           2,
		LastInSession: false,
		LastInEra:     true,
	}
	expectBlock := &blockpb.Block{BlockHash: "hkasdbbjsd"}
	expectRawValidatorPerformance := []*validatorperformancepb.Validator{{StashAccount: "stash1"}}
	expectRawStaking := &stakingpb.Staking{Session: 5, Era: 6}
	expectRawEvents := []*eventpb.Event{{Method: "staking"}}

	t.Run("updates payload", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		ctx := context.Background()

		mockClient := mock.NewMockFetcherClient(ctrl)
		task := NewFetcherTask(mockClient)

		pl := &payload{CurrentHeight: 20}

		mockClient.EXPECT().GetAll(pl.CurrentHeight).Return(&heightpb.GetAllResponse{
			Block: &blockpb.GetByHeightResponse{Block: expectBlock},
			Chain: &chainpb.GetMetaByHeightResponse{
				Chain:         expectHeightMeta.ChainUID,
				Time:          expectTimeStamp,
				SpecVersion:   expectHeightMeta.SpecVersion,
				Session:       expectHeightMeta.Session,
				Era:           expectHeightMeta.Era,
				LastInSession: expectHeightMeta.LastInSession,
				LastInEra:     expectHeightMeta.LastInEra,
			},
			Staking:              &stakingpb.GetByHeightResponse{Staking: expectRawStaking},
			Event:                &eventpb.GetByHeightResponse{Events: expectRawEvents},
			ValidatorPerformance: &validatorperformancepb.GetByHeightResponse{Validators: expectRawValidatorPerformance},
		}, nil).Times(1)

		if err := task.Run(ctx, pl); err != nil {
			t.Errorf("unexpected error; got %v", err)
			return
		}
		if !reflect.DeepEqual(pl.RawBlock, expectBlock) {
			t.Errorf("want: %+v, got: %+v", expectBlock, pl.RawBlock)
			return
		}
		if !reflect.DeepEqual(pl.HeightMeta, expectHeightMeta) {
			t.Errorf("want: %+v, got: %+v", expectHeightMeta, pl.HeightMeta)
			return
		}
		if !reflect.DeepEqual(pl.RawValidatorPerformance, expectRawValidatorPerformance) {
			t.Errorf("want: %+v, got: %+v", expectRawValidatorPerformance, pl.RawValidatorPerformance)
			return
		}
		if !reflect.DeepEqual(pl.RawStaking, expectRawStaking) {
			t.Errorf("want: %+v, got: %+v", expectRawStaking, pl.RawStaking)
			return
		}
		if !reflect.DeepEqual(pl.RawEvents, expectRawEvents) {
			t.Errorf("want: %+v, got: %+v", expectRawEvents, pl.RawEvents)
			return
		}
	})
}
