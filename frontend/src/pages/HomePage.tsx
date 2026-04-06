import { useNavigate } from "react-router-dom";
import { UrlIntakeForm } from "../features/url-intake/UrlIntakeForm";
import { useAnalyzeStore } from "../stores/useAnalyzeStore";
import { usePageTitle } from "../hooks/usePageTitle";

export function HomePage() {
    usePageTitle("Home");
    const setReportUrl = useAnalyzeStore((state) => state.setReportUrl);
    const navigate = useNavigate();

    function handleSubmit(url: string) {
        setReportUrl(url);
        navigate("/analyze");
    }

    return (
        <section className="space-y-8">
            <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8 shadow-xl shadow-slate-950/20">
                <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                    Get started
                </p>
                <h1 className="mt-3 text-4xl font-semibold text-white sm:text-5xl">
                    Paste your Warcraft Logs report URL.
                </h1>
                <p className="mt-4 max-w-2xl text-slate-300">
                    The app will parse the report code, then you can choose a
                    fight and character for analysis.
                </p>
            </div>

            <div className="grid gap-6 lg:grid-cols-[1fr_320px]">
                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <UrlIntakeForm onSubmit={handleSubmit} />
                </div>
                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <h2 className="text-2xl font-semibold text-white">
                        Next steps
                    </h2>
                    <ul className="mt-4 space-y-3 text-slate-300">
                        <li>1. Paste a report URL</li>
                        <li>2. Select a fight and character</li>
                        <li>3. Review metrics and insights</li>
                    </ul>
                </div>
            </div>
        </section>
    );
}
