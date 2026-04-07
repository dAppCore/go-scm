// SPDX-License-Identifier: EUPL-1.2

export interface ScmEvent {
  type: string;
  channel?: string;
  data?: any;
  timestamp?: string;
}

/**
 * Connects to a WebSocket endpoint and dispatches SCM events to a handler.
 * Returns the WebSocket instance for lifecycle management.
 */
export function connectScmEvents(
  wsUrl: string,
  handler: (event: ScmEvent) => void,
): WebSocket {
  const ws = new WebSocket(wsUrl);

  ws.onmessage = (e: MessageEvent) => {
    try {
      const event: ScmEvent = JSON.parse(e.data);
      if (event.type?.startsWith?.('scm.') || event.channel?.startsWith?.('scm.')) {
        handler(event);
      }
    } catch {
      // Ignore malformed messages
    }
  };

  return ws;
}
