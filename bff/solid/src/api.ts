const getAccessKey = () => {
  // In a real app, you might want to get this from a more secure source
  return import.meta.env.VITE_ACCESS_KEY || 'DEFAULT-ACCESS-KEY';
};

export const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
  const headers = new Headers(options.headers || {});
  headers.set('ACCESS_KEY', getAccessKey());

  const response = await fetch(url, {
    ...options,
    headers,
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return response.json();
};
