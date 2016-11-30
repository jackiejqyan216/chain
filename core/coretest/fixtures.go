// Package coretest provides utilities for testing Chain Core.
package coretest

import (
	"context"
	"testing"
	"time"

	"chain/core/account"
	"chain/core/asset"
	"chain/core/pb"
	"chain/core/pin"
	"chain/core/txbuilder"
	"chain/crypto/ed25519/chainkd"
	"chain/errors"
	"chain/protocol"
	"chain/protocol/bc"
	"chain/protocol/state"
	"chain/testutil"
)

func CreatePins(ctx context.Context, t testing.TB, s *pin.Store) {
	pins := []string{account.PinName, asset.PinName, "tx"} // "tx" avoids circular dependency on query
	for _, p := range pins {
		err := s.CreatePin(ctx, p, 0)
		if err != nil {
			testutil.FatalErr(t, err)
		}
	}
}

func CreateAccount(ctx context.Context, t testing.TB, accounts *account.Manager, alias string, tags map[string]interface{}) string {
	keys := []chainkd.XPub{testutil.TestXPub}
	acc, err := accounts.Create(ctx, keys, 1, alias, tags, "")
	if err != nil {
		testutil.FatalErr(t, err)
	}
	return acc.ID
}

func CreateAsset(ctx context.Context, t testing.TB, assets *asset.Registry, def map[string]interface{}, alias string, tags map[string]interface{}) bc.AssetID {
	keys := []chainkd.XPub{testutil.TestXPub}
	asset, err := assets.Define(ctx, keys, 1, def, alias, tags, "")
	if err != nil {
		testutil.FatalErr(t, err)
	}
	return asset.AssetID
}

func IssueAssets(ctx context.Context, t testing.TB, c *protocol.Chain, s txbuilder.Submitter, assets *asset.Registry, accounts *account.Manager, assetID bc.AssetID, amount uint64, accountID string) state.Output {
	assetAmount := bc.AssetAmount{AssetID: assetID, Amount: amount}

	tpl, err := txbuilder.Build(ctx, nil, []txbuilder.Action{
		assets.NewIssueAction(assetAmount, nil), // does not support reference data
		accounts.NewControlAction(bc.AssetAmount{AssetID: assetID, Amount: amount}, accountID, nil),
	}, time.Now().Add(time.Minute))
	if err != nil {
		testutil.FatalErr(t, err)
	}

	SignTxTemplate(t, ctx, tpl, &testutil.TestXPrv)

	txdata, err := bc.NewTxDataFromBytes(tpl.RawTransaction)
	if err != nil {
		t.Log(errors.Stack(err))
		t.Fatal(err)
	}

	err = txbuilder.FinalizeTx(ctx, c, s, bc.NewTx(*txdata))
	if err != nil {
		testutil.FatalErr(t, err)
	}

	return state.Output{
		Outpoint: bc.Outpoint{Hash: txdata.Hash(), Index: 0},
		TxOutput: *txdata.Outputs[0],
	}
}

func Transfer(ctx context.Context, t testing.TB, c *protocol.Chain, s txbuilder.Submitter, actions []txbuilder.Action) *bc.Tx {
	template, err := txbuilder.Build(ctx, nil, actions, time.Now().Add(time.Minute))
	if err != nil {
		testutil.FatalErr(t, err)
	}

	SignTxTemplate(t, ctx, template, &testutil.TestXPrv)

	txdata, err := bc.NewTxDataFromBytes(template.RawTransaction)
	if err != nil {
		t.Log(errors.Stack(err))
		t.Fatal(err)
	}

	tx := bc.NewTx(*txdata)
	err = txbuilder.FinalizeTx(ctx, c, s, tx)
	if err != nil {
		testutil.FatalErr(t, err)
	}

	return tx
}

func SignTxTemplate(t testing.TB, ctx context.Context, template *pb.TxTemplate, priv *chainkd.XPrv) {
	if priv == nil {
		priv = &testutil.TestXPrv
	}
	err := txbuilder.Sign(ctx, template, []chainkd.XPub{priv.XPub()}, func(_ context.Context, _ chainkd.XPub, path [][]byte, data [32]byte) ([]byte, error) {
		derived := priv.Derive(path)
		return derived.Sign(data[:]), nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
