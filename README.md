# Nori

Nori (のり) - Derived from "Norikae" (乗り換え), meaning transfer or switch, this name reflects the idea of efficiently switching or balancing network requests, akin to a load balancer or router.

It's essentially a fork of [Optimism proxyd](https://github.com/ethereum-optimism/optimism/tree/develop/proxyd), adapted to work with Starknet.

This tool implements `nori`, an RPC request router and proxy. It does the following things:

1. Whitelists RPC methods.
2. Routes RPC methods to groups of backend services.
3. Automatically retries failed backend requests.
4. Track backend consensus (`latest`, `safe`, `finalized` blocks), peer count and sync state.
5. Re-write requests and responses to enforce consensus.
6. Load balance requests across backend services.
7. Cache immutable responses from backends.
8. Provides metrics the measure request latency, error rates, and the like.


## Usage

Run `make nori` to build the binary. No additional dependencies are necessary.

To configure `nori` for use, you'll need to create a configuration file to define your proxy backends and routing rules.  Check out [example.config.toml](./example.config.toml) for how to do this alongside a full list of all options with commentary.

Once you have a config file, start the daemon via `nori <path-to-config>.toml`.


## Consensus awareness

Starting on v4.0.0, `nori` is aware of the consensus state of its backends. This helps minimize chain reorgs experienced by clients.

To enable this behavior, you must set `consensus_aware` value to `true` in the backend group.

When consensus awareness is enabled, `nori` will poll the backends for their states and resolve a consensus group based on:
* the common ancestor `latest` block, i.e. if a backend is experiencing a fork, the fork won't be visible to the clients
* the lowest `safe` block
* the lowest `finalized` block
* peer count
* sync state

The backend group then acts as a round-robin load balancer distributing traffic equally across healthy backends in the consensus group, increasing the availability of the proxy.

A backend is considered healthy if it meets the following criteria:
* not banned
* avg 1-min moving window error rate ≤ configurable threshold
* avg 1-min moving window latency ≤ configurable threshold
* peer count ≥ configurable threshold
* `latest` block lag ≤ configurable threshold
* last state update ≤ configurable threshold
* not currently syncing

When a backend is experiencing inconsistent consensus, high error rates or high latency,
the backend will be banned for a configurable amount of time (default 5 minutes)
and won't receive any traffic during this period.


## Tag rewrite

When consensus awareness is enabled, `nori` will enforce the consensus state transparently for all the clients.

For example, if a client requests the `starknet_getBlockWithTxs` method with the `latest` tag,
`nori` will rewrite the request to use the resolved latest block from the consensus group
and forward it to the backend.

The following request methods are rewritten:
* `starknet_call`
* `starknet_estimateFee` *
* `starknet_estimateMessageFee` *
* `starknet_getBlockTransactionCount`
* `starknet_getBlockWithReceipts`
* `starknet_getBlockWithTxHashes` *
* `starknet_getBlockWithTxs`
* `starknet_getClass` *
* `starknet_getClassAt` *
* `starknet_getClassHashAt` *
* `starknet_getNonce` *
* `starknet_getStateUpdate` *
* `starknet_getStorageAt`
* `starknet_getTransactionByBlockIdAndIndex`

\* not implemented

And `starknet_blockNumber` response is overridden with current block consensus.


## Cacheable methods

Cache use Redis and can be enabled for the following immutable methods:

* `starknet_chainId`
* `starknet_getBlockTransactionCount`
* `starknet_getBlockWithReceipts` (block hash only)
* `starknet_getBlockWithTxs`
* `starknet_getTransactionByIdAndIndex`

## Meta method `consensus_getReceipts`
> Only available after v0.7.0-rc0

To support backends with different specifications in the same backend group,
nori exposes a convenient method to fetch receipts abstracting away
what specific backend will serve the request.

Each backend specifies their preferred method to fetch receipts with `consensus_receipts_target` config,
which will be translated from `consensus_getReceipts`.

This method takes a `blockId` (i.e. `tag|qty|hash`)
and returns the receipts for all transactions in the block.

Request example
```json
{
  "jsonrpc":"2.0",
  "id": 1,
  "params": [{"block_hash": "0xc6ef2fc5426d6ad6fd9e2a26abeab0aa2411b7ab17f30a99d3cb96aed1d1055b"}]
}
```

It currently supports translation to the following targets:
* `starknet_getBlockWithReceipts(blockId)` (default)

The selected target is returned in the response, in a wrapped result.

Response example
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "method": "debug_getRawReceipts",
    "result": {
      // the actual raw result from backend
    }
  }
}
```

## Metrics

See `metrics.go` for a list of all available metrics.

The metrics port is configurable via the `metrics.port` and `metrics.host` keys in the config.

## Adding Backend SSL Certificates in Docker

The Docker image runs on Alpine Linux. If you get SSL errors when connecting to a backend within Docker, you may need to add additional certificates to Alpine's certificate store. To do this, bind mount the certificate bundle into a file in `/usr/local/share/ca-certificates`. The `entrypoint.sh` script will then update the store with whatever is in the `ca-certificates` directory prior to starting `nori`.
