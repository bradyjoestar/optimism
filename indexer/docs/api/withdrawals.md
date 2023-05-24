## User Withdrawals

Returns a list of withdrawals

### Request

- **URL** `/api/v0/withdrawals`

- **Method** GET

To make a request pass in the address with any query params

- **Example**

```bash
CURL https://localhost:8080/api/v0/withdrawals/0x8F0EBDaA1cF7106bE861753B0f9F5c0250fE0819?limit=10
```

- **Query Params**

`limit=` - limits the returned records.   Defaults to `limit=10`

`cursor` - Cursor to start fetching the next `limit` of results.  If not provided starts at the first record

`sortDirection=` - Sort direction.   Defaults to `sortDirection=asc`

`sortBy` - Field to sort on.   

### Response

Returns a paginated list of withdrawals

The the request/response is similar to [deposts](./deposits.md) except it includes `prove` and `claim` transaction information as well as the withdrawal state.

- **Example response**

```typescript
{ 
  "cursor": "d26bc98b-24bf-4cc5-96b7-b1092f3fb409",
  "hasNextPage": true,
  "items": {
    "guid":"955dcb6e-32ea-4a6e-a01e-831cc40a7c6e",
    "blockTimestamp":1684873656,
    "from":"0x8F0EBDaA1cF7106bE861753B0f9F5c0250fE0819",
    "to":"0x8F0EBDaA1cF7106bE861753B0f9F5c0250fE0819",
    "transactionHash":"0x5a9148aabbf7a026737d39b1567e920f3a4831af7033da479b1bf3ae33a66d2b"
    "amount":"100000000",
    "blockNumber":9051540,
    "proof": {
      "transactionHash":"0x5a9148aabbf7a026737d39b1567e920f3a4831af7033da479b1bf3ae33a66d2b",
      "blockTimestamp":1684873656,
      "blockNumber":9051540,
    },
    "claim": {
      "transactionHash":"0x5a9148aabbf7a026737d39b1567e920f3a4831af7033da479b1bf3ae33a66d2b",
      "blockTimestamp":1684873656,
      "blockNumber":9051540,
    },
    "withdrawalState": "COMPLETE",
    "l1Token": {
      "chainId":1,
      "address":"0x4242000000000000000000000000000000000042",
      "name":"Example",
      "symbol":"EXAMPLE",
      "decimals":18,
      "logoURI":"https://ethereum-optimism.github.io/data/OP/logo.svg",
      "extensions":{
        "optimismBridgeAddress":"0x636Af16bf2f682dD3109e60102b8E1A089FedAa8",
        "bridgeType": "STANDARD"
      }
    },
    "l2Token": {
      "chainId":10,
      "address":"0x0004242000000000000000000000000000000000",
      "name":"Example",
      "symbol":"EXAMPLE",
      "decimals":18,
      "logoURI":"https://ethereum-optimism.github.io/data/OP/logo.svg",
      "extensions":{
        "optimismBridgeAddress":"0x36Af16bf2f682dD3109e60102b8E1A089FedAa86",
        "bridgeType": "STANDARD"
      }
    },
  }
}
```

- **Typescript Types**

```typescript
import { Address } from 'wagmi'

/**
 * Optimism Tokenlist Type
 * @see https://github.com/ethereum-optimism/ethereum-optimism.github.io/blob/master/optimism.tokenlist.json
 */
export type TokenListItem = {
  chainId: number,
  address: Address,
  name: string,
  symbol: string,
  decimals: number,
  logoURI: "https://ethereum-optimism.github.io/data/ZRX/logo.png",
  extensions: {
    optimismBridgeAddress: Address
    bridgeType: 'STANDARD' | 'DAI' | 'SNX' | string
  }
}

/**
 * ERC20 Token withdrawal
 */
export type TokenWithdrawal = {
  guid: string;
  amount: string;
  blockNumber: number;
  blockTimestamp: number;
  from: string;
  to: string;
  transactionHash: string;
  l1Token: TokenListItem;
  l2Token: TokenListItem;
  withdrawalState: "UNCONFIRMED" | "FAILED" | "STATE_ROOT_NOT_PUBLISHED" | "READY_TO_PROVE" | "CHALLENGE_PERIOD" | "READY_FOR_RELAY" | "COMPLETE"
  proof?: {
    transactionHash: string,
    blockTimestamp: number,
    blockNumber: number,
  },
  claim?: {
    transactionHash: string,
    blockTimestamp: number,
    blockNumber: number,
  },
}

/**
 * Eth withdrawal
 * Same as ERC20 but no l1Token or l2Token properties
 */
type EthWithdrawal = {
  guid: string;
  amount: string;
  blockNumber: number;
  blockTimestamp: number;
  from: string;
  to: string;
  transactionHash: string;
  withdrawalState: "UNCONFIRMED" | "FAILED" | "STATE_ROOT_NOT_PUBLISHED" | "READY_TO_PROVE" | "CHALLENGE_PERIOD" | "READY_FOR_RELAY" | "COMPLETE"
  /**
   * only included if withdrawal is proved
   */
  proof?: {
    transactionHash: string,
    blockTimestamp: number,
    blockNumber: number,
  },
  /**
   * only included if withdrawal is claimed
   */
  claim?: {
    transactionHash: string,
    blockTimestamp: number,
    blockNumber: number,
  },
}

/**
 * The endpoint returns an array of Token and Eth withdrawal
 */
export type Withdrawal = TokenWithdrawal | EthWithdrawal

export type WithdrawalResponse = {
  data: Withdrawal[]
  cursor: string
  hasNextPage: boolean
}
