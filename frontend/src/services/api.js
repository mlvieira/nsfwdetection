const BASE_URL = import.meta.env.VITE_BASE_URL || 'http://localhost:3001';

async function handleFetch(url, options) {
  const res = await fetch(url, options);
  if (!res.ok) {
    const errorMessage = (await res.json())?.error || "Request failed";
    throw new Error(errorMessage);
  }

  return await res.json();
}

export async function loginUser(username, password) {
  const url = `${BASE_URL}/admin/login`;
  return handleFetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
}

export async function fetchUploads(cursorId, limit, reviewed, jwtToken) {
  const url = `${BASE_URL}/admin/images`
  const body = {
    id: cursorId,
    limit: limit,
  }

  if (reviewed !== undefined) {
    body.reviewed = reviewed;
  }

  return handleFetch(url, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${jwtToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  });
}

export async function labelImage(hash, rating, jwtToken) {
  const url = `${BASE_URL}/admin/label/add/${hash}`;
  return handleFetch(url, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${jwtToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      event: 'rate',
      rating,
      sha256: hash,
    }),
  });
}

export async function updateLabel(hash, rating, jwtToken) {
  const url = `${BASE_URL}/admin/label/update/${hash}`;
  return handleFetch(url, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${jwtToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      event: 'update_rate',
      rating,
      sha256: hash,
    }),
  });
}

export async function deleteImage(hash, jwtToken) {
  const url = `${BASE_URL}/admin/delete/${hash}`;
  return handleFetch(url, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${jwtToken}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      event: 'delete',
      sha256: hash,
    }),
  });
}

export async function stats(jwtToken) {
  const url = `${BASE_URL}/admin/stats`;
  return handleFetch(url, {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${jwtToken}`,
    },
  });
}