import { useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAnalyzeStore } from "../stores/useAnalyzeStore";
import { usePageTitle } from "../hooks/usePageTitle";
import { Button } from "../components/ui/Button";
import { analyzeReport, generateReport } from "../lib/api";

export function AnalyzePage() {
    usePageTitle("Analyze");
    const navigate = useNavigate();
    const {
        reportUrl,
        reportId,
        fights,
        characters,
        selectedFight,
        selectedCharacter,
        isLoading,
        error,
        setReportData,
        setSelectedFight,
        setSelectedCharacter,
        setReportResult,
        setLoading,
        setError,
    } = useAnalyzeStore();

    useEffect(() => {
        if (!reportUrl || reportId || isLoading) {
            return;
        }

        setLoading(true);
        setError(null);

        analyzeReport(reportUrl)
            .then((data) => {
                setReportData(data);
            })
            .catch((err) => {
                setError(err.message);
            });
    }, [
        reportUrl,
        isLoading,
        reportId,
        setReportData,
        setLoading,
        setError,
    ]);

    async function handleAnalyzeClick() {
        if (!reportId || !selectedFight || !selectedCharacter) {
            return;
        }

        setLoading(true);
        setError(null);

        try {
            const result = await generateReport(
                reportId,
                selectedFight,
                selectedCharacter,
            );
            setReportResult(result);
            navigate("/report");
        } catch (err) {
            setError(
                err instanceof Error
                    ? err.message
                    : "Failed to generate report",
            );
        } finally {
            setLoading(false);
        }
    }

    if (!reportUrl) {
        return (
            <section className="space-y-8">
                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                        Analyze
                    </p>
                    <h1 className="mt-3 text-3xl font-semibold text-white">
                        No report URL provided
                    </h1>
                    <p className="mt-4 text-slate-300">
                        Please go back to the home page and paste a Warcraft
                        Logs URL.
                    </p>
                    <div className="mt-8">
                        <Link to="/">
                            <Button>Back to home</Button>
                        </Link>
                    </div>
                </div>
            </section>
        );
    }

    return (
        <section className="space-y-8">
            <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                    Analyze
                </p>
                <h1 className="mt-3 text-3xl font-semibold text-white">
                    Fight and character selection
                </h1>
                <p className="mt-4 max-w-2xl text-slate-300">
                    Select a fight and character from your report to analyze.
                </p>

                <div className="mt-8 rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                    <p className="text-sm text-slate-400">Selected report</p>
                    <p className="mt-2 rounded-2xl bg-slate-900 px-4 py-3 text-sm text-slate-100">
                        {reportUrl}
                    </p>
                </div>

                {error && (
                    <div className="mt-6 rounded-3xl border border-rose-800 bg-rose-950/20 p-6">
                        <p className="text-sm text-rose-400">
                            Error loading report data: {error}
                        </p>
                    </div>
                )}

                {isLoading && (
                    <div className="mt-6 rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                        <p className="text-sm text-slate-400">
                            {reportId
                                ? "Generating report..."
                                : "Loading report data..."}
                        </p>
                    </div>
                )}

                {reportId && !isLoading && !error && (
                    <>
                        <div className="mt-6 rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                            <h2 className="text-lg font-semibold text-white">
                                Select a Fight
                            </h2>
                            <div className="mt-4 space-y-2">
                                {fights.map((fight) => (
                                    <label
                                        key={fight.id}
                                        className="flex items-center space-x-3"
                                    >
                                        <input
                                            type="radio"
                                            name="fight"
                                            value={fight.id}
                                            checked={
                                                selectedFight?.id === fight.id
                                            }
                                            onChange={() =>
                                                setSelectedFight(fight)
                                            }
                                            className="text-sky-400 focus:ring-sky-500"
                                        />
                                        <span className="text-slate-200">
                                            {fight.name} ({fight.difficulty}) -{" "}
                                            {Math.floor(fight.killTime / 60)}:
                                            {(fight.killTime % 60)
                                                .toString()
                                                .padStart(2, "0")}
                                        </span>
                                    </label>
                                ))}
                            </div>
                        </div>

                        {selectedFight && (
                            <div className="mt-6 rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                                <h2 className="text-lg font-semibold text-white">
                                    Select a Character
                                </h2>
                                <div className="mt-4 space-y-2">
                                    {characters.map((character) => (
                                        <label
                                            key={character.id}
                                            className="flex items-center space-x-3"
                                        >
                                            <input
                                                type="radio"
                                                name="character"
                                                value={character.id}
                                                checked={
                                                    selectedCharacter?.id ===
                                                    character.id
                                                }
                                                onChange={() =>
                                                    setSelectedCharacter(
                                                        character,
                                                    )
                                                }
                                                className="text-sky-400 focus:ring-sky-500"
                                            />
                                            <span className="text-slate-200">
                                                {character.name} -{" "}
                                                {character.class}{" "}
                                                {character.spec} (
                                                {character.role})
                                            </span>
                                        </label>
                                    ))}
                                </div>
                            </div>
                        )}
                    </>
                )}

                <div className="mt-8 flex flex-col gap-3 sm:flex-row">
                    <Link to="/">
                        <Button variant="secondary">Back to home</Button>
                    </Link>
                    {selectedFight && selectedCharacter && (
                        <Button
                            type="button"
                            disabled={isLoading}
                            onClick={handleAnalyzeClick}
                        >
                            {isLoading ? "Generating report..." : "Analyze fight"}
                        </Button>
                    )}
                </div>
            </div>
        </section>
    );
}
