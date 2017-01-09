package core

import (
	"encoding/json"
	"sync"

	"golang.org/x/net/context"

	"chain/core/pb"
	"chain/core/signers"
	"chain/crypto/ed25519/chainkd"
	cjson "chain/encoding/json"
	"chain/errors"
	"chain/net/http/reqid"
)

// This type enforces JSON field ordering in API output.
type accountResponse struct {
	ID     string          `json:"id"`
	Alias  string          `json:"alias"`
	Keys   []*accountKey   `json:"keys"`
	Quorum int32           `json:"quorum"`
	Tags   json.RawMessage `json:"tags"`
}

type accountKey struct {
	RootXPub              cjson.HexBytes   `json:"root_xpub"`
	AccountXPub           cjson.HexBytes   `json:"account_xpub"`
	AccountDerivationPath []cjson.HexBytes `json:"account_derivation_path"`
}

func (h *Handler) CreateAccounts(ctx context.Context, in *pb.CreateAccountsRequest) (*pb.CreateAccountsResponse, error) {
	responses := make([]*pb.CreateAccountsResponse_Response, len(in.Requests))
	var wg sync.WaitGroup
	wg.Add(len(in.Requests))

	for i := range in.Requests {
		go func(i int) {
			req := in.Requests[i]
			subctx := reqid.NewSubContext(ctx, reqid.New())
			defer wg.Done()
			defer batchRecover(func(err error) {
				detailedErr, _ := errInfo(err)
				responses[i] = &pb.CreateAccountsResponse_Response{
					Error: protobufErr(detailedErr),
				}
			})

			var tags map[string]interface{}
			err := json.Unmarshal(req.Tags, &tags)
			if err != nil {
				detailedErr, _ := errInfo(err)
				responses[i] = &pb.CreateAccountsResponse_Response{
					Error: protobufErr(detailedErr),
				}
				return
			}

			xpubs, err := bytesToKeys(req.RootXpubs)
			if err != nil {
				detailedErr, _ := errInfo(err)
				responses[i] = &pb.CreateAccountsResponse_Response{
					Error: protobufErr(detailedErr),
				}
				return
			}

			acc, err := h.Accounts.Create(subctx, xpubs, int(req.Quorum), req.Alias, tags, req.ClientToken)
			if err != nil {
				detailedErr, _ := errInfo(err)
				responses[i] = &pb.CreateAccountsResponse_Response{
					Error: protobufErr(detailedErr),
				}
				return
			}
			path := signers.Path(acc.Signer, signers.AccountKeySpace)
			var keys []*pb.Account_Key
			for _, xpub := range acc.XPubs {
				derived := xpub.Derive(path)
				keys = append(keys, &pb.Account_Key{
					RootXpub:              xpub[:],
					AccountXpub:           derived[:],
					AccountDerivationPath: path,
				})
			}
			responses[i] = &pb.CreateAccountsResponse_Response{
				Account: &pb.Account{
					Id:     acc.ID,
					Alias:  acc.Alias,
					Keys:   keys,
					Quorum: int32(acc.Quorum),
					Tags:   req.Tags,
				},
			}
		}(i)
	}

	wg.Wait()
	return &pb.CreateAccountsResponse{
		Responses: responses,
	}, nil
}

func bytesToKeys(keyBytes [][]byte) ([]chainkd.XPub, error) {
	xpubs := make([]chainkd.XPub, 0, len(keyBytes))
	for _, k := range keyBytes {
		var xpub chainkd.XPub
		if len(k) != len(xpub) {
			return nil, errors.Wrap(chainkd.ErrBadKeyLen)
		}
		copy(xpub[:], k)
		xpubs = append(xpubs, xpub)
	}
	return xpubs, nil
}
