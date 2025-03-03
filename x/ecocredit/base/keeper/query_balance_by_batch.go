package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/orm/model/ormlist"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	api "github.com/regen-network/regen-ledger/api/regen/ecocredit/v1"
	"github.com/regen-network/regen-ledger/types/ormutil"
	types "github.com/regen-network/regen-ledger/x/ecocredit/base/types/v1"
)

func (k Keeper) BalancesByBatch(ctx context.Context, req *types.QueryBalancesByBatchRequest) (*types.QueryBalancesByBatchResponse, error) {
	pg, err := ormutil.GogoPageReqToPulsarPageReq(req.Pagination)
	if err != nil {
		return nil, err
	}
	batch, err := k.stateStore.BatchTable().GetByDenom(ctx, req.BatchDenom)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("could not get batch with denom %s: %s", req.BatchDenom, err.Error())
	}
	it, err := k.stateStore.BatchBalanceTable().List(ctx, api.BatchBalanceBatchKeyAddressIndexKey{}.WithBatchKey(batch.Key), ormlist.Paginate(pg))
	if err != nil {
		return nil, err
	}
	defer it.Close()

	balances := make([]*types.BatchBalanceInfo, 0, 10) // preallocate
	for it.Next() {
		bal, err := it.Value()
		if err != nil {
			return nil, err
		}
		balances = append(balances, &types.BatchBalanceInfo{
			Address:        sdk.AccAddress(bal.Address).String(),
			BatchDenom:     batch.Denom,
			TradableAmount: bal.TradableAmount,
			RetiredAmount:  bal.RetiredAmount,
			EscrowedAmount: bal.EscrowedAmount,
		})
	}
	pr, err := ormutil.PulsarPageResToGogoPageRes(it.PageResponse())
	if err != nil {
		return nil, err
	}
	return &types.QueryBalancesByBatchResponse{
		Balances:   balances,
		Pagination: pr,
	}, nil
}
