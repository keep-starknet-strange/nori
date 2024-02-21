package nori

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRPCCacheImmutableRPCs(t *testing.T) {
	ctx := context.Background()

	cache := newRPCCache(newMemoryCache())
	ID := []byte(strconv.Itoa(1))

	rpcs := []struct {
		req  *RPCReq
		res  *RPCRes
		name string
	}{
		{
			req: &RPCReq{
				JSONRPC: "2.0",
				Method:  "starknet_chainId",
				ID:      ID,
			},
			res: &RPCRes{
				JSONRPC: "2.0",
				Result:  "0xff",
				ID:      ID,
			},
			name: "starknet_chainId",
		},
		{
			req: &RPCReq{
				JSONRPC: "2.0",
				Method:  "net_version",
				ID:      ID,
			},
			res: &RPCRes{
				JSONRPC: "2.0",
				Result:  "9999",
				ID:      ID,
			},
			name: "net_version",
		},
		{
			req: &RPCReq{
				JSONRPC: "2.0",
				Method:  "starknet_getBlockTransactionCount",
				Params:  mustMarshalJSON([]string{"0xb903239f8543d04b5dc1ba6579132b143087c68db1b2168786408fcbce568238"}),
				ID:      ID,
			},
			res: &RPCRes{
				JSONRPC: "2.0",
				Result:  `{"starknet_getBlockTransactionCount":"!"}`,
				ID:      ID,
			},
			name: "starknet_getBlockTransactionCount",
		},
		{
			req: &RPCReq{
				JSONRPC: "2.0",
				Method:  "starknet_getBlockWithTxs",
				Params:  mustMarshalJSON([]string{"0xc6ef2fc5426d6ad6fd9e2a26abeab0aa2411b7ab17f30a99d3cb96aed1d1055b", "false"}),
				ID:      ID,
			},
			res: &RPCRes{
				JSONRPC: "2.0",
				Result:  `{"starknet_getBlockWithTxs":"!"}`,
				ID:      ID,
			},
			name: "starknet_getBlockWithTxs",
		},
		{
			req: &RPCReq{
				JSONRPC: "2.0",
				Method:  "debug_getRawReceipts",
				Params:  mustMarshalJSON([]string{"0xc6ef2fc5426d6ad6fd9e2a26abeab0aa2411b7ab17f30a99d3cb96aed1d1055b"}),
				ID:      ID,
			},
			res: &RPCRes{
				JSONRPC: "2.0",
				Result:  []interface{}{"a"},
				ID:      ID,
			},
			name: "debug_getRawReceipts",
		},
	}

	for _, rpc := range rpcs {
		t.Run(rpc.name, func(t *testing.T) {
			err := cache.PutRPC(ctx, rpc.req, rpc.res)
			require.NoError(t, err)

			cachedRes, err := cache.GetRPC(ctx, rpc.req)
			require.NoError(t, err)
			require.Equal(t, rpc.res, cachedRes)
		})
	}
}

func TestRPCCacheUnsupportedMethod(t *testing.T) {
	ctx := context.Background()

	cache := newRPCCache(newMemoryCache())
	ID := []byte(strconv.Itoa(1))

	rpcs := []struct {
		req  *RPCReq
		name string
	}{
		{
			name: "starknet_syncing",
			req: &RPCReq{
				JSONRPC: "2.0",
				Method:  "starknet_syncing",
				ID:      ID,
			},
		},
		{
			name: "starknet_blockNumber",
			req: &RPCReq{
				JSONRPC: "2.0",
				Method:  "starknet_blockNumber",
				ID:      ID,
			},
		},
		{
			name: "starknet_getBlockWithTxs",
			req: &RPCReq{
				JSONRPC: "2.0",
				Method:  "starknet_getBlockWithTxs",
				ID:      ID,
			},
		},
		{
			name: "starknet_call",
			req: &RPCReq{
				JSONRPC: "2.0",
				Method:  "starknet_call",
				ID:      ID,
			},
		},
		{
			req: &RPCReq{
				JSONRPC: "2.0",
				Method:  "debug_getRawReceipts",
				Params:  mustMarshalJSON([]string{"0x100"}),
				ID:      ID,
			},
			name: "debug_getRawReceipts",
		},
	}

	for _, rpc := range rpcs {
		t.Run(rpc.name, func(t *testing.T) {
			fakeval := mustMarshalJSON([]string{rpc.name})
			err := cache.PutRPC(ctx, rpc.req, &RPCRes{Result: fakeval})
			require.NoError(t, err)

			cachedRes, err := cache.GetRPC(ctx, rpc.req)
			require.NoError(t, err)
			require.Nil(t, cachedRes)
		})
	}

}
