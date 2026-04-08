import { useEffect } from "react";
import { Link } from "react-router-dom";
import { useAnalyzeStore } from "../stores/useAnalyzeStore";
import { usePageTitle } from "../hooks/usePageTitle";
import { Button } from "../components/ui/Button";
import { MetricCard } from "../components/MetricCard";
import { getReportJob } from "../lib/api";
import type { Character, Fight, ReportJob } from "../types";

const reportStages = [
    { key: "player-data", label: "Fetch Player Data" },
    { key: "rankings", label: "Find Ranking Cohort" },
    { key: "cohort", label: "Load Top Ranked Fights" },
    { key: "analyzing", label: "Run Deterministic Analysis" },
    { key: "insights", label: "Generate Insights" },
    { key: "completed", label: "Complete" },
] as const;

function formatKillTime(seconds: number) {
    return `${Math.floor(seconds / 60)}:${(seconds % 60)
        .toString()
        .padStart(2, "0")}`;
}

function getStageLabel(stage: string) {
    return reportStages.find((entry) => entry.key === stage)?.label ?? stage;
}

function getStageCompletionState(stage: string, status: ReportJob["status"], key: string) {
    const stageIndex = reportStages.findIndex((entry) => entry.key === stage);
    const keyIndex = reportStages.findIndex((entry) => entry.key === key);

    if (status === "completed") {
        return "complete";
    }
    if (status === "failed") {
        if (keyIndex < stageIndex) {
            return "complete";
        }
        if (key === stage) {
            return "failed";
        }
        return "pending";
    }
    if (keyIndex < stageIndex) {
        return "complete";
    }
    if (key === stage) {
        return "active";
    }
    return "pending";
}

function renderSummaryCard(title: string, content: Array<{ label: string; value: string }>) {
    return (
        <div className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
            <h2 className="text-lg font-semibold text-white">{title}</h2>
            <div className="mt-4 space-y-3">
                {content.map((item) => (
                    <div key={item.label}>
                        <p className="text-sm text-slate-400">{item.label}</p>
                        <p className="text-white">{item.value}</p>
                    </div>
                ))}
            </div>
        </div>
    );
}

function renderProgressView(reportJob: ReportJob, fight: Fight, character: Character) {
    const progressPercent =
        reportJob.progress.total > 0
            ? Math.min(
                  100,
                  Math.round(
                      (reportJob.progress.current / reportJob.progress.total) * 100,
                  ),
              )
            : 0;
    const isCohortStage = reportJob.stage === "cohort" && reportJob.progress.total > 0;

    return (
        <>
            <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                    Analysis Progress
                </p>
                <h1 className="mt-3 text-3xl font-semibold text-white">
                    {reportJob.status === "failed"
                        ? "Analysis failed"
                        : reportJob.status === "completed"
                          ? "Analysis complete"
                          : "Analyzing fight"}
                </h1>
                <p className="mt-4 max-w-2xl text-slate-300">
                    {reportJob.message}
                </p>

                <div className="mt-6 rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                        <div>
                            <p className="text-sm uppercase tracking-[0.2em] text-sky-400">
                                {getStageLabel(reportJob.stage)}
                            </p>
                            <p className="mt-2 text-sm text-slate-300">
                                Status: {reportJob.status}
                            </p>
                        </div>
                        <div className="text-sm text-slate-400">
                            {isCohortStage
                                ? `Cohort member ${reportJob.progress.current} of ${reportJob.progress.total}`
                                : `Step ${Math.min(
                                      reportJob.progress.current,
                                      reportJob.progress.total || 1,
                                  )} of ${reportJob.progress.total || 1}`}
                        </div>
                    </div>

                    <div className="mt-4 h-3 overflow-hidden rounded-full bg-slate-800">
                        <div
                            className={`h-full rounded-full transition-all ${
                                reportJob.status === "failed"
                                    ? "bg-rose-500"
                                    : reportJob.status === "completed"
                                      ? "bg-emerald-500"
                                      : "bg-sky-500"
                            }`}
                            style={{ width: `${progressPercent}%` }}
                        />
                    </div>

                    {reportJob.error && (
                        <p className="mt-4 text-sm text-rose-400">
                            {reportJob.error}
                        </p>
                    )}
                </div>
            </div>

            <div className="grid gap-6 lg:grid-cols-2">
                {renderSummaryCard("Fight Summary", [
                    { label: "Encounter", value: fight.name },
                    { label: "Difficulty", value: fight.difficulty },
                    { label: "Result", value: fight.kill ? "Kill" : "Wipe" },
                    { label: "Kill Time", value: formatKillTime(fight.killTime) },
                ])}

                {renderSummaryCard("Character Summary", [
                    { label: "Name", value: character.name },
                    {
                        label: "Class & Spec",
                        value: `${character.class} ${character.spec}`,
                    },
                    { label: "Role", value: character.role },
                    {
                        label: "Server",
                        value: character.serverName || "Unknown server",
                    },
                ])}
            </div>

            <div className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                <h2 className="text-lg font-semibold text-white">Pipeline</h2>
                <div className="mt-4 grid gap-3 lg:grid-cols-3">
                    {reportStages.map((step) => {
                        const state = getStageCompletionState(
                            reportJob.stage,
                            reportJob.status,
                            step.key,
                        );
                        const stateClasses =
                            state === "complete"
                                ? "border-emerald-700/60 bg-emerald-950/20 text-emerald-200"
                                : state === "active"
                                  ? "border-sky-500/60 bg-sky-950/20 text-sky-100"
                                  : state === "failed"
                                    ? "border-rose-700/60 bg-rose-950/20 text-rose-200"
                                    : "border-slate-800 bg-slate-900/80 text-slate-400";

                        return (
                            <div
                                key={step.key}
                                className={`rounded-3xl border p-4 ${stateClasses}`}
                            >
                                <p className="text-xs uppercase tracking-[0.2em]">
                                    {state}
                                </p>
                                <p className="mt-2 font-medium">{step.label}</p>
                            </div>
                        );
                    })}
                </div>
            </div>

            <div className="flex flex-col gap-3 sm:flex-row">
                <Link to="/select">
                    <Button variant="secondary">Back to selection</Button>
                </Link>
                <Link to="/analyze">
                    <Button variant="secondary">Back to logs</Button>
                </Link>
            </div>
        </>
    );
}

