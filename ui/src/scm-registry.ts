// SPDX-Licence-Identifier: EUPL-1.2

import { LitElement, html, css, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { ScmApi } from './shared/api.js';

interface RepoSummary {
  name: string;
  type: string;
  description?: string;
  depends_on?: string[];
  path?: string;
  exists: boolean;
}

/**
 * <core-scm-registry> — Show repos.yaml registry status.
 * Read-only display of repository status.
 */
@customElement('core-scm-registry')
export class ScmRegistry extends LitElement {
  static styles = css`
    :host {
      display: block;
      font-family: system-ui, -apple-system, sans-serif;
    }

    .list {
      display: flex;
      flex-direction: column;
      gap: 0.375rem;
    }

    .repo {
      border: 1px solid #e5e7eb;
      border-radius: 0.5rem;
      padding: 0.75rem 1rem;
      background: #fff;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .repo-info {
      flex: 1;
    }

    .repo-name {
      font-weight: 600;
      font-size: 0.9375rem;
      font-family: monospace;
    }

    .repo-desc {
      font-size: 0.8125rem;
      colour: #6b7280;
      margin-top: 0.125rem;
    }

    .repo-meta {
      display: flex;
      gap: 0.5rem;
      align-items: center;
      margin-top: 0.25rem;
    }

    .type-badge {
      font-size: 0.6875rem;
      padding: 0.0625rem 0.5rem;
      border-radius: 1rem;
      font-weight: 600;
    }

    .type-badge.foundation {
      background: #dbeafe;
      colour: #1e40af;
    }

    .type-badge.module {
      background: #f3e8ff;
      colour: #6b21a8;
    }

    .type-badge.product {
      background: #dcfce7;
      colour: #166534;
    }

    .type-badge.template {
      background: #fef3c7;
      colour: #92400e;
    }

    .deps {
      font-size: 0.75rem;
      colour: #9ca3af;
    }

    .status {
      display: flex;
      align-items: center;
      gap: 0.375rem;
    }

    .status-dot {
      width: 0.5rem;
      height: 0.5rem;
      border-radius: 50%;
    }

    .status-dot.present {
      background: #22c55e;
    }

    .status-dot.missing {
      background: #ef4444;
    }

    .status-label {
      font-size: 0.75rem;
      colour: #6b7280;
    }

    .empty {
      text-align: center;
      padding: 2rem;
      colour: #9ca3af;
      font-size: 0.875rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      colour: #6b7280;
    }

    .error {
      colour: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border-radius: 0.375rem;
      font-size: 0.875rem;
      margin-bottom: 1rem;
    }
  `;

  @property({ attribute: 'api-url' }) apiUrl = '';

  @state() private repos: RepoSummary[] = [];
  @state() private loading = true;
  @state() private error = '';

  private api!: ScmApi;

  connectedCallback() {
    super.connectedCallback();
    this.api = new ScmApi(this.apiUrl);
    this.loadRegistry();
  }

  async loadRegistry() {
    this.loading = true;
    this.error = '';
    try {
      this.repos = await this.api.registry();
    } catch (e: any) {
      this.error = e.message ?? 'Failed to load registry';
    } finally {
      this.loading = false;
    }
  }

  render() {
    if (this.loading) {
      return html`<div class="loading">Loading registry\u2026</div>`;
    }

    return html`
      ${this.error ? html`<div class="error">${this.error}</div>` : nothing}
      ${this.repos.length === 0
        ? html`<div class="empty">No repositories in registry.</div>`
        : html`
            <div class="list">
              ${this.repos.map(
                (repo) => html`
                  <div class="repo">
                    <div class="repo-info">
                      <div class="repo-name">${repo.name}</div>
                      ${repo.description
                        ? html`<div class="repo-desc">${repo.description}</div>`
                        : nothing}
                      <div class="repo-meta">
                        <span class="type-badge ${repo.type}">${repo.type}</span>
                        ${repo.depends_on && repo.depends_on.length > 0
                          ? html`<span class="deps">depends: ${repo.depends_on.join(', ')}</span>`
                          : nothing}
                      </div>
                    </div>
                    <div class="status">
                      <span class="status-dot ${repo.exists ? 'present' : 'missing'}"></span>
                      <span class="status-label">${repo.exists ? 'Present' : 'Missing'}</span>
                    </div>
                  </div>
                `,
              )}
            </div>
          `}
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'core-scm-registry': ScmRegistry;
  }
}
