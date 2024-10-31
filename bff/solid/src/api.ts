const getAccessKey = () => {
  // In production, you might want to get this from a more secure source
  return process.env.VITE_ACCESS_KEY || "DEFAULT-ACCESS-KEY";
};

export const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
  const headers = new Headers(options.headers || {});
  headers.set("Autodba-Access-Key", getAccessKey());

  const response = await fetch(url, {
    ...options,
    headers,
  });

  return response;
};
