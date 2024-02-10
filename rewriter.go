package nori

import (
	"encoding/json"
	"errors"

	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type RewriteContext struct {
	latest        hexutil.Uint64
	safe          hexutil.Uint64
	finalized     hexutil.Uint64
	maxBlockRange uint64
}

type RewriteResult uint8

const (
	// RewriteNone means request should be forwarded as-is
	RewriteNone RewriteResult = iota

	// RewriteOverrideError means there was an error attempting to rewrite
	RewriteOverrideError

	// RewriteOverrideRequest means the modified request should be forwarded to the backend
	RewriteOverrideRequest

	// RewriteOverrideResponse means to skip calling the backend and serve the overridden response
	RewriteOverrideResponse
)

var (
	ErrRewriteBlockOutOfRange = errors.New("block is out of range")
	ErrRewriteRangeTooLarge   = errors.New("block range is too large")
)

// RewriteTags modifies the request and the response based on block tags
func RewriteTags(rctx RewriteContext, req *RPCReq, res *RPCRes) (RewriteResult, error) {
	rw, err := RewriteResponse(rctx, req, res)
	if rw == RewriteOverrideResponse {
		return rw, err
	}
	return RewriteRequest(rctx, req, res)
}

// RewriteResponse modifies the response object to comply with the rewrite context
// after the method has been called at the backend
// RewriteResult informs the decision of the rewrite
func RewriteResponse(rctx RewriteContext, req *RPCReq, res *RPCRes) (RewriteResult, error) {
	switch req.Method {
	case "starknet_blockNumber":
		res.Result = rctx.latest
		return RewriteOverrideResponse, nil
	}
	return RewriteNone, nil
}

// RewriteRequest modifies the request object to comply with the rewrite context
// before the method has been called at the backend
// it returns false if nothing was changed
func RewriteRequest(rctx RewriteContext, req *RPCReq, res *RPCRes) (RewriteResult, error) {
	switch req.Method {
	case "debug_getRawReceipts", "consensus_getReceipts":
		return rewriteParam(rctx, req, res, 0, true, false)
	case "starknet_call":
		return rewriteParam(rctx, req, res, 1, false, true)
	case "starknet_getStorageAt":
		return rewriteParam(rctx, req, res, 2, false, true)
	case "starknet_getBlockTransactionCount",
		"starknet_getBlockByNumber",
        "starknet_getTransactionByBlockIdAndIndex":
		return rewriteParam(rctx, req, res, 0, false, false)
	}
	return RewriteNone, nil
}

func rewriteParam(rctx RewriteContext, req *RPCReq, res *RPCRes, pos int, required bool, blockNrOrHash bool) (RewriteResult, error) {
	var p []interface{}
	err := json.Unmarshal(req.Params, &p)
	if err != nil {
		return RewriteOverrideError, err
	}

	// we assume latest if the param is missing,
	// and we don't rewrite if there is not enough params
	if len(p) == pos && !required {
		p = append(p, "latest")
	} else if len(p) <= pos {
		return RewriteNone, nil
	}

	// support for https://eips.ethereum.org/EIPS/eip-1898
	var val interface{}
	var rw bool
	if blockNrOrHash {
		bnh, err := remarshalBlockID(p[pos])
		if err != nil {
			// fallback to string
			s, ok := p[pos].(string)
			if ok {
				val, rw, err = rewriteTag(rctx, s)
				if err != nil {
					return RewriteOverrideError, err
				}
			} else {
				return RewriteOverrideError, errors.New("expected BlockNumberOrHash or string")
			}
		} else {
			val, rw, err = rewriteTagBlockID(rctx, bnh)
			if err != nil {
				return RewriteOverrideError, err
			}
		}
	} else {
		s, ok := p[pos].(string)
		if !ok {
			return RewriteOverrideError, errors.New("expected string")
		}

		val, rw, err = rewriteTag(rctx, s)
		if err != nil {
			return RewriteOverrideError, err
		}
	}

	if rw {
		p[pos] = val
		paramRaw, err := json.Marshal(p)
		if err != nil {
			return RewriteOverrideError, err
		}
		req.Params = paramRaw
		return RewriteOverrideRequest, nil
	}
	return RewriteNone, nil
}

func blockNumber(m map[string]interface{}, key string, latest uint64) (uint64, error) {
	current, ok := m[key].(string)
	if !ok {
		return 0, errors.New("expected string")
	}
	// the latest/safe/finalized tags are already replaced by rewriteTag
	if current == "earliest" {
		return 0, nil
	}
	if current == "pending" {
		return latest + 1, nil
	}
	return hexutil.DecodeUint64(current)
}

func remarshalBlockID(current interface{}) (*rpc.BlockID, error) {
	jv, err := json.Marshal(current)
	if err != nil {
		return nil, err
	}

	var bnh rpc.BlockID
	err = bnh.Hash.UnmarshalJSON(jv)
	if err != nil {
		return nil, err
	}

	return &bnh, nil
}
