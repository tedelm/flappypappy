const CACHE_NAME = "flappy-beer-__BUILD_ID__";

const PRECACHE_URLS = [
  "./",
  "./index.html",
  "./manifest.webmanifest",
  "./wasm_exec.js",
  "./flappy.wasm",
  "./icons/icon-192.png",
  "./icons/icon-512.png",
];

const NETWORK_FIRST = ["index.html", "flappy.wasm", "wasm_exec.js", "config.js"];

async function cacheFirst(request) {
  const cached = await caches.match(request);
  if (cached) {
    return cached;
  }
  return fetch(request);
}

async function networkFirst(request) {
  try {
    const response = await fetch(request);
    if (response && response.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, response.clone());
    }
    return response;
  } catch (_) {
    const cached = await caches.match(request);
    if (cached) {
      return cached;
    }
    throw _;
  }
}

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(PRECACHE_URLS))
  );
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key))
      )
    )
  );
  self.clients.claim();
});

self.addEventListener("fetch", (event) => {
  if (event.request.method !== "GET") {
    return;
  }
  const url = new URL(event.request.url);
  if (NETWORK_FIRST.some((name) => url.pathname.endsWith(name))) {
    event.respondWith(networkFirst(event.request));
    return;
  }
  event.respondWith(cacheFirst(event.request));
});
