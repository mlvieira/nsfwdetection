import { writable } from 'svelte/store';

export const pendingRatings = writable(new Set());
export const completedRatings = writable(new Set());