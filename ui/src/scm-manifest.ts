// SPDX-License-Identifier: EUPL-1.2

import { LitElement, html, css, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { ScmApi } from './shared/api.js';

interface ManifestData {
  code: string;
  name: string;
  description?: string;
  version: string;
  sign?: string;
  layout?: string;
  slots?: Record<string, string>;
  permissions?: {
    read?: string[];
    write?: string[];
    net?: string[];
    run?: string[];
  };
  modules?: string[];
}

/**
 * <core-scm-manifest> — View and verify a .core/manifest.yaml file.
 */
@customElement('core-scm-manifest')
export class ScmManifest extends LitElement {
  static styles = css`
    :host {
      display: block;
      font-family:
        Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI',
        sans-serif;
      color: #111827;
    }

    .manifest {
      border: 1px solid rgba(226, 232, 240, 0.95);
      border-radius: 1rem;
      padding: 1.25rem;
      background:
        linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.98));
      box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
    }

    .header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 1rem;
      margin-bottom: 1rem;
      padding-bottom: 0.9rem;
      border-bottom: 1px solid #e2e8f0;
    }

    .header-copy {
      display: flex;
      flex-direction: column;
      gap: 0.35rem;
    }

    h3 {
      margin: 0;
      font-size: 1.2rem;
      font-weight: 800;
      color: #0f172a;
    }

    .version {
      font-size: 0.75rem;
      padding: 0.25rem 0.625rem;
      background: #eef2ff;
      border-radius: 999px;
      color: #4338ca;
      font-weight: 700;
    }

    .meta-row {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
    }

    .meta-chip {
      display: inline-flex;
      align-items: center;
      gap: 0.35rem;
      padding: 0.25rem 0.6rem;
      border-radius: 999px;
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      font-size: 0.6875rem;
      font-weight: 700;
      color: #475569;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    .field {
      margin-bottom: 0.875rem;
    }

    .field-label {
      font-size: 0.75rem;
      font-weight: 800;
      color: #64748b;
      text-transform: uppercase;
      letter-spacing: 0.025em;
      margin-bottom: 0.25rem;
    }

    .field-value {
      font-size: 0.875rem;
      color: #0f172a;
    }

    .code {
      font-family: monospace;
      font-size: 0.8125rem;
      background: #f8fafc;
      padding: 0.35rem 0.55rem;
      border-radius: 0.45rem;
      border: 1px solid #e2e8f0;
      color: #1f2937;
    }

    .slots {
      display: grid;
      grid-template-columns: auto 1fr;
      gap: 0.25rem 1rem;
      font-size: 0.8125rem;
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      border-radius: 0.75rem;
      padding: 0.75rem;
    }

    .slot-key {
      font-weight: 700;
      color: #334155;
    }

    .slot-value {
      font-family: monospace;
      color: #64748b;
    }

    .permissions-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
      gap: 0.5rem;
    }

    .perm-group {
      border: 1px solid #e2e8f0;
      border-radius: 0.75rem;
      padding: 0.5rem;
      background: #fff;
    }

    .perm-group-label {
      font-size: 0.6875rem;
      font-weight: 800;
      color: #64748b;
      text-transform: uppercase;
      margin-bottom: 0.25rem;
    }

    .perm-item {
      font-size: 0.8125rem;
      font-family: monospace;
      color: #374151;
      word-break: break-word;
    }

    .signature {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      margin-top: 1rem;
      padding: 0.75rem;
      border-radius: 0.75rem;
      font-size: 0.875rem;
      font-weight: 600;
    }

    .signature.signed {
      background: #f0fdf4;
      border: 1px solid #bbf7d0;
    }

    .signature.unsigned {
      background: #fffbeb;
      border: 1px solid #fde68a;
    }

    .signature.invalid {
      background: #fef2f2;
      border: 1px solid #fecaca;
    }

    .badge {
      font-size: 0.75rem;
      font-weight: 800;
      padding: 0.25rem 0.55rem;
      border-radius: 999px;
      text-transform: uppercase;
      letter-spacing: 0.04em;
    }

    .badge.verified {
      background: #dcfce7;
      color: #166534;
    }

    .badge.signed {
      background: #e0e7ff;
      color: #4338ca;
    }

    .badge.unsigned {
      background: #fef3c7;
      color: #92400e;
    }

    .badge.invalid {
      background: #fee2e2;
      color: #991b1b;
    }

    .actions {
      margin-top: 1rem;
      display: flex;
      gap: 0.5rem;
      flex-wrap: wrap;
    }

    .verify-input {
      flex: 1;
      min-width: 14rem;
      padding: 0.6rem 0.85rem;
      border: 1px solid #cbd5e1;
      border-radius: 0.75rem;
      font-size: 0.8125rem;
      font-family: monospace;
      background: #fff;
    }

    button {
      padding: 0.6rem 1rem;
      border-radius: 0.75rem;
      font-size: 0.8125rem;
      cursor: pointer;
      border: 1px solid #d1d5db;
      background: #fff;
      transition:
        background 0.15s ease,
        transform 0.15s ease;
    }

    button:hover {
      background: #f3f4f6;
      transform: translateY(-1px);
    }

    button.primary {
      background: linear-gradient(180deg, #6366f1, #4f46e5);
      color: #fff;
      border-color: #6366f1;
    }

    button.primary:hover {
      background: linear-gradient(180deg, #4f46e5, #4338ca);
    }

    .empty {
      text-align: center;
      padding: 2rem;
      color: #64748b;
      font-size: 0.875rem;
    }

    .error {
      color: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border: 1px solid #fecaca;
      border-radius: 0.75rem;
      font-size: 0.875rem;
      margin-bottom: 0.75rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      color: #64748b;
    }

    .status {
      display: flex;
      align-items: center;
      gap: 0.35rem;
    }

    .status-dot {
      width: 0.5rem;
      height: 0.5rem;
      border-radius: 999px;
    }

    .status-dot.verified {
      background: #22c55e;
    }

    .status-dot.signed {
      background: #6366f1;
    }

    .status-dot.unsigned {
      background: #f59e0b;
    }

    .status-dot.invalid {
      background: #ef4444;
    }

    @media (max-width: 720px) {
      .manifest {
        padding: 1rem;
      }

      .header {
        flex-direction: column;
        align-items: flex-start;
      }

      .actions {
        flex-direction: column;
      }
    }
  `;

  @property({ attribute: 'api-url' }) apiUrl = '';
  @property() path = '';

  @state() private manifest: ManifestData | null = null;
  @state() private loading = true;
  @state() private error = '';
  @state() private verifyKey = '';
  @state() private verifyResult: { valid: boolean } | null = null;

  private api!: ScmApi;

  connectedCallback() {
    super.connectedCallback();
    this.api = new ScmApi(this.apiUrl);
    this.loadManifest();
  }

  async loadManifest() {
    this.loading = true;
    this.error = '';
    try {
      this.manifest = await this.api.manifest();
    } catch (e: any) {
      this.error = e.message ?? 'Failed to load manifest';
    } finally {
      this.loading = false;
    }
  }

  async refresh() {
    this.verifyResult = null;
    await this.loadManifest();
  }

  private async handleVerify() {
    if (!this.verifyKey.trim()) return;
    try {
      this.verifyResult = await this.api.verify(this.verifyKey.trim());
    } catch (e: any) {
      this.error = e.message ?? 'Verification failed';
    }
  }

  private async handleSign() {
    const key = prompt('Enter hex-encoded Ed25519 private key:');
    if (!key) return;
    try {
      await this.api.sign(key);
      await this.loadManifest();
    } catch (e: any) {
      this.error = e.message ?? 'Signing failed';
    }
  }

  private renderPermissions(perms: ManifestData['permissions']) {
    if (!perms) return nothing;

    const groups = [
      { label: 'Read', items: perms.read },
      { label: 'Write', items: perms.write },
      { label: 'Network', items: perms.net },
      { label: 'Run', items: perms.run },
    ].filter((g) => g.items && g.items.length > 0);

    if (groups.length === 0) return nothing;

    return html`
      <div class="field">
        <div class="field-label">Permissions</div>
        <div class="permissions-grid">
          ${groups.map(
            (g) => html`
              <div class="perm-group">
                <div class="perm-group-label">${g.label}</div>
                ${g.items!.map((item) => html`<div class="perm-item">${item}</div>`)}
              </div>
            `,
          )}
        </div>
      </div>
    `;
  }

  render() {
    if (this.loading) {
      return html`<div class="loading">Loading manifest\u2026</div>`;
    }
    if (this.error && !this.manifest) {
      return html`<div class="error">${this.error}</div>`;
    }
    if (!this.manifest) {
      return html`<div class="empty">No manifest found. Create a .core/manifest.yaml to get started.</div>`;
    }

    const m = this.manifest;
    const hasSig = !!m.sign;
    const status = hasSig
      ? this.verifyResult
        ? this.verifyResult.valid
          ? 'verified'
          : 'invalid'
        : 'signed'
      : 'unsigned';
    const statusLabel = hasSig
      ? this.verifyResult
        ? this.verifyResult.valid
          ? 'Verified'
          : 'Invalid'
        : 'Signed'
      : 'Unsigned';
    const permissionCount =
      (m.permissions?.read?.length ?? 0) +
      (m.permissions?.write?.length ?? 0) +
      (m.permissions?.net?.length ?? 0) +
      (m.permissions?.run?.length ?? 0);

    return html`
      ${this.error ? html`<div class="error">${this.error}</div>` : nothing}
      <div class="manifest">
        <div class="header">
          <div class="header-copy">
            <h3>${m.name}</h3>
            <div class="meta-row">
              <span class="meta-chip">${m.code}</span>
              <span class="meta-chip">${permissionCount} permissions</span>
              ${m.modules?.length
                ? html`<span class="meta-chip">${m.modules.length} modules</span>`
                : nothing}
            </div>
          </div>
          <span class="version">v${m.version}</span>
        </div>

        ${m.description
          ? html`
              <div class="field">
                <div class="field-label">Description</div>
                <div class="field-value">${m.description}</div>
              </div>
            `
          : nothing}
        ${m.layout
          ? html`
              <div class="field">
                <div class="field-label">Layout</div>
                <div class="field-value code">${m.layout}</div>
              </div>
            `
          : nothing}
        ${m.slots && Object.keys(m.slots).length > 0
          ? html`
              <div class="field">
                <div class="field-label">Slots</div>
                <div class="slots">
                  ${Object.entries(m.slots).map(
                    ([k, v]) => html`
                      <span class="slot-key">${k}</span>
                      <span class="slot-value">${v}</span>
                    `,
                  )}
                </div>
              </div>
            `
          : nothing}

        ${this.renderPermissions(m.permissions)}
        ${m.modules && m.modules.length > 0
          ? html`
              <div class="field">
                <div class="field-label">Modules</div>
                ${m.modules.map((mod) => html`<div class="code" style="margin-bottom:0.35rem">${mod}</div>`)}
              </div>
            `
          : nothing}

        <div class="signature ${status}">
          <span class="badge ${status}">${statusLabel}</span>
          <span class="status">
            <span class="status-dot ${status}"></span>
            ${hasSig ? html`<span>Signature present</span>` : html`<span>No signature</span>`}
          </span>
        </div>

        <div class="actions">
          <input
            type="text"
            class="verify-input"
            placeholder="Public key (hex)\u2026"
            .value=${this.verifyKey}
            @input=${(e: Event) => (this.verifyKey = (e.target as HTMLInputElement).value)}
          />
          <button @click=${this.handleVerify}>Verify</button>
          <button class="primary" @click=${this.handleSign}>Sign</button>
        </div>
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'core-scm-manifest': ScmManifest;
  }
}
