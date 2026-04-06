import { Link, Outlet } from "react-router-dom";

export function AppShell() {
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
                    <nav className="flex flex-wrap gap-3">
                        <Link
                            to="/"
                            className="rounded-md bg-slate-800 px-4 py-2 text-sm font-medium text-slate-100 hover:bg-slate-700"
                        >
                            Home
                        </Link>
                        <Link
                            to="/analyze"
                            className="rounded-md bg-slate-800 px-4 py-2 text-sm font-medium text-slate-100 hover:bg-slate-700"
                        >
                            Analyze
                        </Link>
                        <Link
                            to="/report"
                            className="rounded-md bg-slate-800 px-4 py-2 text-sm font-medium text-slate-100 hover:bg-slate-700"
                        >
                            Report
                        </Link>
                    </nav>
                </div>
            </header>

            <main className="mx-auto max-w-6xl px-4 py-8">
                <Outlet />
            </main>
        </div>
    );
}
