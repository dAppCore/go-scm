// SPDX-License-Identifier: EUPL-1.2

import { LitElement, html, css, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { ScmApi } from './shared/api.js';

interface Module {
  code: string;
  name: string;
  repo: string;
  sign_key: string;
  category: string;
}

/**
 * <core-scm-marketplace> — Browse, search, and install providers
 * from the SCM marketplace.
 */
@customElement('core-scm-marketplace')
export class ScmMarketplace extends LitElement {
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

    .summary-stats {
      display: flex;
      gap: 0.5rem;
      flex-wrap: wrap;
      justify-content: flex-end;
    }

    .stat {
      min-width: 4.5rem;
      padding: 0.55rem 0.75rem;
      border-radius: 0.85rem;
      background: linear-gradient(180deg, #f8fafc, #eef2ff);
      border: 1px solid #e2e8f0;
      text-align: center;
    }

    .stat-value {
      display: block;
      font-size: 1rem;
      font-weight: 800;
      color: #312e81;
      line-height: 1;
    }

    .stat-label {
      display: block;
      margin-top: 0.25rem;
      font-size: 0.6875rem;
      letter-spacing: 0.06em;
      text-transform: uppercase;
      color: #64748b;
    }

    .toolbar {
      display: flex;
      gap: 0.5rem;
      align-items: center;
      margin-bottom: 1rem;
      flex-wrap: wrap;
    }

    .search {
      flex: 1;
      min-width: 14rem;
      padding: 0.7rem 0.9rem;
      border: 1px solid #cbd5e1;
      border-radius: 0.85rem;
      font-size: 0.875rem;
      outline: none;
      background: #fff;
      transition:
        border-color 0.15s ease,
        box-shadow 0.15s ease;
    }

    .search:focus {
      border-color: #6366f1;
      box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.12);
    }

    .categories {
      display: flex;
      gap: 0.25rem;
      flex-wrap: wrap;
    }

    .category-btn {
      padding: 0.35rem 0.8rem;
      border: 1px solid #e2e8f0;
      border-radius: 1rem;
      background: #fff;
      font-size: 0.75rem;
      font-weight: 700;
      color: #475569;
      cursor: pointer;
      transition:
        background 0.15s ease,
        color 0.15s ease,
        border-color 0.15s ease,
        transform 0.15s ease;
    }

    .category-btn:hover {
      background: #f8fafc;
      transform: translateY(-1px);
    }

    .category-btn.active {
      background: linear-gradient(180deg, #6366f1, #4f46e5);
      color: #fff;
      border-color: #4f46e5;
      box-shadow: 0 6px 16px rgba(99, 102, 241, 0.22);
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
      gap: 0.875rem;
      margin-top: 1rem;
    }

    .card {
      border: 1px solid #e2e8f0;
      border-radius: 1rem;
      padding: 1rem;
      background:
        linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.98));
      transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        border-color 0.15s ease;
      min-height: 10rem;
      display: flex;
      flex-direction: column;
      justify-content: space-between;
    }

    .card:hover {
      box-shadow: 0 16px 36px rgba(15, 23, 42, 0.08);
      border-color: #c7d2fe;
      transform: translateY(-2px);
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 0.75rem;
      margin-bottom: 0.75rem;
    }

    .card-name {
      font-weight: 800;
      font-size: 0.95rem;
      color: #0f172a;
      line-height: 1.2;
    }

    .card-code {
      margin-top: 0.15rem;
      font-size: 0.775rem;
      color: #64748b;
      font-family: monospace;
    }

    .card-category {
      font-size: 0.6875rem;
      padding: 0.2rem 0.5rem;
      background: #e0e7ff;
      border-radius: 999px;
      color: #4338ca;
      font-weight: 700;
      white-space: nowrap;
    }

    .card-repo {
      font-size: 0.7875rem;
      color: #475569;
      font-family: monospace;
      word-break: break-word;
      margin-bottom: 0.4rem;
    }

    .card-sign {
      display: inline-flex;
      align-items: center;
      gap: 0.35rem;
      font-size: 0.6875rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: #166534;
      margin-bottom: 0.4rem;
    }

    .card-sign::before {
      content: '';
      width: 0.45rem;
      height: 0.45rem;
      border-radius: 999px;
      background: #22c55e;
    }

    .card-sign.unsigned {
      color: #64748b;
    }

    .card-sign.unsigned::before {
      background: #f59e0b;
    }

    .card-actions {
      margin-top: 0.75rem;
      display: flex;
      gap: 0.5rem;
    }

    button.install {
      padding: 0.45rem 1rem;
      background: linear-gradient(180deg, #6366f1, #4f46e5);
      color: #fff;
      border: none;
      border-radius: 0.75rem;
      font-size: 0.8125rem;
      font-weight: 700;
      cursor: pointer;
      transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        background 0.15s ease;
      box-shadow: 0 8px 16px rgba(99, 102, 241, 0.2);
    }

    button.install:hover {
      background: linear-gradient(180deg, #4f46e5, #4338ca);
      transform: translateY(-1px);
    }

    button.install:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    button.remove {
      padding: 0.45rem 1rem;
      background: #fff;
      color: #dc2626;
      border: 1px solid #dc2626;
      border-radius: 0.75rem;
      font-size: 0.8125rem;
      font-weight: 700;
      cursor: pointer;
      transition:
        background 0.15s ease,
        transform 0.15s ease;
    }

    button.remove:hover {
      background: #fef2f2;
      transform: translateY(-1px);
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
    }

    @media (max-width: 720px) {
      .shell {
        padding: 0.875rem;
        border-radius: 0.875rem;
      }

      .summary {
        flex-direction: column;
        align-items: flex-start;
      }

      .summary-stats {
        justify-content: flex-start;
      }

      .toolbar {
        flex-direction: column;
        align-items: stretch;
      }
    }
  `;

  @property({ attribute: 'api-url' }) apiUrl = '';
  @property() category = '';

  @state() private modules: Module[] = [];
  @state() private categories: string[] = [];
  @state() private searchQuery = '';
  @state() private activeCategory = '';
  @state() private loading = true;
  @state() private error = '';
  @state() private installing = new Set<string>();

  private api!: ScmApi;

  connectedCallback() {
    super.connectedCallback();
    this.api = new ScmApi(this.apiUrl);
    this.activeCategory = this.category;
    this.loadModules();
  }

  async loadModules() {
    this.loading = true;
    this.error = '';
    try {
      this.modules = await this.api.marketplace(
        this.searchQuery || undefined,
        this.activeCategory || undefined,
      );
      // Extract unique categories
      const cats = new Set<string>();
      this.modules.forEach((m) => {
        if (m.category) cats.add(m.category);
      });
      this.categories = Array.from(cats).sort();
    } catch (e: any) {
      this.error = e.message ?? 'Failed to load marketplace';
    } finally {
      this.loading = false;
    }
  }

  async refresh() {
    await this.loadModules();
  }

  private handleSearch(e: Event) {
    this.searchQuery = (e.target as HTMLInputElement).value;
    this.loadModules();
  }

  private handleCategoryClick(cat: string) {
    this.activeCategory = this.activeCategory === cat ? '' : cat;
    this.loadModules();
  }

  private async handleInstall(code: string) {
    this.installing = new Set([...this.installing, code]);
    try {
      await this.api.install(code);
      this.dispatchEvent(
        new CustomEvent('scm-installed', { detail: { code }, bubbles: true }),
      );
    } catch (e: any) {
      this.error = e.message ?? 'Installation failed';
    } finally {
      const next = new Set(this.installing);
      next.delete(code);
      this.installing = next;
    }
  }

  private async handleRemove(code: string) {
    try {
      await this.api.remove(code);
      this.dispatchEvent(
        new CustomEvent('scm-removed', { detail: { code }, bubbles: true }),
      );
    } catch (e: any) {
      this.error = e.message ?? 'Removal failed';
    }
  }

  render() {
    return html`
      <div class="shell">
        <div class="summary">
          <div class="summary-copy">
            <span class="summary-title">Marketplace</span>
            <span class="summary-subtitle">
              Browse, filter, and install providers from the current index.
            </span>
          </div>
          <div class="summary-stats">
            <div class="stat">
              <span class="stat-value">${this.modules.length}</span>
              <span class="stat-label">Results</span>
            </div>
            <div class="stat">
              <span class="stat-value">${this.categories.length}</span>
              <span class="stat-label">Categories</span>
            </div>
          </div>
        </div>

        <div class="toolbar">
          <input
            type="text"
            class="search"
            placeholder="Search providers\u2026"
            .value=${this.searchQuery}
            @input=${this.handleSearch}
          />
          ${this.categories.length > 0
            ? html`
                <div class="categories">
                  ${this.categories.map(
                    (cat) => html`
                      <button
                        class="category-btn ${this.activeCategory === cat ? 'active' : ''}"
                        @click=${() => this.handleCategoryClick(cat)}
                      >
                        ${cat}
                      </button>
                    `,
                  )}
                </div>
              `
            : nothing}
        </div>

        ${this.error ? html`<div class="error">${this.error}</div>` : nothing}
        ${this.loading
          ? html`<div class="loading">Loading marketplace\u2026</div>`
          : this.modules.length === 0
            ? html`<div class="empty">No providers found.</div>`
            : html`
                <div class="grid">
                  ${this.modules.map(
                    (mod) => html`
                      <div class="card">
                        <div>
                          <div class="card-header">
                            <div>
                              <div class="card-name">${mod.name}</div>
                              <div class="card-code">${mod.code}</div>
                            </div>
                            ${mod.category
                              ? html`<span class="card-category">${mod.category}</span>`
                              : nothing}
                          </div>
                          <div class="card-repo">${mod.repo}</div>
                          <div class="card-sign ${mod.sign_key ? '' : 'unsigned'}">
                            ${mod.sign_key ? 'Signed module' : 'Unsigned module'}
                          </div>
                        </div>
                        <div class="card-actions">
                          <button
                            class="install"
                            ?disabled=${this.installing.has(mod.code)}
                            @click=${() => this.handleInstall(mod.code)}
                          >
                            ${this.installing.has(mod.code) ? 'Installing\u2026' : 'Install'}
                          </button>
                          <button class="remove" @click=${() => this.handleRemove(mod.code)}>
                            Remove
                          </button>
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
    'core-scm-marketplace': ScmMarketplace;
  }
}
