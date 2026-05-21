import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
    createReportJob,
    listSavedReports,
    type SavedReport,
} from "../lib/api";
import { Button } from "../components/ui/Button";
import { usePageTitle } from "../hooks/usePageTitle";
import { useAnalyzeStore } from "../stores/useAnalyzeStore";
import type { Character, Fight } from "../types";

function formatAnalyzedAt(value: string): string {
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) {
        return value;
    }
    return parsed.toLocaleString(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
    });
}

// RFC3339 zero time — Go's time.Time zero value when (un)marshaled as JSON.
// Empty strings would fail the time.Time decoder; this satisfies it.
const ZERO_TIME = "0001-01-01T00:00:00Z";

/**
 * Build a minimal Fight/Character pair from a SavedReport. Adequate for the
 * cache-hit reload flow (backend only uses IDs to look up the cached result;
 * the cached result itself carries the complete fight/character objects).
 */
function reconstructRequest(report: SavedReport): {
    fight: Fight;
    character: Character;
} {
    return {
        fight: {
            id: report.fightId,
            name: report.encounterName,
            difficulty: report.difficulty,
            kill: true,
            killTime: 0,
            encounterId: 0,
            startTime: ZERO_TIME,
            endTime: ZERO_TIME,
            friendlyPlayers: [],
        },
        character: {
            id: report.characterId,
            name: report.characterName,
            class: report.characterClass,
            spec: report.characterSpec,
            role: "",
            serverName: "",
        },
    };
}

