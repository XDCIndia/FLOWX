"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api";
import { useToast } from "@/lib/toast-context";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000";

export default function LoginPage() {
  const [tab, setTab] = useState<"signin" | "signup">("signin");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const { login } = useAuth();
  const router = useRouter();
  const { toast } = useToast();
  const [apiKey, setApiKey] = useState("");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [createdKey, setCreatedKey] = useState("");  const [copied, setCopied] = useState(false);
  const handleSignIn = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);
    try {
      login(apiKey);
      await api.getHealth();
      toast("Connected to FlowX API", "success");
      router.push("/overview");
    } catch {
      setError("Invalid API key or API unreachable");
      localStorage.removeItem("flowx_api_key");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSignUp = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);
    try {
      const regRes = await fetch(API_URL + "/v1/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, email, password }),
      });
      if (!regRes.ok) {
        const err = await regRes.json().catch(() => ({}));
        throw new Error(err.error || "Registration failed");
      }

      const loginRes = await fetch(API_URL + "/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      if (!loginRes.ok) throw new Error("Login after registration failed");
      const loginData = await loginRes.json();
      const token = loginData.access_token;

      const keyRes = await fetch(API_URL + "/v1/keys", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: "Bearer " + token,
        },
        body: JSON.stringify({ label: "default" }),
      });
      if (!keyRes.ok) throw new Error("Failed to create API key");
      const keyData = await keyRes.json();

      setCreatedKey(keyData.key);    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setIsLoading(false);
    }
  };

  const handleCopyKey = () => {    navigator.clipboard.writeText(createdKey);    setCopied(true);    setTimeout(() => setCopied(false), 2000);  };  const handleContinueToDashboard = () => {    login(createdKey);    toast("Account created! Welcome to FlowX.", "success");    router.push("/overview");  };  if (createdKey) {    return (      <div className="flex min-h-screen items-center justify-center p-6">        <Card className="w-full max-w-[480px] animate-in fade-in slide-in-from-bottom-4 duration-500">          <CardHeader className="text-center">            <div className="mx-auto flex h-10 w-10 items-center justify-center rounded-lg bg-green-500 text-white">              <span className="text-lg">✓</span>            </div>            <CardTitle className="text-xl">Account Created!</CardTitle>            <CardDescription>              Save this API key — it won’t be shown again.            </CardDescription>          </CardHeader>          <CardContent className="space-y-4">            <div className="rounded-lg border border-border bg-muted p-4">              <label className="text-xs font-medium text-muted-foreground uppercase tracking-wide">                Your API Key              </label>              <div className="mt-2 flex items-center gap-2">                <code className="flex-1 break-all text-sm font-mono text-foreground">                  {createdKey}                </code>                <Button variant="secondary" size="sm" onClick={handleCopyKey} className="shrink-0">                  {copied ? "Copied!" : "Copy"}                </Button>              </div>            </div>            <div className="rounded-md bg-amber-500/10 px-4 py-3 text-sm text-amber-600 dark:text-amber-400">              This key is shown only once. Copy it now before continuing.            </div>            <Button className="w-full" size="lg" onClick={handleContinueToDashboard}>              Continue to Dashboard            </Button>          </CardContent>        </Card>      </div>    );  }  return (
    <div className="flex min-h-screen items-center justify-center p-6">
      <Card className="w-full max-w-[420px] animate-in fade-in slide-in-from-bottom-4 duration-500">
        <CardHeader className="text-center">
          <div className="mx-auto flex h-10 w-10 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <span className="font-bold">F</span>
          </div>
          <CardTitle className="text-xl">
            {tab === "signin" ? "Sign in to FlowX" : "Create your account"}
          </CardTitle>
          <CardDescription>
            {tab === "signin"
              ? "Enter your secret API key to access the dashboard."
              : "Get started with FlowX in seconds."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex rounded-lg border border-border bg-muted p-1 mb-5">
            <button
              type="button"
              onClick={() => { setTab("signin"); setError(""); }}
              className={tab === "signin"
                ? "flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors bg-background text-foreground shadow-sm"
                : "flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors text-muted-foreground"}
            >
              Sign In
            </button>
            <button
              type="button"
              onClick={() => { setTab("signup"); setError(""); }}
              className={tab === "signup"
                ? "flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors bg-background text-foreground shadow-sm"
                : "flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors text-muted-foreground"}
            >
              Create Account
            </button>
          </div>

          {error && (
            <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {error}
            </div>
          )}

          {tab === "signin" ? (
            <form onSubmit={handleSignIn} className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">API Key</label>
                <Input
                  type="password"
                  placeholder="sk_live_..."
                  value={apiKey}
                  onChange={(e) => setApiKey(e.target.value)}
                  required
                />
              </div>
              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? "Connecting..." : "Sign In"}
              </Button>
            </form>
          ) : (
            <form onSubmit={handleSignUp} className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Name</label>
                <Input
                  type="text"
                  placeholder="Your name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Email</label>
                <Input
                  type="email"
                  placeholder="you@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Password</label>
                <Input
                  type="password"
                  placeholder="Min 6 characters"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  minLength={6}
                  required
                />
              </div>
              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading ? "Creating account..." : "Create Account & Get API Key"}
              </Button>
            </form>
          )}

          <p className="mt-5 text-center text-xs text-muted-foreground">
            <a href="/" className="underline hover:text-foreground">
              Back to FlowX
            </a>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
