/**
 * Returns the explorer base URL for the current chain backend.
 * Reads NEXT_PUBLIC_CHAIN_BACKEND from env (set at build time).
 */
export function explorerBase(): string {
  const backend = process.env.NEXT_PUBLIC_CHAIN_BACKEND || "xdc";
  if (backend === "stellar") {
    return "https://testnet.stellar.expert/explorer/public";
  }
  return "https://testnet.xdcscan.com";
}

/** Build an explorer URL for a transaction hash. */
export function txExplorerUrl(txHash: string): string {
  const base = explorerBase();
  const backend = process.env.NEXT_PUBLIC_CHAIN_BACKEND || "xdc";
  if (backend === "stellar") {
    return `${base}/tx/${txHash}`;
  }
  return `${base}/tx/${txHash.replace(/^0x/, "")}`;
}

/** Build an explorer URL for an account/address. */
export function accountExplorerUrl(address: string): string {
  const base = explorerBase();
  const backend = process.env.NEXT_PUBLIC_CHAIN_BACKEND || "xdc";
  if (backend === "stellar") {
    return `${base}/account/${address}`;
  }
  const normalized = address.toLowerCase().startsWith("xdc")
    ? "0x" + address.slice(3)
    : address;
  return `${base}/address/${normalized}`;
}

/** Returns the explorer display name. */
export function explorerName(): string {
  const backend = process.env.NEXT_PUBLIC_CHAIN_BACKEND || "xdc";
  return backend === "stellar" ? "Stellar Expert" : "BlocksScan";
}
