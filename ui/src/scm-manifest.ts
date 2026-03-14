// SPDX-Licence-Identifier: EUPL-1.2

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
      font-family: system-ui, -apple-system, sans-serif;
    }

    .manifest {
      border: 1px solid #e5e7eb;
      border-radius: 0.5rem;
      padding: 1.25rem;
      background: #fff;
    }

    .header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 1rem;
    }

    h3 {
      margin: 0;
      font-size: 1.125rem;
      font-weight: 600;
    }

    .version {
      font-size: 0.75rem;
      padding: 0.125rem 0.5rem;
      background: #f3f4f6;
      border-radius: 1rem;
      colour: #6b7280;
    }

    .field {
      margin-bottom: 0.75rem;
    }

    .field-label {
      font-size: 0.75rem;
      font-weight: 600;
      colour: #6b7280;
      text-transform: uppercase;
      letter-spacing: 0.025em;
      margin-bottom: 0.25rem;
    }

    .field-value {
      font-size: 0.875rem;
    }

    .code {
      font-family: monospace;
      font-size: 0.8125rem;
      background: #f9fafb;
      padding: 0.25rem 0.5rem;
      border-radius: 0.25rem;
    }

    .slots {
      display: grid;
      grid-template-columns: auto 1fr;
      gap: 0.25rem 1rem;
      font-size: 0.8125rem;
    }

    .slot-key {
      font-weight: 600;
      colour: #374151;
    }

    .slot-value {
      font-family: monospace;
      colour: #6b7280;
    }

    .permissions-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
      gap: 0.5rem;
    }

    .perm-group {
      border: 1px solid #e5e7eb;
      border-radius: 0.375rem;
      padding: 0.5rem;
    }

    .perm-group-label {
      font-size: 0.6875rem;
      font-weight: 700;
      colour: #6b7280;
      text-transform: uppercase;
      margin-bottom: 0.25rem;
    }

    .perm-item {
      font-size: 0.8125rem;
      font-family: monospace;
      colour: #374151;
    }

    .signature {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      margin-top: 1rem;
      padding: 0.75rem;
      border-radius: 0.375rem;
      font-size: 0.875rem;
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
      font-weight: 600;
      padding: 0.125rem 0.5rem;
      border-radius: 1rem;
    }

    .badge.verified {
      background: #dcfce7;
      colour: #166534;
    }

    .badge.unsigned {
      background: #fef3c7;
      colour: #92400e;
    }

    .badge.invalid {
      background: #fee2e2;
      colour: #991b1b;
    }

    .actions {
      margin-top: 1rem;
      display: flex;
      gap: 0.5rem;
    }

    .verify-input {
      flex: 1;
      padding: 0.375rem 0.75rem;
      border: 1px solid #d1d5db;
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      font-family: monospace;
    }

    button {
      padding: 0.375rem 1rem;
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      cursor: pointer;
      border: 1px solid #d1d5db;
      background: #fff;
      transition: background 0.15s;
    }

    button:hover {
      background: #f3f4f6;
    }

    button.primary {
      background: #6366f1;
      colour: #fff;
      border-colour: #6366f1;
    }

    button.primary:hover {
      background: #4f46e5;
    }

    .empty {
      text-align: center;
      padding: 2rem;
      colour: #9ca3af;
      font-size: 0.875rem;
    }

    .error {
      colour: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border-radius: 0.375rem;
      font-size: 0.875rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      colour: #6b7280;
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

    return html`
      ${this.error ? html`<div class="error">${this.error}</div>` : nothing}
      <div class="manifest">
        <div class="header">
          <div>
            <h3>${m.name}</h3>
            <span class="code">${m.code}</span>
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
                ${m.modules.map((mod) => html`<div class="code" style="margin-bottom:0.25rem">${mod}</div>`)}
              </div>
            `
          : nothing}

        <div class="signature ${hasSig ? (this.verifyResult ? (this.verifyResult.valid ? 'signed' : 'invalid') : 'signed') : 'unsigned'}">
          <span class="badge ${hasSig ? (this.verifyResult ? (this.verifyResult.valid ? 'verified' : 'invalid') : 'unsigned') : 'unsigned'}">
            ${hasSig ? (this.verifyResult ? (this.verifyResult.valid ? 'Verified' : 'Invalid') : 'Signed') : 'Unsigned'}
          </span>
          ${hasSig ? html`<span>Signature present</span>` : html`<span>No signature</span>`}
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
