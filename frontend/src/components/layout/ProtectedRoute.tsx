import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";
import { useBrowserStore } from "../../stores/useBrowserStore";

/**
 * Gates a route behind an authenticated session. While auth state is still
 * being fetched (initial mount), renders nothing so we don't briefly bounce
 * the user to home. Once resolved, unauthenticated users get redirected to /.
 */
export function ProtectedRoute({ children }: { children: ReactNode }) {
    const { auth, isAuthLoading } = useBrowserStore();

    if (isAuthLoading && auth === null) {
        return null;
    }

    if (!auth?.authenticated) {
        return <Navigate to="/" replace />;
    }

    return <>{children}</>;
}
