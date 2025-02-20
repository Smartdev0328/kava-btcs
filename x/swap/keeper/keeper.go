package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/kava-labs/kava/x/swap/types"
)

// Keeper keeper for the swap module
type Keeper struct {
	key           storetypes.StoreKey
	cdc           codec.Codec
	paramSubspace paramtypes.Subspace
	hooks         types.SwapHooks
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

// NewKeeper creates a new keeper
func NewKeeper(
	cdc codec.Codec,
	key storetypes.StoreKey,
	paramstore paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		key:           key,
		cdc:           cdc,
		paramSubspace: paramstore,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

// SetHooks adds hooks to the keeper.
func (k *Keeper) SetHooks(sh types.SwapHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set swap hooks twice")
	}
	k.hooks = sh
	return k
}

// ClearHooks clears the hooks on the keeper
func (k *Keeper) ClearHooks() {
	k.hooks = nil
}

// GetParams returns the params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var p types.Params
	k.paramSubspace.GetParamSet(ctx, &p)
	return p
}

// SetParams sets params on the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetSwapFee returns the swap fee set in the module parameters
func (k Keeper) GetSwapFee(ctx sdk.Context) sdk.Dec {
	return k.GetParams(ctx).SwapFee
}

// GetSwapModuleAccount returns the swap ModuleAccount
func (k Keeper) GetSwapModuleAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.accountKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
}

// GetPool retrieves a pool record from the store
func (k Keeper) GetPool(ctx sdk.Context, poolID string) (types.PoolRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PoolKeyPrefix)

	bz := store.Get(types.PoolKey(poolID))
	if bz == nil {
		return types.PoolRecord{}, false
	}

	var record types.PoolRecord
	k.cdc.MustUnmarshal(bz, &record)

	return record, true
}

// SetPool_Raw saves a pool record to the store without any validation
func (k Keeper) SetPool_Raw(ctx sdk.Context, record types.PoolRecord) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PoolKeyPrefix)
	bz := k.cdc.MustMarshal(&record)
	store.Set(types.PoolKey(record.PoolID), bz)
}

// SetPool saves a pool to the store and panics if the record is invalid
func (k Keeper) SetPool(ctx sdk.Context, record types.PoolRecord) {
	if err := record.Validate(); err != nil {
		panic(fmt.Sprintf("invalid pool record: %s", err))
	}

	k.SetPool_Raw(ctx, record)
}

// DeletePool deletes a pool record from the store
func (k Keeper) DeletePool(ctx sdk.Context, poolID string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PoolKeyPrefix)
	store.Delete(types.PoolKey(poolID))
}

// IteratePools iterates over all pool objects in the store and performs a callback function
func (k Keeper) IteratePools(ctx sdk.Context, cb func(record types.PoolRecord) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.PoolKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var record types.PoolRecord
		k.cdc.MustUnmarshal(iterator.Value(), &record)
		if cb(record) {
			break
		}
	}
}

// GetAllPools returns all pool records from the store
func (k Keeper) GetAllPools(ctx sdk.Context) (records types.PoolRecords) {
	k.IteratePools(ctx, func(record types.PoolRecord) bool {
		records = append(records, record)
		return false
	})
	return
}

// GetPoolShares gets the total shares in a pool from the store
func (k Keeper) GetPoolShares(ctx sdk.Context, poolID string) (sdk.Int, bool) {
	pool, found := k.GetPool(ctx, poolID)
	if !found {
		return sdk.Int{}, false
	}
	return pool.TotalShares, true
}

// GetDepositorShares gets a share record from the store
func (k Keeper) GetDepositorShares(ctx sdk.Context, depositor sdk.AccAddress, poolID string) (types.ShareRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositorPoolSharesPrefix)
	bz := store.Get(types.DepositorPoolSharesKey(depositor, poolID))
	if bz == nil {
		return types.ShareRecord{}, false
	}
	var record types.ShareRecord
	k.cdc.MustUnmarshal(bz, &record)
	return record, true
}

// SetDepositorShares_Raw saves a share record to the store without validation
func (k Keeper) SetDepositorShares_Raw(ctx sdk.Context, record types.ShareRecord) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositorPoolSharesPrefix)
	bz := k.cdc.MustMarshal(&record)
	store.Set(types.DepositorPoolSharesKey(record.Depositor, record.PoolID), bz)
}

