// SPDX-License-Identifier: EUPL-1.2

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
      font-family:
        Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI',
        sans-serif;
      color: #111827;
    }

    .shell {
      background: rgba(255, 255, 255, 0.84);
      border: 1px solid rgba(226, 232, 240, 0.95);
      border-radius: 1rem;
      padding: 1rem;
      box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
      backdrop-filter: blur(12px);
    }

    .summary {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
      margin-bottom: 1rem;
      padding-bottom: 0.75rem;
      border-bottom: 1px solid #e2e8f0;
    }

    .summary-copy {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }

    .summary-title {
      font-size: 1rem;
      font-weight: 800;
      color: #0f172a;
    }

    .summary-subtitle {
      font-size: 0.8125rem;
      color: #64748b;
    }

    .list {
      display: flex;
      flex-direction: column;
      gap: 0.625rem;
    }

    .repo {
      border: 1px solid #e2e8f0;
      border-radius: 1rem;
      padding: 0.75rem 1rem;
      background:
        linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.98));
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
    }

    .repo-info {
      flex: 1;
    }

    .repo-name {
      font-weight: 800;
      font-size: 0.95rem;
      font-family: monospace;
      color: #0f172a;
    }

    .repo-desc {
      font-size: 0.8125rem;
      color: #64748b;
      margin-top: 0.125rem;
    }

    .repo-meta {
      display: flex;
      gap: 0.5rem;
      align-items: center;
      margin-top: 0.25rem;
      flex-wrap: wrap;
    }

    .type-badge {
      font-size: 0.6875rem;
      padding: 0.2rem 0.5rem;
      border-radius: 999px;
      font-weight: 800;
      text-transform: uppercase;
      letter-spacing: 0.04em;
    }

    .type-badge.foundation {
      background: #dbeafe;
      color: #1e40af;
    }

    .type-badge.module {
      background: #f3e8ff;
      color: #6b21a8;
    }

    .type-badge.product {
      background: #dcfce7;
      color: #166534;
    }

    .type-badge.template {
      background: #fef3c7;
      color: #92400e;
    }

    .deps {
      font-size: 0.75rem;
      color: #64748b;
    }

    .path {
      font-size: 0.75rem;
      font-family: monospace;
      color: #475569;
      word-break: break-word;
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
      box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.15);
    }

    .status-dot.missing {
      background: #ef4444;
      box-shadow: 0 0 0 4px rgba(239, 68, 68, 0.14);
    }

    .status-label {
      font-size: 0.75rem;
      color: #64748b;
      font-weight: 700;
    }

    .empty {
      text-align: center;
      padding: 2rem;
      color: #64748b;
      font-size: 0.875rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      color: #64748b;
    }

    .error {
      color: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border: 1px solid #fecaca;
      border-radius: 0.75rem;
      font-size: 0.875rem;
      margin-bottom: 1rem;
    }

    @media (max-width: 720px) {
      .shell {
        padding: 0.875rem;
      }

      .summary,
      .repo {
        flex-direction: column;
        align-items: flex-start;
      }
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

  async refresh() {
    await this.loadRegistry();
  }

  render() {
    if (this.loading) {
      return html`<div class="loading">Loading registry\u2026</div>`;
    }

    return html`
      <div class="shell">
        <div class="summary">
          <div class="summary-copy">
            <span class="summary-title">Registry</span>
            <span class="summary-subtitle">
              Workspace repositories and dependency order from repos.yaml.
            </span>
          </div>
          <div class="summary-copy" style="text-align:right">
            <span class="summary-title">${this.repos.length}</span>
            <span class="summary-subtitle">Entries</span>
          </div>
        </div>

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
                        ${repo.path ? html`<div class="path">${repo.path}</div>` : nothing}
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
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'core-scm-registry': ScmRegistry;
  }
}
