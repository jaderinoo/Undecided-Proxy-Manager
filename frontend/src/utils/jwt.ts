// Decodes a JWT payload without verifying the signature. Signature
// verification happens server-side; this is only used to make a fast,
// local guess at whether a token has expired, before or between backend
// round-trips.
export function decodeJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const payload = token.split('.')[1];
    if (!payload) return null;

    const base64 = payload.replace(/-/g, '+').replace(/_/g, '/');
    const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), '=');
    const json = decodeURIComponent(
      atob(padded)
        .split('')
        .map(c => '%' + c.charCodeAt(0).toString(16).padStart(2, '0'))
        .join('')
    );
    return JSON.parse(json);
  } catch {
    return null;
  }
}

// Returns true only when the token's `exp` claim has definitively passed.
// If the token can't be decoded or has no `exp` claim, returns false so
// callers fall back to an authoritative backend check instead of guessing.
export function isTokenExpired(token: string | null): boolean {
  if (!token) return true;

  const payload = decodeJwtPayload(token);
  const exp = payload?.exp;
  if (typeof exp !== 'number') return false;

  return Date.now() >= exp * 1000;
}
