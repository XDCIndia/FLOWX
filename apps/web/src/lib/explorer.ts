/**
 * Returns the explorer base URL for the XDC Apothem testnet.
 */
export function explorerBase(): string {
  return "https://testnet.xdcscan.com";
}

/** Build an explorer URL for a transaction hash. */
export function txExplorerUrl(txHash: string): string {
  return `${explorerBase()}/tx/${txHash.replace(/^0x/, "")}`;
}

/** Build an explorer URL for an account/address. */
export function accountExplorerUrl(address: string): string {
  const normalized = address.toLowerCase().startsWith("xdc")
    ? "0x" + address.slice(3)
    : address;
  return `${explorerBase()}/address/${normalized}`;
}

/** Returns the explorer display name. */
export function explorerName(): string {
  return "XDCScan";
}