export function ReportPage() {
    usePageTitle("Report");
    const { reportJob, reportResult, setReportJob, setReportResult, setError } =
        useAnalyzeStore();
    const reportJobId = reportJob?.jobId;
    const reportJobStatus = reportJob?.status;

    useEffect(() => {
        if (
            !reportJobId ||
            reportJobStatus === "completed" ||
            reportJobStatus === "failed"
        ) {
            return;
        }

        let cancelled = false;

        const refreshJob = async () => {
            try {
                const nextJob = await getReportJob(reportJobId);
                if (cancelled) {
                    return;
                }

                setReportJob(nextJob);
                if (nextJob.result) {
                    setReportResult(nextJob.result);
                }
                if (nextJob.status === "failed" && nextJob.error) {
                    setError(nextJob.error);
                }
            } catch (err) {
                if (cancelled) {
                    return;
                }

                setError(
                    err instanceof Error
                        ? err.message
                        : "Failed to refresh report status",
                );
            }
        };

        void refreshJob();
        const intervalId = window.setInterval(() => {
            void refreshJob();
        }, 5000);

        return () => {
            cancelled = true;
            window.clearInterval(intervalId);
        };
    }, [reportJobId, reportJobStatus, setError, setReportJob, setReportResult]);

    if (!reportJob && !reportResult) {
        return (
            <section className="space-y-8">
                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                        Report
                    </p>
                    <h1 className="mt-3 text-3xl font-semibold text-white">
                        No analysis data available
                    </h1>
                    <p className="mt-4 max-w-2xl text-slate-300">
                        Please go back to the analyze page and select a fight and character.
                    </p>
                    <div className="mt-8 flex flex-col gap-3 sm:flex-row">
                        <Link to="/analyze">
                            <Button variant="secondary">Back to analyze</Button>
                        </Link>
                        <Link to="/">
                            <Button variant="secondary">Back to home</Button>
                        </Link>
                    </div>
                </div>
            </section>
        );
    }

    if (!reportResult && reportJob) {
        return <section className="space-y-8">{renderProgressView(reportJob, reportJob.fight, reportJob.character)}</section>;
    }

    const { fight, character, comparison, ai } = reportResult!;
    const { cohortStats, deltas } = comparison;

    return (
        <section className="space-y-8">
            {reportJob &&
                reportJob.status !== "completed" &&
                renderProgressView(reportJob, reportJob.fight, reportJob.character)}

            <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                    Report
                </p>
                <h1 className="mt-3 text-3xl font-semibold text-white">
                    Analysis Results
                </h1>
                <p className="mt-4 max-w-2xl text-slate-300">
                    Comparison against {cohortStats.sampleSize} high-performing
                    players.
                </p>
            </div>

            <div className="grid gap-6 lg:grid-cols-2">
                {renderSummaryCard("Fight Summary", [
                    { label: "Encounter", value: fight.name },
                    { label: "Difficulty", value: fight.difficulty },
                    { label: "Result", value: fight.kill ? "Kill" : "Wipe" },
                    { label: "Kill Time", value: formatKillTime(fight.killTime) },
                ])}

                {renderSummaryCard("Character Summary", [
                    { label: "Name", value: character.name },
                    {
                        label: "Class & Spec",
                        value: `${character.class} ${character.spec}`,
                    },
                    { label: "Role", value: character.role },
                    {
                        label: "Server",
                        value: character.serverName || "Unknown server",
                    },
                ])}
            </div>

            {ai.warning && (
                <div className="rounded-3xl border border-amber-700/60 bg-amber-950/20 p-6">
                    <p className="text-sm text-amber-300">{ai.warning}</p>
                </div>
            )}

            <div className="space-y-6">
                <h2 className="text-xl font-semibold text-white">
                    Performance Metrics
                </h2>

                <div className="grid gap-6 lg:grid-cols-2">
                    <MetricCard
                        title="Casts per Minute"
                        description="Average spell casts per minute during the fight"
                        playerValue={deltas.castsPerMin.playerValue}
                        cohortValue={deltas.castsPerMin.cohortValue}
                        delta={deltas.castsPerMin.difference}
                        percentile={deltas.castsPerMin.percentile}
                        confidence={
                            deltas.castsPerMin.confidence as
                                | "high"
                                | "medium"
                                | "low"
                        }
                        caution={deltas.castsPerMin.caution}
                    />

                    <MetricCard
                        title="Major Cooldown Count"
                        description="Number of major defensive/healing cooldowns used"
                        playerValue={deltas.majorCdCount.playerValue}
                        cohortValue={deltas.majorCdCount.cohortValue}
                        delta={deltas.majorCdCount.difference}
                        percentile={deltas.majorCdCount.percentile}
                        confidence={
                            deltas.majorCdCount.confidence as
                                | "high"
                                | "medium"
                                | "low"
                        }
                        caution={deltas.majorCdCount.caution}
                    />

                    <MetricCard
                        title="Major Cooldown Timing Drift"
                        description="Average deviation from optimal cooldown timing (seconds)"
                        playerValue={deltas.majorCdDrift.playerValue}
                        cohortValue={deltas.majorCdDrift.cohortValue}
                        delta={deltas.majorCdDrift.difference}
                        percentile={deltas.majorCdDrift.percentile}
                        confidence={
                            deltas.majorCdDrift.confidence as
                                | "high"
                                | "medium"
                                | "low"
                        }
                        caution={deltas.majorCdDrift.caution}
                        unit="s"
                    />

                    <MetricCard
                        title="Buff Uptime"
                        description="Percentage of fight time with key buffs active"
                        playerValue={deltas.buffUptime.playerValue}
                        cohortValue={deltas.buffUptime.cohortValue}
                        delta={deltas.buffUptime.difference}
                        percentile={deltas.buffUptime.percentile}
                        confidence={
                            deltas.buffUptime.confidence as
                                | "high"
                                | "medium"
                                | "low"
                        }
                        caution={deltas.buffUptime.caution}
                        unit="%"
                    />

                    <MetricCard
                        title="Downtime Percentage"
                        description="Percentage of fight time spent not casting"
                        playerValue={deltas.downtimePct.playerValue}
                        cohortValue={deltas.downtimePct.cohortValue}
                        delta={deltas.downtimePct.difference}
                        percentile={deltas.downtimePct.percentile}
                        confidence={
                            deltas.downtimePct.confidence as
                                | "high"
                                | "medium"
                                | "low"
                        }
                        caution={deltas.downtimePct.caution}
                        unit="%"
                    />
                </div>
            </div>

            <div className="space-y-6">
                <h2 className="text-xl font-semibold text-white">Insights</h2>

                {ai.insights.length > 0 ? (
                    <div className="grid gap-6 lg:grid-cols-3">
                        {ai.insights.map((insight) => (
                            <article
                                key={insight.metricKey}
                                className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6"
                            >
                                <p className="text-sm uppercase tracking-[0.2em] text-sky-400">
                                    {insight.confidence} confidence
                                </p>
                                <h3 className="mt-3 text-lg font-semibold text-white">
                                    {insight.title}
                                </h3>
                                <p className="mt-3 text-sm text-slate-300">
                                    {insight.summary}
                                </p>
                                {insight.caution && (
                                    <p className="mt-3 text-xs text-amber-300">
                                        {insight.caution}
                                    </p>
                                )}
                            </article>
                        ))}
                    </div>
                ) : (
                    <div className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                        <p className="text-sm text-slate-300">
                            AI insights are not available for this report, but
                            the deterministic comparison metrics above are still
                            valid.
                        </p>
                    </div>
                )}

                {ai.focusRecommendation && (
                    <div className="rounded-3xl border border-sky-500/30 bg-sky-950/20 p-6">
                        <p className="text-sm uppercase tracking-[0.2em] text-sky-400">
                            Focus Recommendation
                        </p>
                        <h3 className="mt-3 text-xl font-semibold text-white">
                            {ai.focusRecommendation.title}
                        </h3>
                        <p className="mt-3 text-slate-200">
                            {ai.focusRecommendation.recommendation}
                        </p>
                        <p className="mt-3 text-sm text-slate-400">
                            {ai.focusRecommendation.reasoning}
                        </p>
                    </div>
                )}
            </div>

            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
                <Link to="/analyze">
                    <Button variant="secondary">Analyze another fight</Button>
                </Link>
                <Link to="/">
                    <Button variant="secondary">Back to home</Button>
                </Link>
            </div>
        </section>
    );
}
