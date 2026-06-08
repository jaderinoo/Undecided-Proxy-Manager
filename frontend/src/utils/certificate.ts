export function getDaysUntilExpiry(expiresAt: string): number {
  const now = new Date();
  const expiry = new Date(expiresAt);
  return Math.ceil(
    (expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  );
}

export function isExpiringSoon(expiresAt: string): boolean {
  return getDaysUntilExpiry(expiresAt) <= 30;
}

export function isCertificateExpired(expiresAt: string): boolean {
  return getDaysUntilExpiry(expiresAt) <= 0;
}
