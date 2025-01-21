import { writable } from 'svelte/store';

export const token = writable(localStorage.getItem('token') || '');

token.subscribe((val) => {
  if (val) {
    localStorage.setItem('token', val);
  } else {
    localStorage.removeItem('token');
  }
});
