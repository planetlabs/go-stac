import useSWR from 'swr';

function getProxyUrl() {
  const proxyPath = import.meta.env.VITE_PROXY;
  if (!proxyPath) {
    return 'http://localhost:8000/';
  }
  const url = new URL(proxyPath, location.href);
  return url.href;
}

export const proxyUrl = getProxyUrl();

function getResourceUrl(href) {
  const resource = new URL(href, proxyUrl);
  return resource.href;
}

function useFetch() {
  return function (url, options = {}) {
    return fetch(url, options);
  };
}

function useFetchJSON(key) {
  const fetcher = useFetch();

  return async function (url, options = {}) {
    const headers = {...options.headers, 'content-type': 'application/json'};
    const response = await fetcher(url, {
      ...options,
      body: options.body ? JSON.stringify(options.body) : null,
      headers,
    });
    if (!response.ok) {
      let message;
      try {
        const body = await response.json();
        if (body && body.message) {
          message = body.message;
        }
      } catch (err) {
        // pass
      }
      if (!message) {
        message = `Unexpected response: ${response.statusText}`;
      }
      throw new Error(message);
    }
    if (response.status === 204) {
      return null;
    }
    const body = await response.json();
    return key ? body[key] : body;
  };
}

export function useProxy(path) {
  const fetcher = useFetchJSON();
  return useSWR(getResourceUrl(path), fetcher);
}
