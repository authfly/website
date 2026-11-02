// Backend for frontend (Bun): OpenID Connect code flow + PKCE against Authfly with openid-client,
// the session in a signed HttpOnly cookie, and the built SPA. The browser never sees a token.
import * as oidc from "openid-client";
import { join } from "node:path";

const env = (k: string, d?: string) => {
  const v = process.env[k] ?? d;
  if (v === undefined) throw new Error(`${k} is not set`);
  return v;
};
const ISSUER = env("OIDC_ISSUER", "http://localhost:4000"); // Authfly public_url
const CLIENT_ID = env("OIDC_CLIENT_ID", "my-app"); // app.clients[].id in authfly.yaml
const APP_URL = env("APP_URL", "http://localhost:5173");
const REDIRECT_URI = `${APP_URL}/callback`; // app.clients[].redirect_uris in authfly.yaml
const SECRET = env("SESSION_SECRET"); // 32+ random bytes: openssl rand -hex 32
const secure = APP_URL.startsWith("https://");

// Public client (no secret) with PKCE. http:// issuers are allowed only for local development.
const config = await oidc.discovery(new URL(ISSUER), CLIENT_ID, undefined, oidc.None(),
  ISSUER.startsWith("http://") ? { execute: [oidc.allowInsecureRequests] } : undefined);

// Signed cookies (HMAC-SHA256): value.signature
const enc = new TextEncoder();
const key = await crypto.subtle.importKey("raw", enc.encode(SECRET), { name: "HMAC", hash: "SHA-256" }, false, ["sign", "verify"]);
const b64 = (b: ArrayBuffer | Uint8Array) => Buffer.from(b instanceof Uint8Array ? b : new Uint8Array(b)).toString("base64url");
async function seal(v: unknown, ttl: number) {
  const p = b64(enc.encode(JSON.stringify({ v, exp: Math.floor(Date.now() / 1000) + ttl })));
  return `${p}.${b64(await crypto.subtle.sign("HMAC", key, enc.encode(p)))}`;
}
async function unseal<T>(t?: string): Promise<T | null> {
  const [p, s] = (t ?? "").split(".");
  if (!p || !s || !(await crypto.subtle.verify("HMAC", key, Buffer.from(s, "base64url"), enc.encode(p)))) return null;
  const { v, exp } = JSON.parse(Buffer.from(p, "base64url").toString());
  return exp > Date.now() / 1000 ? (v as T) : null;
}
const getCookie = (req: Request, n: string) => req.headers.get("cookie")?.split(/;\s*/).find((c) => c.startsWith(n + "="))?.slice(n.length + 1);
const setCookie = (n: string, v: string, maxAge: number) => `${n}=${v}; Path=/; HttpOnly; SameSite=Lax; Max-Age=${maxAge}${secure ? "; Secure" : ""}`;
const redirect = (to: string, cookies: string[] = []) => {
  const h = new Headers({ Location: to });
  for (const c of cookies) h.append("Set-Cookie", c);
  return new Response(null, { status: 303, headers: h });
};
type Pending = { verifier: string; state: string; nonce: string };
type User = { sub: string; email: string; name: string; username: string };
const dist = join(import.meta.dir, "dist");

Bun.serve({
  port: Number(env("PORT", "3000")),
  async fetch(req) {
    const url = new URL(req.url);
    if (url.pathname === "/login") {
      const p: Pending = { verifier: oidc.randomPKCECodeVerifier(), state: oidc.randomState(), nonce: oidc.randomNonce() };
      const to = oidc.buildAuthorizationUrl(config, {
        redirect_uri: REDIRECT_URI, scope: "openid email profile", state: p.state, nonce: p.nonce,
        code_challenge: await oidc.calculatePKCECodeChallenge(p.verifier), code_challenge_method: "S256",
      });
      return redirect(to.href, [setCookie("app_pending", await seal(p, 900), 900)]);
    }
    if (url.pathname === "/callback") {
      const p = await unseal<Pending>(getCookie(req, "app_pending"));
      try {
        if (!p) throw new Error("no pending sign-in");
        const tokens = await oidc.authorizationCodeGrant(config, new URL(req.url, APP_URL), {
          pkceCodeVerifier: p.verifier, expectedState: p.state, expectedNonce: p.nonce,
        });
        const c = tokens.claims()!;
        const email = String(c.email ?? "");
        const user: User = { sub: c.sub, email, name: String(c.name ?? ""), username: String(c.preferred_username ?? email.split("@")[0]) };
        return redirect("/", [setCookie("app_session", await seal(user, 12 * 3600), 12 * 3600), setCookie("app_pending", "", 0)]);
      } catch (e) {
        console.error("callback:", (e as Error).message);
        return redirect("/?error=sign_in_failed", [setCookie("app_pending", "", 0)]);
      }
    }
    if (url.pathname === "/logout" && req.method === "POST") return redirect("/", [setCookie("app_session", "", 0)]);
    if (url.pathname === "/api/me") {
      const user = await unseal<User>(getCookie(req, "app_session"));
      return user ? Response.json({ user }) : Response.json({ error: "unauthenticated" }, { status: 401 });
    }
    const file = Bun.file(join(dist, url.pathname === "/" ? "index.html" : url.pathname));
    return (await file.exists()) ? new Response(file) : new Response(Bun.file(join(dist, "index.html")));
  },
});
console.log(`app on :${env("PORT", "3000")} (issuer ${ISSUER})`);
