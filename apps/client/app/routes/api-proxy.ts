import type { ActionFunctionArgs, LoaderFunctionArgs } from "react-router";

const BACKEND_URL = process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8080/api/v1";

async function handleRequest(request: Request, params: any) {
  const url = new URL(request.url);
  // Extract the path after /api/v1
  const path = params["*"];
  const targetUrl = `${BACKEND_URL}/${path}${url.search}`;

  const headers = new Headers(request.headers);
  
  // Forward cookies from the request to the backend
  // In React Router 7, cookies are already in the headers
  
  // Remove host header to avoid conflicts
  headers.delete("host");

  try {
    const response = await fetch(targetUrl, {
      method: request.method,
      headers: headers,
      body: request.method !== "GET" && request.method !== "HEAD" ? await request.blob() : undefined,
    });

    // Create response headers
    const responseHeaders = new Headers();
    response.headers.forEach((value, key) => {
      // Forward Set-Cookie and other safe headers
      if (
        key.toLowerCase() === "set-cookie" ||
        key.toLowerCase() === "content-type" ||
        key.toLowerCase() === "cache-control"
      ) {
        responseHeaders.append(key, value);
      }
    });

    return new Response(await response.blob(), {
      status: response.status,
      headers: responseHeaders,
    });
  } catch (error) {
    console.error("Proxy Error (Backend Unreachable):", error);
    return new Response(
      JSON.stringify({
        success: false,
        code: "BACKEND_OFFLINE",
        message: "Gagal terhubung ke API Server.",
      }),
      { status: 502, headers: { "Content-Type": "application/json" } }
    );
  }
}

export async function loader({ request, params }: LoaderFunctionArgs) {
  return handleRequest(request, params);
}

export async function action({ request, params }: ActionFunctionArgs) {
  return handleRequest(request, params);
}
