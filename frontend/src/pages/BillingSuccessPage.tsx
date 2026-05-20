import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getBillingStatus } from "../lib/api";
import { useBrowserStore } from "../stores/useBrowserStore";

export function BillingSuccessPage() {
    const navigate = useNavigate();
    const { auth, setAuth } = useBrowserStore();
    const [checking, setChecking] = useState(true);

    useEffect(() => {
        async function syncTier() {
            try {
                const status = await getBillingStatus();
                if (auth) {
                    setAuth({ ...auth, tier: status.tier });
                }
            } finally {
                setChecking(false);
                setTimeout(() => navigate("/analyze"), 3000);
            }
        }
        void syncTier();
    }, []); // eslint-disable-line react-hooks/exhaustive-deps

    return (
        <div className="flex flex-col items-center justify-center py-24 text-center">
            <div className="text-4xl mb-4">🎉</div>
            <h1 className="text-2xl font-bold text-white mb-2">You're now a Pro member!</h1>
            <p className="text-slate-400 mb-6">
                {checking ? "Activating your subscription..." : "Your subscription is active. Redirecting you now..."}
            </p>
        </div>
    );
}
