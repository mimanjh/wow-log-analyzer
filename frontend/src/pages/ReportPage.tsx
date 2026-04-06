import { Link } from "react-router-dom";
import { useAnalyzeStore } from "../stores/useAnalyzeStore";
import { usePageTitle } from "../hooks/usePageTitle";
import { Button } from "../components/ui/Button";
import { MetricCard } from "../components/MetricCard";

export function ReportPage() {
    usePageTitle("Report");
    const { reportResult } = useAnalyzeStore();

    if (!reportResult) {
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
                        Please go back to the analyze page and select a fight
                        and character.
                    </p>
                    <div className="mt-8">
                        <Link to="/analyze">
                            <Button>Back to analyze</Button>
                        </Link>
                    </div>
                </div>
            </section>
        );
    }

    const { fight, character, comparison, ai } = reportResult;
    const { cohortStats, deltas } = comparison;

    return (
        <section className="space-y-8">
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

            {/* Fight and Character Summary */}
            <div className="grid gap-6 lg:grid-cols-2">
                <div className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                    <h2 className="text-lg font-semibold text-white">
                        Fight Summary
                    </h2>
                    <div className="mt-4 space-y-3">
                        <div>
                            <p className="text-sm text-slate-400">Encounter</p>
                            <p className="text-white">{fight.name}</p>
                        </div>
                        <div>
                            <p className="text-sm text-slate-400">Difficulty</p>
                            <p className="text-white">{fight.difficulty}</p>
                        </div>
                        <div>
                            <p className="text-sm text-slate-400">Kill Time</p>
                            <p className="text-white">
                                {Math.floor(fight.killTime / 60)}:
                                {(fight.killTime % 60)
                                    .toString()
                                    .padStart(2, "0")}
                            </p>
                        </div>
                    </div>
                </div>

                <div className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                    <h2 className="text-lg font-semibold text-white">
                        Character Summary
                    </h2>
                    <div className="mt-4 space-y-3">
                        <div>
                            <p className="text-sm text-slate-400">Name</p>
                            <p className="text-white">{character.name}</p>
                        </div>
                        <div>
                            <p className="text-sm text-slate-400">
                                Class & Spec
                            </p>
                            <p className="text-white">
                                {character.class} {character.spec}
                            </p>
                        </div>
                        <div>
                            <p className="text-sm text-slate-400">Role</p>
                            <p className="text-white">{character.role}</p>
                        </div>
                    </div>
                </div>
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
