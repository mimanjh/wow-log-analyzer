import { Component, type ErrorInfo, type ReactNode } from "react";

interface ErrorBoundaryProps {
    children: ReactNode;
}

interface ErrorBoundaryState {
    error: Error | null;
}

export class ErrorBoundary extends Component<
    ErrorBoundaryProps,
    ErrorBoundaryState
> {
    state: ErrorBoundaryState = { error: null };

    static getDerivedStateFromError(error: Error): ErrorBoundaryState {
        return { error };
    }

    componentDidCatch(error: Error, info: ErrorInfo) {
        console.error("Unhandled render error:", error, info.componentStack);
    }

    render() {
        if (!this.state.error) {
            return this.props.children;
        }

        return (
            <div className="flex min-h-screen items-center justify-center bg-slate-950 px-6">
                <div className="max-w-lg rounded-3xl border border-slate-800 bg-slate-900/80 p-8 text-center">
                    <p className="text-sm uppercase tracking-[0.25em] text-rose-400">
                        Something went wrong
                    </p>
                    <h1 className="mt-3 text-2xl font-semibold text-white">
                        The page hit an unexpected error
                    </h1>
                    <p className="mt-4 text-sm text-slate-300">
                        {this.state.error.message}
                    </p>
                    <button
                        type="button"
                        className="mt-8 rounded-full bg-sky-500 px-6 py-2 text-sm font-semibold text-white transition hover:bg-sky-400"
                        onClick={() => {
                            window.location.href = "/";
                        }}
                    >
                        Back to home
                    </button>
                </div>
            </div>
        );
    }
}
