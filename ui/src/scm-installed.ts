// SPDX-License-Identifier: EUPL-1.2

import { LitElement, html, css, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { ScmApi } from './shared/api.js';

interface InstalledModule {
  code: string;
  name: string;
  version: string;
  repo: string;
  entry_point: string;
  permissions: {
    read?: string[];
    write?: string[];
    net?: string[];
    run?: string[];
  };
  sign_key?: string;
  installed_at: string;
}

/**
 * <core-scm-installed> — Manage installed providers.
 * Displays installed provider list with update/remove actions.
 */
@customElement('core-scm-installed')
export class ScmInstalled extends LitElement {
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
      gap: 0.75rem;
    }

    .item {
      border: 1px solid #e2e8f0;
      border-radius: 1rem;
      padding: 1rem;
      background:
        linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.98));
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
      transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        border-color 0.15s ease;
    }

    .item:hover {
      box-shadow: 0 16px 36px rgba(15, 23, 42, 0.08);
      border-color: #c7d2fe;
      transform: translateY(-2px);
    }

    .item-info {
      flex: 1;
    }

    .item-name {
      font-weight: 800;
      font-size: 0.95rem;
      color: #0f172a;
    }

    .item-meta {
      font-size: 0.75rem;
      color: #64748b;
      margin-top: 0.35rem;
      display: flex;
      gap: 1rem;
      flex-wrap: wrap;
    }

    .item-code {
      font-family: monospace;
      color: #334155;
      font-weight: 700;
    }

    .item-repo,
    .item-entry {
      font-family: monospace;
      color: #475569;
      word-break: break-word;
    }

    .badge {
      display: inline-flex;
      align-items: center;
      gap: 0.3rem;
      font-size: 0.6875rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: #166534;
    }

    .badge::before {
      content: '';
      width: 0.45rem;
      height: 0.45rem;
      border-radius: 999px;
      background: #22c55e;
    }

    .badge.unsigned {
      color: #64748b;
    }

    .badge.unsigned::before {
      background: #f59e0b;
    }

    .item-actions {
      display: flex;
      gap: 0.5rem;
    }

    button {
      padding: 0.45rem 0.85rem;
      border-radius: 0.75rem;
      font-size: 0.8125rem;
      cursor: pointer;
      font-weight: 700;
      transition:
        background 0.15s ease,
        transform 0.15s ease,
        box-shadow 0.15s ease;
    }

    button.update {
      background: #fff;
      color: #4338ca;
      border: 1px solid #6366f1;
    }

    button.update:hover {
      background: #eef2ff;
      transform: translateY(-1px);
    }

    button.update:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    button.remove {
      background: #fff;
      color: #dc2626;
      border: 1px solid #dc2626;
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
      margin-bottom: 1rem;
    }

    @media (max-width: 720px) {
      .shell {
        padding: 0.875rem;
      }

      .summary,
      .item {
        flex-direction: column;
        align-items: flex-start;
      }

      .item-actions {
        width: 100%;
      }

      button {
        flex: 1;
      }
    }
  `;

  @property({ attribute: 'api-url' }) apiUrl = '';

  @state() private modules: InstalledModule[] = [];
  @state() private loading = true;
  @state() private error = '';
  @state() private updating = new Set<string>();

  private api!: ScmApi;

  connectedCallback() {
    super.connectedCallback();
    this.api = new ScmApi(this.apiUrl);
    this.loadInstalled();
  }

  async loadInstalled() {
    this.loading = true;
    this.error = '';
    try {
      this.modules = await this.api.installed();
    } catch (e: any) {
      this.error = e.message ?? 'Failed to load installed providers';
    } finally {
      this.loading = false;
    }
  }

  async refresh() {
    await this.loadInstalled();
  }

  private async handleUpdate(code: string) {
    this.updating = new Set([...this.updating, code]);
    try {
      await this.api.updateInstalled(code);
      await this.loadInstalled();
    } catch (e: any) {
      this.error = e.message ?? 'Update failed';
    } finally {
      const next = new Set(this.updating);
      next.delete(code);
      this.updating = next;
    }
  }

  private async handleRemove(code: string) {
    try {
      await this.api.remove(code);
      this.dispatchEvent(
        new CustomEvent('scm-removed', { detail: { code }, bubbles: true }),
      );
      await this.loadInstalled();
    } catch (e: any) {
      this.error = e.message ?? 'Removal failed';
    }
  }

  private formatDate(iso: string): string {
    try {
      return new Date(iso).toLocaleDateString('en-GB', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
      });
    } catch {
      return iso;
    }
  }

  render() {
    if (this.loading) {
      return html`<div class="loading">Loading installed providers\u2026</div>`;
    }

    return html`
      <div class="shell">
        <div class="summary">
          <div class="summary-copy">
            <span class="summary-title">Installed providers</span>
            <span class="summary-subtitle">
              Review local modules, update them, or remove stale installs.
            </span>
          </div>
          <div class="summary-copy" style="text-align:right">
            <span class="summary-title">${this.modules.length}</span>
            <span class="summary-subtitle">Installed</span>
          </div>
        </div>

        ${this.error ? html`<div class="error">${this.error}</div>` : nothing}
        ${this.modules.length === 0
          ? html`<div class="empty">No providers installed.</div>`
          : html`
              <div class="list">
                ${this.modules.map(
                  (mod) => html`
                    <div class="item">
                      <div class="item-info">
                        <div class="item-name">${mod.name}</div>
                        <div class="item-meta">
                          <span class="item-code">${mod.code}</span>
                          <span>v${mod.version}</span>
                          <span>Installed ${this.formatDate(mod.installed_at)}</span>
                        </div>
                        <div class="item-meta">
                          <span class="item-repo">${mod.repo}</span>
                          <span class="item-entry">entry: ${mod.entry_point}</span>
                        </div>
                        <div class="badge ${mod.sign_key ? '' : 'unsigned'}">
                          ${mod.sign_key ? 'Signed manifest' : 'Unsigned manifest'}
                        </div>
                      </div>
                      <div class="item-actions">
                        <button
                          class="update"
                          ?disabled=${this.updating.has(mod.code)}
                          @click=${() => this.handleUpdate(mod.code)}
                        >
                          ${this.updating.has(mod.code) ? 'Updating\u2026' : 'Update'}
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
    'core-scm-installed': ScmInstalled;
  }
}
