import { describe, expect, it } from 'vitest';
import { makeToken } from '../test/makeToken';
import { decodeJwtPayload, isTokenExpired } from './jwt';

describe('decodeJwtPayload', () => {
  it('decodes a well-formed token payload', () => {
    const token = makeToken({ is_admin: true, exp: 12345 });
    expect(decodeJwtPayload(token)).toEqual({ is_admin: true, exp: 12345 });
  });

  it('returns null for a token with no payload segment', () => {
    expect(decodeJwtPayload('notatoken')).toBeNull();
  });

  it('returns null for a payload segment that is not valid base64/JSON', () => {
    expect(decodeJwtPayload('header.@@@not-base64@@@.sig')).toBeNull();
  });

  it('returns null for an empty string', () => {
    expect(decodeJwtPayload('')).toBeNull();
  });
});

describe('isTokenExpired', () => {
  it('treats a null token as expired', () => {
    expect(isTokenExpired(null)).toBe(true);
  });

  it('treats an empty string token as expired', () => {
    expect(isTokenExpired('')).toBe(true);
  });

  it('returns true for a token whose exp claim is in the past', () => {
    const pastExp = Math.floor(Date.now() / 1000) - 3600; // 1 hour ago
    const token = makeToken({ is_admin: true, exp: pastExp });
    expect(isTokenExpired(token)).toBe(true);
  });

  it('returns false for a token whose exp claim is in the future', () => {
    const futureExp = Math.floor(Date.now() / 1000) + 3600; // 1 hour from now
    const token = makeToken({ is_admin: true, exp: futureExp });
    expect(isTokenExpired(token)).toBe(false);
  });

  it('treats a token at exactly its exp boundary as expired', () => {
    const nowSeconds = Math.floor(Date.now() / 1000);
    const token = makeToken({ is_admin: true, exp: nowSeconds });
    expect(isTokenExpired(token)).toBe(true);
  });

  it('returns false (defers to backend) when the token has no exp claim', () => {
    const token = makeToken({ is_admin: true });
    expect(isTokenExpired(token)).toBe(false);
  });

  it('returns false (defers to backend) when the token cannot be decoded', () => {
    expect(isTokenExpired('garbage')).toBe(false);
  });
});
