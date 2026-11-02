// Signed out: a notice with "Sign in" (OIDC code flow + PKCE through the Authfly hosted pages).
// Signed in: the account widget. Primitives come from `npx ui8kit@3 init`, widgets from the Authfly registry.
import { useEffect, useState } from "react";
import { Box } from "./ui/box/box";
import { AuthflyAccount } from "./blocks/authfly-account";
import { AuthflyNotice } from "./blocks/authfly-notice";

type User = { sub: string; email: string; name: string; username: string; amr: string };
const t = {
  title: "Welcome",
  text: "This app keeps no passwords. Sign in with your Authfly account.",
  cta: "Sign in",
  account: "Your account",
  accountText: "Signed in through Authfly (OpenID Connect).",
  labels: { name: "Name", email: "Email", verified: "Verified", unverified: "Not verified", security: "Sign-in methods", signout: "Sign out" },
};

export function App() {
  const [user, setUser] = useState<User | null | undefined>(undefined);
  useEffect(() => {
    fetch("/api/me").then(async (r) => setUser(r.ok ? ((await r.json()) as { user: User }).user : null));
  }, []);
  return (
    <Box tag="main" className="w-full px-4 py-12">
      {user === undefined ? null : user ? (
        <AuthflyAccount id="me" title={t.account} text={t.accountText} name={user.name || user.username} email={user.email}
          verified methods={[]} labels={t.labels} signoutUrl="/logout" />
      ) : (
        <AuthflyNotice id="welcome" title={t.title} text={t.text} icon="log-in" tone="default"
          labels={{ cta: t.cta, ctaHref: "/login", secondary: "", secondaryHref: "" }} />
      )}
    </Box>
  );
}
