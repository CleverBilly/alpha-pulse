import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { verifySessionToken } from "@/lib/auth/session";

const AUTH_ENABLED = process.env.NEXT_PUBLIC_AUTH_ENABLED === "true";
const AUTH_COOKIE_NAME = process.env.AUTH_COOKIE_NAME ?? "alpha_pulse_session";
const AUTH_SESSION_SECRET = process.env.AUTH_SESSION_SECRET ?? "";
const PROTECTED_PREFIXES = ["/dashboard", "/chart", "/signals", "/market"];

export async function middleware(request: NextRequest) {
  if (!AUTH_ENABLED) {
    return NextResponse.next();
  }

  const pathname = request.nextUrl.pathname;
  const isLoginRoute = pathname === "/login";
  const isProtectedRoute = pathname === "/" || PROTECTED_PREFIXES.some((prefix) => pathname.startsWith(prefix));

  if (!isLoginRoute && !isProtectedRoute) {
    return NextResponse.next();
  }

  const session = await verifySessionToken({
    token: request.cookies.get(AUTH_COOKIE_NAME)?.value,
    secret: AUTH_SESSION_SECRET,
  });

  if (isLoginRoute && session.valid) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  if (isProtectedRoute && !session.valid) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/", "/login", "/dashboard/:path*", "/chart/:path*", "/signals/:path*", "/market/:path*"],
};
