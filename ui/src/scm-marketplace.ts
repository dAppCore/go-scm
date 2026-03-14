// SPDX-Licence-Identifier: EUPL-1.2

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
      font-family: system-ui, -apple-system, sans-serif;
    }

    .toolbar {
      display: flex;
      gap: 0.5rem;
      align-items: center;
      margin-bottom: 1rem;
    }

    .search {
      flex: 1;
      padding: 0.5rem 0.75rem;
      border: 1px solid #d1d5db;
      border-radius: 0.375rem;
      font-size: 0.875rem;
      outline: none;
    }

    .search:focus {
      border-colour: #6366f1;
      box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.2);
    }

    .categories {
      display: flex;
      gap: 0.25rem;
      flex-wrap: wrap;
    }

    .category-btn {
      padding: 0.25rem 0.75rem;
      border: 1px solid #e5e7eb;
      border-radius: 1rem;
      background: #fff;
      font-size: 0.75rem;
      cursor: pointer;
      transition: all 0.15s;
    }

    .category-btn:hover {
      background: #f3f4f6;
    }

    .category-btn.active {
      background: #6366f1;
      colour: #fff;
      border-colour: #6366f1;
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
      gap: 1rem;
      margin-top: 1rem;
    }

    .card {
      border: 1px solid #e5e7eb;
      border-radius: 0.5rem;
      padding: 1rem;
      background: #fff;
      transition: box-shadow 0.15s;
    }

    .card:hover {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 0.5rem;
    }

    .card-name {
      font-weight: 600;
      font-size: 0.9375rem;
    }

    .card-code {
      font-size: 0.75rem;
      colour: #6b7280;
      font-family: monospace;
    }

    .card-category {
      font-size: 0.6875rem;
      padding: 0.125rem 0.5rem;
      background: #f3f4f6;
      border-radius: 1rem;
      colour: #6b7280;
    }

    .card-actions {
      margin-top: 0.75rem;
      display: flex;
      gap: 0.5rem;
    }

    button.install {
      padding: 0.375rem 1rem;
      background: #6366f1;
      colour: #fff;
      border: none;
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      cursor: pointer;
      transition: background 0.15s;
    }

    button.install:hover {
      background: #4f46e5;
    }

    button.install:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    button.remove {
      padding: 0.375rem 1rem;
      background: #fff;
      colour: #dc2626;
      border: 1px solid #dc2626;
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      cursor: pointer;
    }

    button.remove:hover {
      background: #fef2f2;
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
      <div class="toolbar">
        <input
          type="text"
          class="search"
          placeholder="Search providers\u2026"
          .value=${this.searchQuery}
          @input=${this.handleSearch}
        />
      </div>

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
                      <div class="card-header">
                        <div>
                          <div class="card-name">${mod.name}</div>
                          <div class="card-code">${mod.code}</div>
                        </div>
                        ${mod.category
                          ? html`<span class="card-category">${mod.category}</span>`
                          : nothing}
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
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'core-scm-marketplace': ScmMarketplace;
  }
}
