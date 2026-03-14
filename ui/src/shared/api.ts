// SPDX-Licence-Identifier: EUPL-1.2

/**
 * ScmApi provides a typed fetch wrapper for the /api/v1/scm/* endpoints.
 */
export class ScmApi {
  constructor(private baseUrl: string = '') {}

  private get base(): string {
    return `${this.baseUrl}/api/v1/scm`;
  }

  private async request<T>(path: string, opts?: RequestInit): Promise<T> {
    const res = await fetch(`${this.base}${path}`, opts);
    const json = await res.json();
    if (!json.success) {
      throw new Error(json.error?.message ?? 'Request failed');
    }
    return json.data as T;
  }

  marketplace(query?: string, category?: string) {
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    if (category) params.set('category', category);
    const qs = params.toString();
    return this.request<any[]>(`/marketplace${qs ? `?${qs}` : ''}`);
  }

  marketplaceItem(code: string) {
    return this.request<any>(`/marketplace/${code}`);
  }

  install(code: string) {
    return this.request<any>(`/marketplace/${code}/install`, { method: 'POST' });
  }

  remove(code: string) {
    return this.request<any>(`/marketplace/${code}`, { method: 'DELETE' });
  }

  installed() {
    return this.request<any[]>('/installed');
  }

  updateInstalled(code: string) {
    return this.request<any>(`/installed/${code}/update`, { method: 'POST' });
  }

  manifest() {
    return this.request<any>('/manifest');
  }

  verify(publicKey: string) {
    return this.request<any>('/manifest/verify', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ public_key: publicKey }),
    });
  }

  sign(privateKey: string) {
    return this.request<any>('/manifest/sign', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ private_key: privateKey }),
    });
  }

  permissions() {
    return this.request<any>('/manifest/permissions');
  }

  registry() {
    return this.request<any[]>('/registry');
  }
}
