"use client";

export const dynamic = "force-dynamic";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Loader2, Lock } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

export default function PublicSharePage({ params }: { params: Promise<{ token: string }> }) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [password, setPassword] = useState("");
  const [isLocked, setIsLocked] = useState(false);
  const [token, setToken] = useState<string>("");

  useEffect(() => {
    const loadParams = async () => {
      const { token: resolvedToken } = await params;
      setToken(resolvedToken);
    };
    loadParams();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params]);

  useEffect(() => {
    if (token) {
      fetchData();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  const fetchData = async (pwd?: string) => {
    setLoading(true);
    setError("");
    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const headers: any = {};
      if (pwd) headers["x-share-password"] = pwd;

      const res = await fetch(`/api/share/${token}`, { headers });

      if (res.status === 403) {
        setIsLocked(true);
        setLoading(false);
        return;
      }

      if (!res.ok) {
        const json = await res.json();
        setError(json.error || "Failed to load");
        setLoading(false);
        return;
      }

      const json = await res.json();
      setData(json);
      setIsLocked(false);
    } catch (_e) {
      setError("Network error");
    } finally {
      setLoading(false);
    }
  };

  const handleUnlock = (e: React.FormEvent) => {
    e.preventDefault();
    fetchData(password);
  };

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-50">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (isLocked) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-50">
        <Card className="w-full max-w-sm">
          <CardHeader className="text-center">
            <div className="mx-auto bg-gray-100 p-3 rounded-full mb-2 w-fit">
              <Lock className="h-6 w-6 text-gray-600" />
            </div>
            <CardTitle>Protected Content</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleUnlock} className="space-y-4">
              <Input
                type="password"
                placeholder="Enter Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
              <Button type="submit" className="w-full">
                Unlock
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-screen items-center justify-center bg-gray-50">
        <div className="text-center text-red-500">
          <h1 className="text-2xl font-bold">Error</h1>
          <p>{error}</p>
        </div>
      </div>
    );
  }

  if (!data) return <p className="text-center mt-10">No data found</p>;

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-6xl mx-auto">
        {data.type === "QUERY" && (
          <Card>
            <CardHeader>
              <CardTitle>Shared Query Result</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="bg-gray-100 p-4 rounded font-mono text-sm overflow-auto">
                {data.data.sql}
              </div>
            </CardContent>
          </Card>
        )}
        {/* Add other types as needed */}
      </div>
    </div>
  );
}
