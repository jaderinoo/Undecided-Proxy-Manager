// Builds an unsigned, well-formed-looking JWT for tests. Only the payload
// shape matters here (decodeJwtPayload/isTokenExpired never verify the
// signature), so ASCII-only payloads are enough — plain btoa is sufficient.
function base64url(json: string): string {
  return btoa(json).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

export function makeToken(payload: Record<string, unknown>): string {
  const header = base64url(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const body = base64url(JSON.stringify(payload));
  return `${header}.${body}.fakesignature`;
}