// SetDepositorShares saves a share record to the store and panics if the record is invalid
func (k Keeper) SetDepositorShares(ctx sdk.Context, record types.ShareRecord) {
	if err := record.Validate(); err != nil {
		panic(fmt.Sprintf("invalid share record: %s", err))
	}

	k.SetDepositorShares_Raw(ctx, record)
}

// DeleteDepositorShares deletes a share record from the store
func (k Keeper) DeleteDepositorShares(ctx sdk.Context, depositor sdk.AccAddress, poolID string) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositorPoolSharesPrefix)
	store.Delete(types.DepositorPoolSharesKey(depositor, poolID))
}

// IterateDepositorShares iterates over all pool objects in the store and performs a callback function
func (k Keeper) IterateDepositorShares(ctx sdk.Context, cb func(record types.ShareRecord) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositorPoolSharesPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var record types.ShareRecord
		k.cdc.MustUnmarshal(iterator.Value(), &record)
		if cb(record) {
			break
		}
	}
}

// GetAllDepositorShares returns all depositor share records from the store
func (k Keeper) GetAllDepositorShares(ctx sdk.Context) (records types.ShareRecords) {
	k.IterateDepositorShares(ctx, func(record types.ShareRecord) bool {
		records = append(records, record)
		return false
	})
	return
}

// IterateDepositorSharesByOwner iterates over share records for a specific address and performs a callback function
func (k Keeper) IterateDepositorSharesByOwner(ctx sdk.Context, owner sdk.AccAddress, cb func(record types.ShareRecord) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.key), types.DepositorPoolSharesPrefix)
	iterator := sdk.KVStorePrefixIterator(store, owner.Bytes())
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var record types.ShareRecord
		k.cdc.MustUnmarshal(iterator.Value(), &record)
		if cb(record) {
			break
		}
	}
}

// GetAllDepositorSharesByOwner returns all depositor share records from the store for a specific address
func (k Keeper) GetAllDepositorSharesByOwner(ctx sdk.Context, owner sdk.AccAddress) (records types.ShareRecords) {
	k.IterateDepositorSharesByOwner(ctx, owner, func(record types.ShareRecord) bool {
		records = append(records, record)
		return false
	})
	return
}

// GetDepositorSharesAmount gets a depositor's shares in a pool from the store
func (k Keeper) GetDepositorSharesAmount(ctx sdk.Context, depositor sdk.AccAddress, poolID string) (sdk.Int, bool) {
	record, found := k.GetDepositorShares(ctx, depositor, poolID)
	if !found {
		return sdk.Int{}, false
	}
	return record.SharesOwned, true
}

// updatePool updates a pool, deleting the pool record if the shares are zero
func (k Keeper) updatePool(ctx sdk.Context, poolID string, pool *types.DenominatedPool) {
	if pool.TotalShares().IsZero() {
		k.DeletePool(ctx, poolID)
	} else {
		k.SetPool(ctx, types.NewPoolRecordFromPool(pool))
	}
}

// updateDepositorShares updates a depositor share records for a pool, deleting the record if the new shares are zero
func (k Keeper) updateDepositorShares(ctx sdk.Context, owner sdk.AccAddress, poolID string, shares sdk.Int) {
	if shares.IsZero() {
		k.DeleteDepositorShares(ctx, owner, poolID)
	} else {
		shareRecord := types.NewShareRecord(owner, poolID, shares)
		k.SetDepositorShares(ctx, shareRecord)
	}
}

func (k Keeper) loadDenominatedPool(ctx sdk.Context, poolID string) (*types.DenominatedPool, error) {
	poolRecord, found := k.GetPool(ctx, poolID)
	if !found {
		return &types.DenominatedPool{}, types.ErrInvalidPool
	}
	denominatedPool, err := types.NewDenominatedPoolWithExistingShares(poolRecord.Reserves(), poolRecord.TotalShares)
	if err != nil {
		return &types.DenominatedPool{}, types.ErrInvalidPool
	}
	return denominatedPool, nil
}
