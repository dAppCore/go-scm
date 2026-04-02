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
      font-family: system-ui, -apple-system, sans-serif;
    }

    .list {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }

    .item {
      border: 1px solid #e5e7eb;
      border-radius: 0.5rem;
      padding: 1rem;
      background: #fff;
      display: flex;
      justify-content: space-between;
      align-items: center;
      transition: box-shadow 0.15s;
    }

    .item:hover {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    }

    .item-info {
      flex: 1;
    }

    .item-name {
      font-weight: 600;
      font-size: 0.9375rem;
    }

    .item-meta {
      font-size: 0.75rem;
      colour: #6b7280;
      margin-top: 0.25rem;
      display: flex;
      gap: 1rem;
    }

    .item-code {
      font-family: monospace;
    }

    .item-actions {
      display: flex;
      gap: 0.5rem;
    }

    button {
      padding: 0.375rem 0.75rem;
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      cursor: pointer;
      transition: background 0.15s;
    }

    button.update {
      background: #fff;
      colour: #6366f1;
      border: 1px solid #6366f1;
    }

    button.update:hover {
      background: #eef2ff;
    }

    button.update:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    button.remove {
      background: #fff;
      colour: #dc2626;
      border: 1px solid #dc2626;
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
      margin-bottom: 1rem;
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
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'core-scm-installed': ScmInstalled;
  }
}
