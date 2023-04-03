This code interacts with layer 2 through the `engine_api`. It facilitates the following transaction types:

- Deposit ETH
- Deposit ERC20 (Boba)
- Normal Transaction
- Submit Transaction to `TX Pool`

To submit transactions, the `engine_api` follows these steps:

1. Use `engine_forkchoiceUpdatedV1` to obtain the next `payloadID`.
2. Use `engine_getPayloadV1` to retrieve the next block data. The `parent.hash` value should be identical to our latest block hash.
3. Use `engine_newPayloadV1` to validate or execute our payload from `engine_getPayloadV1`.
4. Use `engine_forkchoiceUpdatedV1` to update our next block. During this process, the `attribute` is set to `nil`.