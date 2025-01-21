import { writable } from 'svelte/store';
import { uploads, newUploads } from "../stores/uploads"
import { completedRatings, pendingRatings } from '../stores/ratings';
import {
  updateImageLabel,
  deleteImageFromStore,
  addNewUploads,
} from "../utils/uploads";

export const wsMessages = writable([]);
const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
export const wsUrl = `${protocol}://${window.location.host}/ws`;

export const isWebSocketConnected = writable(false);

let socket = null;
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
const reconnectDelay = 3000;

const handlers = {
  'ack_rating': handleAckRating,
  'in_progress': handleInProgress,
  'new_upload': handleNewImage,
  'ack_delete': handleAckDelete,
  'delete': handleDelete,
  'rate': handleRate,
  'update_rate': handleUpdateRate,
  'error': handleError
};

export function initWebSocket(token) {
  if (!token || token.startsWith('ws')) {
    console.error('[WebSocket] Invalid token!');
    return;
  }

  if (socket) {
    console.warn('[WebSocket] Closing previous WebSocket connection...');
    socket.close();
    socket = null;
  }

  console.log('[WebSocket] Connecting to:', wsUrl);

  socket = new WebSocket(wsUrl, [token]);

  socket.onopen = () => {
    console.log('[WebSocket] Connected');
    isWebSocketConnected.set(true);
    reconnectAttempts = 0;
  };

  socket.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);

      const handler = handlers[msg.event];
      if (handler) {
        handler(msg);
      } else {
        console.warn('[WebSocket] Unhandled event type:', msg.event);
      }
    } catch (err) {
      console.error('[WebSocket] JSON parse error', err);
    }
  };

  socket.onerror = (err) => {
    console.error('[WebSocket] Error:', err);
  };

  socket.onclose = () => {
    console.log('[WebSocket] Disconnected');
    socket = null;
    isWebSocketConnected.set(false);

    if (reconnectAttempts < maxReconnectAttempts) {
      reconnectAttempts++;
      console.log(`[WebSocket] Attempting reconnect (${reconnectAttempts}/${maxReconnectAttempts})...`);
      setTimeout(() => initWebSocket(token), reconnectDelay);
    } else {
      console.error('[WebSocket] Max reconnect attempts reached. Connection failed.');
    }
  };
}

export function closeWebSocket() {
  if (socket) {
    socket.close();
    socket = null;
    isWebSocketConnected.set(false);
    console.log('[WebSocket] Connection closed');
  }
}

function handleInProgress(msg) {
  console.log('[WebSocket] In Progress:', msg);
  pendingRatings.update((set) => set.add(msg.sha256));
}

function handleAckRating(msg) {
  console.log('[WebSocket] ACK Rating:', msg);

  if (msg.status === "success") {
    pendingRatings.update((set) => {
      set.delete(msg.sha256);
      return set;
    });

    completedRatings.update((set) => set.add(msg.sha256));
    updateImageLabel(uploads, msg.sha256, msg.rating);
  } else {
    console.error('[WebSocket] Rating failed for hash:', msg.sha256);
  }

  wsMessages.update((prev) => [...prev, msg]);
}

function handleError(msg) {
  console.error('[WebSocket] Error:', msg);
  wsMessages.update((prev) => [...prev, msg]);
}

function handleNewImage(msg) {
  addNewUploads(newUploads, [msg.data]);
}

function handleDelete(msg) {
  deleteImageFromStore(uploads, msg.sha256);
}

function handleAckDelete(msg) {
  if (msg.status === "success") {
    deleteImageFromStore(uploads, msg.sha256);
  } else {
    console.error('[WebSocket] Delete failed for hash:', msg.sha256);
  }
}

function handleRate(msg) {
  updateImageLabel(uploads, msg.sha256, msg.rating);
}

function handleUpdateRate(msg) {
  updateImageLabel(uploads, msg.sha256, msg.rating);
  completedRatings.update((set) => set.add(msg.sha256));
}