package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/orm/model/ormlist"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	api "github.com/regen-network/regen-ledger/api/regen/ecocredit/v1"
	regentypes "github.com/regen-network/regen-ledger/types"
	"github.com/regen-network/regen-ledger/types/ormutil"
	types "github.com/regen-network/regen-ledger/x/ecocredit/base/types/v1"
)

// BatchesByProject queries for all batches in the given credit class.
func (k Keeper) BatchesByProject(ctx context.Context, request *types.QueryBatchesByProjectRequest) (*types.QueryBatchesByProjectResponse, error) {
	pg, err := ormutil.GogoPageReqToPulsarPageReq(request.Pagination)
	if err != nil {
		return nil, err
	}

	project, err := k.stateStore.ProjectTable().GetById(ctx, request.ProjectId)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("could not get project with id %s: %s", request.ProjectId, err.Error())
	}

	it, err := k.stateStore.BatchTable().List(ctx, api.BatchProjectKeyIndexKey{}.WithProjectKey(project.Key), ormlist.Paginate(pg))
	if err != nil {
		return nil, err
	}
	defer it.Close()

	batches := make([]*types.BatchInfo, 0)
	for it.Next() {
		batch, err := it.Value()
		if err != nil {
			return nil, err
		}

		issuer := sdk.AccAddress(batch.Issuer)

		info := types.BatchInfo{
			Issuer:       issuer.String(),
			ProjectId:    project.Id,
			Denom:        batch.Denom,
			Metadata:     batch.Metadata,
			StartDate:    regentypes.ProtobufToGogoTimestamp(batch.StartDate),
			EndDate:      regentypes.ProtobufToGogoTimestamp(batch.EndDate),
			IssuanceDate: regentypes.ProtobufToGogoTimestamp(batch.IssuanceDate),
			Open:         batch.Open,
		}

		batches = append(batches, &info)
	}

	pr, err := ormutil.PulsarPageResToGogoPageRes(it.PageResponse())
	if err != nil {
		return nil, err
	}

	return &types.QueryBatchesByProjectResponse{
		Batches:    batches,
		Pagination: pr,
	}, nil
}
