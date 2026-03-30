const DEFAULT_API_PROXY_TARGET = "http://127.0.0.1:8080";
const HOP_BY_HOP_HEADERS = new Set([
  "connection",
  "content-length",
  "host",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailer",
  "transfer-encoding",
  "upgrade",
]);

type RouteContext = {
  params: {
    path: string[];
  };
};

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

export async function GET(request: Request, context: RouteContext) {
  return proxyRequest(request, context);
}

export async function POST(request: Request, context: RouteContext) {
  return proxyRequest(request, context);
}

export async function PUT(request: Request, context: RouteContext) {
  return proxyRequest(request, context);
}

export async function PATCH(request: Request, context: RouteContext) {
  return proxyRequest(request, context);
}

export async function DELETE(request: Request, context: RouteContext) {
  return proxyRequest(request, context);
}

export async function OPTIONS(request: Request, context: RouteContext) {
  return proxyRequest(request, context);
}

async function proxyRequest(request: Request, context: RouteContext) {
  const requestURL = new URL(request.url);
  const upstreamURL = buildUpstreamURL(context.params.path ?? [], requestURL.search);
  const headers = new Headers(request.headers);
  headers.set("x-forwarded-host", requestURL.host);
  headers.set("x-forwarded-proto", requestURL.protocol.replace(":", ""));

  for (const header of HOP_BY_HOP_HEADERS) {
    headers.delete(header);
  }

  const init: RequestInit = {
    method: request.method,
    headers,
    redirect: "manual",
    cache: "no-store",
  };

  if (!["GET", "HEAD"].includes(request.method.toUpperCase())) {
    init.body = await request.arrayBuffer();
  }

  try {
    const upstream = await fetch(upstreamURL, init);
    const responseHeaders = new Headers(upstream.headers);
    for (const header of HOP_BY_HOP_HEADERS) {
      responseHeaders.delete(header);
    }
    responseHeaders.set("cache-control", "no-store");
    return new Response(upstream.body, {
      status: upstream.status,
      statusText: upstream.statusText,
      headers: responseHeaders,
    });
  } catch (error) {
    return Response.json(
      {
        code: 502,
        message: error instanceof Error ? error.message : "api proxy request failed",
      },
      { status: 502 },
    );
  }
}

function buildUpstreamURL(pathSegments: string[], search: string) {
  const base = process.env.API_PROXY_TARGET || DEFAULT_API_PROXY_TARGET;
  const url = new URL(base);
  const joined = pathSegments.filter(Boolean).join("/");
  url.pathname = `/api/${joined}`.replace(/\/+/g, "/");
  url.search = search;
  return url;
}
