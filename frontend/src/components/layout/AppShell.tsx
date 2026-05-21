import { useEffect } from "react";
import { Link, Outlet, useNavigate } from "react-router-dom";
import { createCheckoutSession, getAuthStatus, logout } from "../../lib/api";
import { useBrowserStore } from "../../stores/useBrowserStore";
import { Button } from "../ui/Button";

export function AppShell() {
    const navigate = useNavigate();
    const {
        auth,
        isAuthLoading,
        setAuth,
        setLoadingState,
        setError,
        reset,
    } = useBrowserStore();

    useEffect(() => {
        let cancelled = false;

        async function syncAuth() {
            try {
                setLoadingState("isAuthLoading", true);
                const status = await getAuthStatus();
                if (cancelled) {
                    return;
                }

                setAuth(status);
            } catch (err) {
                if (cancelled) {
                    return;
                }

                setError(
                    err instanceof Error
                        ? err.message
                        : "Failed to load authentication state",
                );
            } finally {
                if (!cancelled) {
                    setLoadingState("isAuthLoading", false);
                }
            }
        }

        void syncAuth();

        return () => {
            cancelled = true;
        };
    }, [setAuth, setError, setLoadingState]);

    function handleUpgrade() {
        // Stripe checkout isn't wired up yet — keep the createCheckoutSession
        // import around so we can flip back here when it is.
        void createCheckoutSession;
        window.alert(
            "Pro upgrades aren't available yet — we're working on it. Check back soon!",
        );
    }

    async function handleLogout() {
        try {
            await logout();
        } finally {
            reset();
            navigate("/");
        }
    }

    return (
        <div className="min-h-screen bg-slate-950 text-slate-100">
            <header className="border-b border-slate-800 bg-slate-900/90">
                <div className="mx-auto flex max-w-6xl flex-col gap-4 px-4 py-4 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                        <Link
                            to="/"
                            className="text-xl font-semibold text-white"
                        >
                            WoW Log Analyzer
                        </Link>
                        <p className="text-sm text-slate-400">
                            Compare your fight against a high-performing cohort.
                        </p>
                    </div>
                    <div className="flex flex-wrap items-center gap-3">
                        {auth?.authenticated ? (
                            <div className="flex items-center gap-3">
                                <Link
                                    to="/reports"
                                    className="text-sm text-slate-300 hover:text-white"
                                >
                                    My Reports
                                </Link>
                                {auth.user?.name && (
                                    <span className="text-sm text-slate-300">
                                        {auth.user.name}
                                    </span>
                                )}
                                {auth.tier !== "pro" && (
                                    <Button
                                        type="button"
                                        onClick={handleUpgrade}
                                    >
                                        Upgrade to Pro
                                    </Button>
                                )}
                                <Button
                                    type="button"
                                    variant="secondary"
                                    onClick={handleLogout}
                                >
                                    Log out
                                </Button>
                            </div>
                        ) : (
                            <a href="/api/auth/login">
                                <Button
                                    type="button"
                                    disabled={isAuthLoading}
                                >
                                    {isAuthLoading ? "Checking..." : "Log in"}
                                </Button>
                            </a>
                        )}
                    </div>
                </div>
            </header>

            <main className="mx-auto max-w-6xl px-4 py-8">
                <Outlet />
            </main>
        </div>
    );
}
