import { getApiBaseUrl, TEST_USER } from "../helpers/env";

/**
 * API fixture for direct backend HTTP calls.
 * Useful for test data setup/teardown that bypasses the UI.
 */
export class ApiFixture {
  private baseUrl: string;
  private token: string | null = null;

  constructor() {
    this.baseUrl = getApiBaseUrl();
  }

  /** Authenticate and store the JWT token. Returns the full login response. */
  async login(
    email: string = TEST_USER.email,
    password: string = TEST_USER.password
  ): Promise<{ token: string; refresh_token: string; user: unknown }> {
    const res = await this.fetchWithRetry(`${this.baseUrl}/api/v1/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) throw new Error(`Login failed: ${res.status}`);
    const data = await res.json();
    this.token = data.token;
    return data;
  }

  /** Switch authentication to a different user. */
  async loginAs(email: string, password: string): Promise<string> {
    const data = await this.login(email, password);
    return data.token;
  }

  /** GET request with authentication. */
  async get(path: string): Promise<Response> {
    await this.ensureToken();
    return fetch(`${this.baseUrl}${path}`, {
      headers: { Authorization: `Bearer ${this.token}` },
    });
  }

  /** POST request with authentication. */
  async post(path: string, body?: unknown): Promise<Response> {
    await this.ensureToken();
    return fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.token}`,
      },
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  /** PUT request with authentication. */
  async put(path: string, body: unknown): Promise<Response> {
    await this.ensureToken();
    return fetch(`${this.baseUrl}${path}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.token}`,
      },
      body: JSON.stringify(body),
    });
  }

  /** PATCH request with authentication. */
  async patch(path: string, body: unknown): Promise<Response> {
    await this.ensureToken();
    return fetch(`${this.baseUrl}${path}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.token}`,
      },
      body: JSON.stringify(body),
    });
  }

  /** DELETE request with authentication. */
  async delete(path: string): Promise<Response> {
    await this.ensureToken();
    return fetch(`${this.baseUrl}${path}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${this.token}` },
    });
  }

  /** POST without authentication (for public endpoints like register). */
  async postPublic(path: string, body: unknown): Promise<Response> {
    return this.fetchWithRetry(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
  }

  /** POST with a specific token (for token-dependent flows). */
  async postWithToken(path: string, body: unknown, token: string): Promise<Response> {
    return this.fetchWithRetry(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  /** GET with a specific token. */
  async getWithToken(path: string, token: string): Promise<Response> {
    return fetch(`${this.baseUrl}${path}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  }

  /**
   * Retry-aware fetch: automatically retries on 429 (rate limited).
   * Used internally by public endpoint methods that hit auth rate limits.
   */
  private async fetchWithRetry(
    url: string,
    init: RequestInit,
    maxRetries = 3
  ): Promise<Response> {
    for (let i = 0; i <= maxRetries; i++) {
      const res = await fetch(url, init);
      if (res.status !== 429 || i === maxRetries) return res;
      const delay = Math.min(1000 * 2 ** i, 5000);
      await new Promise((r) => setTimeout(r, delay));
    }
    return fetch(url, init); // unreachable, satisfies TS
  }

  private async ensureToken(): Promise<void> {
    if (!this.token) await this.login();
  }
}