export function ReportsPage() {
    usePageTitle("Saved Reports");
    const navigate = useNavigate();
    const { setReportJob, setReportResult } = useAnalyzeStore();
    const [reports, setReports] = useState<SavedReport[] | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [confirmReport, setConfirmReport] = useState<SavedReport | null>(
        null,
    );
    const [loadingReportId, setLoadingReportId] = useState<string | null>(null);

    useEffect(() => {
        let cancelled = false;

        async function load() {
            try {
                setLoading(true);
                setError(null);
                const list = await listSavedReports();
                if (!cancelled) {
                    setReports(list);
                }
            } catch (err) {
                if (!cancelled) {
                    setError(
                        err instanceof Error
                            ? err.message
                            : "Failed to load reports",
                    );
                }
            } finally {
                if (!cancelled) {
                    setLoading(false);
                }
            }
        }

        void load();

        return () => {
            cancelled = true;
        };
    }, []);

    const groupedKey = useMemo(
        () => (report: SavedReport) =>
            `${report.reportId}:${report.fightId}:${report.characterId}`,
        [],
    );

    async function openReport(report: SavedReport) {
        const key = groupedKey(report);
        try {
            setLoadingReportId(key);
            setError(null);
            const { fight, character } = reconstructRequest(report);
            const job = await createReportJob(
                report.reportId,
                fight,
                character,
            );
            setReportJob(job);
            if (job.result) {
                setReportResult(job.result);
            }
            navigate("/report");
        } catch (err) {
            setError(
                err instanceof Error ? err.message : "Failed to load report",
            );
        } finally {
            setLoadingReportId(null);
        }
    }

    function handleCardClick(report: SavedReport) {
        if (report.cached) {
            void openReport(report);
            return;
        }
        setConfirmReport(report);
    }

    function confirmReanalyze() {
        if (!confirmReport) return;
        const target = confirmReport;
        setConfirmReport(null);
        void openReport(target);
    }

    return (
        <section className="space-y-8">
            <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                    Your Reports
                </p>
                <h1 className="mt-3 text-3xl font-semibold text-white">
                    Saved Analyses
                </h1>
                <p className="mt-4 max-w-2xl text-slate-300">
                    Every analysis you've run is listed here. Cached reports
                    open instantly. Expired reports require a fresh
                    re-analysis, which uses one of your daily analyses.
                </p>
            </div>

            {error && (
                <div className="rounded-3xl border border-rose-800 bg-rose-950/20 p-6">
                    <p className="text-sm text-rose-300">{error}</p>
                </div>
            )}

            {loading && (
                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <p className="text-slate-300">Loading your reports…</p>
                </div>
            )}

            {!loading && reports && reports.length === 0 && (
                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <h2 className="text-xl font-semibold text-white">
                        No saved reports yet
                    </h2>
                    <p className="mt-3 text-slate-300">
                        Run an analysis from the home page and it will show up
                        here for quick re-access.
                    </p>
                    <Button
                        className="mt-6"
                        onClick={() => navigate("/")}
                    >
                        Analyze a report
                    </Button>
                </div>
            )}

            {!loading && reports && reports.length > 0 && (
                <div className="grid gap-6 sm:grid-cols-2 xl:grid-cols-3">
                    {reports.map((report) => {
                        const key = groupedKey(report);
                        const isLoading = loadingReportId === key;
                        return (
                            <button
                                key={key}
                                type="button"
                                onClick={() => handleCardClick(report)}
                                disabled={isLoading}
                                className={`rounded-3xl border-2 p-5 text-left transition ${
                                    report.cached
                                        ? "border-emerald-500/40 bg-emerald-950/10 hover:border-emerald-400/70"
                                        : "border-slate-800 bg-slate-900/70 hover:border-slate-700"
                                } ${isLoading ? "opacity-60" : ""}`}
                            >
                                <div className="flex items-start justify-between gap-2">
                                    <div>
                                        <p className="text-xs uppercase tracking-[0.2em] text-sky-400">
                                            {report.difficulty || "Encounter"}
                                        </p>
                                        <h3 className="mt-1 text-lg font-semibold text-white">
                                            {report.encounterName ||
                                                "Unknown encounter"}
                                        </h3>
                                    </div>
                                    {report.cached ? (
                                        <span className="rounded-full bg-emerald-950/60 px-2 py-0.5 text-xs text-emerald-300 ring-1 ring-emerald-700/40">
                                            Cached
                                        </span>
                                    ) : (
                                        <span className="rounded-full bg-amber-950/60 px-2 py-0.5 text-xs text-amber-300 ring-1 ring-amber-700/40">
                                            Expired
                                        </span>
                                    )}
                                </div>
                                <div className="mt-4 space-y-1 text-sm text-slate-300">
                                    <p className="font-medium text-white">
                                        {report.characterName}
                                    </p>
                                    <p>
                                        {report.characterSpec}{" "}
                                        {report.characterClass}
                                    </p>
                                </div>
                                <p className="mt-4 text-xs text-slate-400">
                                    Analyzed{" "}
                                    {formatAnalyzedAt(report.analyzedAt)}
                                </p>
                                {isLoading && (
                                    <p className="mt-3 text-xs text-sky-300">
                                        Loading…
                                    </p>
                                )}
                            </button>
                        );
                    })}
                </div>
            )}

            {confirmReport && (
                <div
                    className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 p-4"
                    role="dialog"
                    aria-modal="true"
                >
                    <div className="w-full max-w-lg rounded-3xl border border-slate-800 bg-slate-900 p-8">
                        <h2 className="text-xl font-semibold text-white">
                            Re-analyze expired report?
                        </h2>
                        <p className="mt-4 text-slate-300">
                            This report's cached result is no longer available.
                            Re-running the analysis will use one of your daily
                            analyses and may take a minute.
                        </p>
                        <div className="mt-2 rounded-2xl border border-slate-800 bg-slate-950/60 p-4 text-sm text-slate-200">
                            <p className="font-medium text-white">
                                {confirmReport.encounterName} (
                                {confirmReport.difficulty})
                            </p>
                            <p className="text-slate-400">
                                {confirmReport.characterName} —{" "}
                                {confirmReport.characterSpec}{" "}
                                {confirmReport.characterClass}
                            </p>
                        </div>
                        <div className="mt-6 flex flex-col gap-3 sm:flex-row sm:justify-end">
                            <Button
                                variant="secondary"
                                onClick={() => setConfirmReport(null)}
                            >
                                Cancel
                            </Button>
                            <Button onClick={confirmReanalyze}>
                                Re-analyze
                            </Button>
                        </div>
                    </div>
                </div>
            )}
        </section>
    );
}
