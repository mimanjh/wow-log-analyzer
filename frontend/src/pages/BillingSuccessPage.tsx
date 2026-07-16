import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { getBillingStatus } from "../lib/api";
import { useBrowserStore } from "../stores/useBrowserStore";

type BillingCheckState = "checking" | "pro" | "pending" | "error";

export function BillingSuccessPage() {
    const navigate = useNavigate();
    const { auth, setAuth } = useBrowserStore();
    const [state, setState] = useState<BillingCheckState>("checking");

    useEffect(() => {
        let cancelled = false;
        let redirectTimer = 0;

        async function syncTier() {
            try {
                const status = await getBillingStatus();
                if (cancelled) {
                    return;
                }
                if (auth) {
                    setAuth({ ...auth, tier: status.tier });
                }
                setState(status.tier === "pro" ? "pro" : "pending");
                redirectTimer = window.setTimeout(
                    () => navigate("/analyze"),
                    3000,
                );
            } catch {
                if (!cancelled) {
                    setState("error");
                }
            }
        }
        void syncTier();

        return () => {
            cancelled = true;
            window.clearTimeout(redirectTimer);
        };
    }, []); // eslint-disable-line react-hooks/exhaustive-deps

    if (state === "error") {
        return (
            <div className="flex flex-col items-center justify-center py-24 text-center">
                <h1 className="text-2xl font-bold text-white mb-2">
                    We couldn't confirm your subscription
                </h1>
                <p className="text-slate-400 mb-6">
                    Your payment may still have gone through. Refresh in a
                    moment, or head back and check your account status.
                </p>
                <Link className="text-sky-400 hover:text-sky-300" to="/analyze">
                    Back to analyze
                </Link>
            </div>
        );
    }

    return (
        <div className="flex flex-col items-center justify-center py-24 text-center">
            <div className="text-4xl mb-4">{state === "pro" ? "🎉" : "⏳"}</div>
            <h1 className="text-2xl font-bold text-white mb-2">
                {state === "pro"
                    ? "You're now a Pro member!"
                    : state === "pending"
                      ? "Payment received — activating your upgrade"
                      : "Confirming your subscription..."}
            </h1>
            <p className="text-slate-400 mb-6">
                {state === "checking"
                    ? "Checking your subscription status..."
                    : state === "pro"
                      ? "Your subscription is active. Redirecting you now..."
                      : "Your upgrade can take a minute to apply. Redirecting you now..."}
            </p>
        </div>
    );
}
