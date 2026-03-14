// SPDX-Licence-Identifier: EUPL-1.2

import { LitElement, html, css, nothing } from 'lit';
import { customElement, property, state } from 'lit/decorators.js';
import { connectScmEvents, type ScmEvent } from './shared/events.js';

// Side-effect imports to register child elements
import './scm-marketplace.js';
import './scm-installed.js';
import './scm-manifest.js';
import './scm-registry.js';

type TabId = 'marketplace' | 'installed' | 'manifest' | 'registry';

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
      font-family: system-ui, -apple-system, sans-serif;
      height: 100%;
      background: #fafafa;
    }

    /* H — Header */
    .header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0.75rem 1rem;
      background: #fff;
      border-bottom: 1px solid #e5e7eb;
    }

    .title {
      font-weight: 700;
      font-size: 1rem;
      colour: #111827;
    }

    .refresh-btn {
      padding: 0.375rem 0.75rem;
      border: 1px solid #d1d5db;
      border-radius: 0.375rem;
      background: #fff;
      font-size: 0.8125rem;
      cursor: pointer;
      transition: background 0.15s;
    }

    .refresh-btn:hover {
      background: #f3f4f6;
    }

    /* H-L — Tabs */
    .tabs {
      display: flex;
      gap: 0;
      background: #fff;
      border-bottom: 1px solid #e5e7eb;
      padding: 0 1rem;
    }

    .tab {
      padding: 0.625rem 1rem;
      font-size: 0.8125rem;
      font-weight: 500;
      colour: #6b7280;
      cursor: pointer;
      border-bottom: 2px solid transparent;
      transition: all 0.15s;
      background: none;
      border-top: none;
      border-left: none;
      border-right: none;
    }

    .tab:hover {
      colour: #374151;
    }

    .tab.active {
      colour: #6366f1;
      border-bottom-colour: #6366f1;
    }

    /* C — Content */
    .content {
      flex: 1;
      padding: 1rem;
      overflow-y: auto;
    }

    /* F — Footer / Status bar */
    .footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0.5rem 1rem;
      background: #fff;
      border-top: 1px solid #e5e7eb;
      font-size: 0.75rem;
      colour: #9ca3af;
    }

    .ws-status {
      display: flex;
      align-items: center;
      gap: 0.375rem;
    }

    .ws-dot {
      width: 0.5rem;
      height: 0.5rem;
      border-radius: 50%;
    }

    .ws-dot.connected {
      background: #22c55e;
    }

    .ws-dot.disconnected {
      background: #ef4444;
    }

    .ws-dot.idle {
      background: #d1d5db;
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

  disconnectedCallback() {
    super.disconnectedCallback();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  private connectWs() {
    this.ws = connectScmEvents(this.wsUrl, (event: ScmEvent) => {
      this.lastEvent = event.channel ?? event.type ?? '';
      this.requestUpdate();
    });
    this.ws.onopen = () => {
      this.wsConnected = true;
    };
    this.ws.onclose = () => {
      this.wsConnected = false;
    };
  }

  private handleTabClick(tab: TabId) {
    this.activeTab = tab;
  }

  private handleRefresh() {
    // Force re-render of active child by toggling a key
    const content = this.shadowRoot?.querySelector('.content');
    if (content) {
      const child = content.firstElementChild;
      if (child && 'loadModules' in child) {
        (child as any).loadModules();
      } else if (child && 'loadInstalled' in child) {
        (child as any).loadInstalled();
      } else if (child && 'loadManifest' in child) {
        (child as any).loadManifest();
      } else if (child && 'loadRegistry' in child) {
        (child as any).loadRegistry();
      }
    }
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
        <span class="title">SCM</span>
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
