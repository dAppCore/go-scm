// SPDX-License-Identifier: EUPL-1.2

import { LitElement, html, css, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import type { PropertyValues } from 'lit';
import { connectScmEvents, type ScmEvent } from './shared/events.js';

// Side-effect imports to register child elements
import './scm-marketplace.js';
import './scm-installed.js';
import './scm-manifest.js';
import './scm-registry.js';

type TabId = 'marketplace' | 'installed' | 'manifest' | 'registry';
type RefreshableElement = HTMLElement & {
  refresh?: () => Promise<void> | void;
};

/**
 * <core-scm-panel> — Top-level HLCRF panel with tabs.
 *
 * Arranges child elements in HLCRF layout:
 * - H: Title bar with refresh button
 * - H-L: Navigation tabs
 * - C: Active tab content (one of the child elements)
 * - F: Status bar (connection state, last refresh)
 */
@customElement('core-scm-panel')
export class ScmPanel extends LitElement {
  static styles = css`
    :host {
      display: flex;
      flex-direction: column;
      font-family:
        Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI',
        sans-serif;
      height: 100%;
      background:
        radial-gradient(circle at top left, rgba(99, 102, 241, 0.12), transparent 30%),
        linear-gradient(180deg, #eef2ff 0%, #f8fafc 28%, #f3f4f6 100%);
      color: #111827;
    }

    /* H — Header */
    .header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
      padding: 1rem 1.25rem;
      background: rgba(255, 255, 255, 0.86);
      backdrop-filter: blur(18px);
      border-bottom: 1px solid rgba(226, 232, 240, 0.9);
    }

    .title-wrap {
      display: flex;
      flex-direction: column;
      gap: 0.125rem;
    }

    .title {
      font-size: 1rem;
      font-weight: 800;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      color: #0f172a;
    }

    .subtitle {
      font-size: 0.8125rem;
      color: #64748b;
    }

    .refresh-btn {
      padding: 0.5rem 0.875rem;
      border: 1px solid rgba(99, 102, 241, 0.25);
      border-radius: 999px;
      background: linear-gradient(180deg, #ffffff, #eef2ff);
      color: #4338ca;
      font-weight: 600;
      font-size: 0.8125rem;
      cursor: pointer;
      transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        background 0.15s ease;
      box-shadow: 0 1px 1px rgba(15, 23, 42, 0.04);
    }

    .refresh-btn:hover {
      background: linear-gradient(180deg, #ffffff, #e0e7ff);
      transform: translateY(-1px);
      box-shadow: 0 8px 20px rgba(99, 102, 241, 0.12);
    }

    /* H-L — Tabs */
    .tabs {
      display: flex;
      gap: 0.375rem;
      padding: 0.75rem 1rem 0;
      background: rgba(255, 255, 255, 0.72);
      backdrop-filter: blur(18px);
      border-bottom: 1px solid rgba(226, 232, 240, 0.9);
      overflow-x: auto;
    }

    .tab {
      padding: 0.7rem 1rem;
      font-size: 0.8125rem;
      font-weight: 700;
      letter-spacing: 0.01em;
      color: #64748b;
      cursor: pointer;
      border: 1px solid transparent;
      border-radius: 999px 999px 0 0;
      transition:
        color 0.15s ease,
        background 0.15s ease,
        border-color 0.15s ease,
        transform 0.15s ease;
      background: transparent;
    }

    .tab:hover {
      color: #334155;
      transform: translateY(-1px);
    }

    .tab.active {
      color: #4338ca;
      background: rgba(255, 255, 255, 0.96);
      border-color: rgba(226, 232, 240, 0.9);
      border-bottom-color: rgba(255, 255, 255, 0.96);
      box-shadow: 0 -1px 0 rgba(255, 255, 255, 0.6), 0 -8px 24px rgba(15, 23, 42, 0.04);
    }

    /* C — Content */
    .content {
      flex: 1;
      padding: 1.25rem;
      overflow-y: auto;
      display: flex;
      justify-content: center;
      align-items: flex-start;
    }

    .content > * {
      width: min(100%, 1120px);
    }

    /* F — Footer / Status bar */
    .footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
      padding: 0.75rem 1.25rem;
      background: rgba(255, 255, 255, 0.84);
      backdrop-filter: blur(18px);
      border-top: 1px solid rgba(226, 232, 240, 0.9);
      font-size: 0.75rem;
      color: #64748b;
    }

    .ws-status {
      display: flex;
      align-items: center;
      gap: 0.375rem;
      font-weight: 600;
    }

    .ws-dot {
      width: 0.5rem;
      height: 0.5rem;
      border-radius: 50%;
    }

    .ws-dot.connected {
      background: #22c55e;
      box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.15);
    }

    .ws-dot.disconnected {
      background: #ef4444;
      box-shadow: 0 0 0 4px rgba(239, 68, 68, 0.14);
    }

    .ws-dot.idle {
      background: #d1d5db;
    }

    @media (max-width: 720px) {
      .header,
      .footer {
        flex-direction: column;
        align-items: flex-start;
      }

      .tabs {
        padding-inline: 0.75rem;
      }

      .content {
        padding: 0.875rem;
      }
    }
  `;

  @property({ attribute: 'api-url' }) apiUrl = '';
  @property({ attribute: 'ws-url' }) wsUrl = '';

  @state() private activeTab: TabId = 'marketplace';
  @state() private wsConnected = false;
  @state() private lastEvent = '';

  private ws: WebSocket | null = null;

  connectedCallback() {
    super.connectedCallback();
    if (this.wsUrl) {
      this.connectWs();
    }
  }

  updated(changedProperties: PropertyValues<this>) {
    super.updated(changedProperties);
    if (changedProperties.has('wsUrl') && this.isConnected) {
      this.connectWs();
    }
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this.disconnectWs();
  }

  private connectWs() {
    this.disconnectWs();
    if (!this.wsUrl) {
      return;
    }

    this.ws = connectScmEvents(this.wsUrl, (event: ScmEvent) => {
      this.lastEvent = event.channel ?? event.type ?? '';
      this.requestUpdate();
      this.refreshForEvent(event);
    });
    this.ws.onopen = () => {
      this.wsConnected = true;
    };
    this.ws.onclose = () => {
      this.wsConnected = false;
    };
  }

  private disconnectWs() {
    if (!this.ws) {
      return;
    }

    this.ws.close();
    this.ws = null;
  }

  private handleTabClick(tab: TabId) {
    this.activeTab = tab;
  }

  private async handleRefresh() {
    await this.refreshActiveTab();
  }

  private refreshForEvent(event: ScmEvent) {
    const targets = this.tabsForChannel(event.channel ?? event.type ?? '');
    if (targets.includes(this.activeTab)) {
      void this.refreshActiveTab();
    }
  }

  private tabsForChannel(channel: string): TabId[] {
    if (channel.startsWith('scm.marketplace.')) {
      return ['marketplace', 'installed'];
    }
    if (channel.startsWith('scm.installed.')) {
      return ['installed'];
    }
    if (channel === 'scm.manifest.verified') {
      return ['manifest'];
    }
    if (channel === 'scm.registry.changed') {
      return ['registry'];
    }
    return [];
  }

  private async refreshActiveTab() {
    const child = this.shadowRoot?.querySelector('.content > *') as RefreshableElement | null;
    if (!child?.refresh) {
      return;
    }

    await child.refresh();
  }

  private renderContent() {
    switch (this.activeTab) {
      case 'marketplace':
        return html`<core-scm-marketplace api-url=${this.apiUrl}></core-scm-marketplace>`;
      case 'installed':
        return html`<core-scm-installed api-url=${this.apiUrl}></core-scm-installed>`;
      case 'manifest':
        return html`<core-scm-manifest api-url=${this.apiUrl}></core-scm-manifest>`;
      case 'registry':
        return html`<core-scm-registry api-url=${this.apiUrl}></core-scm-registry>`;
      default:
        return nothing;
    }
  }

  private tabs: { id: TabId; label: string }[] = [
    { id: 'marketplace', label: 'Marketplace' },
    { id: 'installed', label: 'Installed' },
    { id: 'manifest', label: 'Manifest' },
    { id: 'registry', label: 'Registry' },
  ];

  render() {
    const wsState = this.wsUrl
      ? this.wsConnected
        ? 'connected'
        : 'disconnected'
      : 'idle';

    return html`
      <div class="header">
        <div class="title-wrap">
          <span class="title">SCM</span>
          <span class="subtitle">Marketplace, manifests, installed modules, and registry status</span>
        </div>
        <button class="refresh-btn" @click=${this.handleRefresh}>Refresh</button>
      </div>

      <div class="tabs">
        ${this.tabs.map(
          (tab) => html`
            <button
              class="tab ${this.activeTab === tab.id ? 'active' : ''}"
              @click=${() => this.handleTabClick(tab.id)}
            >
              ${tab.label}
            </button>
          `,
        )}
      </div>

      <div class="content">${this.renderContent()}</div>

      <div class="footer">
        <div class="ws-status">
          <span class="ws-dot ${wsState}"></span>
          <span>${wsState === 'connected' ? 'Connected' : wsState === 'disconnected' ? 'Disconnected' : 'No WebSocket'}</span>
        </div>
        ${this.lastEvent ? html`<span>Last: ${this.lastEvent}</span>` : nothing}
      </div>
    `;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    'core-scm-panel': ScmPanel;
  }
}
