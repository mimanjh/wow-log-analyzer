import { useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAnalyzeStore } from "../stores/useAnalyzeStore";
import { usePageTitle } from "../hooks/usePageTitle";
import { Button } from "../components/ui/Button";
import { createReportJob, getCharacters } from "../lib/api";

function getClassBorderClasses(characterClass: string, selected: boolean) {
    const palette: Record<string, string> = {
        "Death Knight": selected
            ? "border-red-500 bg-red-950/20"
            : "border-red-700/60 bg-slate-950/80 hover:border-red-500/70",
        "Demon Hunter": selected
            ? "border-violet-500 bg-violet-950/20"
            : "border-violet-700/60 bg-slate-950/80 hover:border-violet-500/70",
        Druid: selected
            ? "border-orange-500 bg-orange-950/20"
            : "border-orange-700/60 bg-slate-950/80 hover:border-orange-500/70",
        Evoker: selected
            ? "border-emerald-500 bg-emerald-950/20"
            : "border-emerald-700/60 bg-slate-950/80 hover:border-emerald-500/70",
        Hunter: selected
            ? "border-lime-500 bg-lime-950/20"
            : "border-lime-700/60 bg-slate-950/80 hover:border-lime-500/70",
        Mage: selected
            ? "border-sky-400 bg-sky-950/20"
            : "border-sky-700/60 bg-slate-950/80 hover:border-sky-400/70",
        Monk: selected
            ? "border-teal-500 bg-teal-950/20"
            : "border-teal-700/60 bg-slate-950/80 hover:border-teal-500/70",
        Paladin: selected
            ? "border-pink-400 bg-pink-950/20"
            : "border-pink-700/60 bg-slate-950/80 hover:border-pink-400/70",
        Priest: selected
            ? "border-stone-300 bg-stone-950/20"
            : "border-stone-600/70 bg-slate-950/80 hover:border-stone-300/70",
        Rogue: selected
            ? "border-amber-400 bg-amber-950/20"
            : "border-amber-700/60 bg-slate-950/80 hover:border-amber-400/70",
        Shaman: selected
            ? "border-blue-500 bg-blue-950/20"
            : "border-blue-700/60 bg-slate-950/80 hover:border-blue-500/70",
        Warlock: selected
            ? "border-fuchsia-500 bg-fuchsia-950/20"
            : "border-fuchsia-700/60 bg-slate-950/80 hover:border-fuchsia-500/70",
        Warrior: selected
            ? "border-yellow-700 bg-yellow-950/20"
            : "border-yellow-800/80 bg-slate-950/80 hover:border-yellow-700/80",
    };

    return (
        palette[characterClass] ??
        (selected
            ? "border-sky-500 bg-sky-950/20"
            : "border-slate-700 bg-slate-950/80 hover:border-slate-500")
    );
}

export function FightSelectPage() {
    usePageTitle("Select Fight");
    const navigate = useNavigate();
    const {
        reportUrl,
        reportId,
        fights,
        characters,
        charactersFightId,
        selectedFight,
        selectedCharacter,
        isLoading,
        error,
        setCharactersForFight,
        setSelectedFight,
        setSelectedCharacter,
        setReportJob,
        setReportResult,
        setLoading,
        setError,
    } = useAnalyzeStore();

    useEffect(() => {
        if (
            !reportId ||
            !selectedFight ||
            isLoading ||
            (charactersFightId === selectedFight.id && characters.length > 0)
        ) {
            return;
        }

        setLoading(true);
        setError(null);

        getCharacters(reportId, selectedFight.id)
            .then((nextCharacters) => {
                setCharactersForFight(selectedFight.id, nextCharacters);
                setSelectedCharacter(null);
            })
            .catch((err) => {
                setError(
                    err instanceof Error
                        ? err.message
                        : "Failed to load characters",
                );
            });
    }, [
        reportId,
        selectedFight,
        charactersFightId,
        characters.length,
        isLoading,
        setCharactersForFight,
        setSelectedCharacter,
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
            setReportResult(null);
            const job = await createReportJob(
                reportId,
                selectedFight,
                selectedCharacter,
            );
            setReportJob(job);
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

    if (!reportUrl || !reportId) {
        return (
            <section className="space-y-8">
                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                        Select Fight
                    </p>
                    <h1 className="mt-3 text-3xl font-semibold text-white">
                        No report selected
                    </h1>
                    <p className="mt-4 text-slate-300">
                        Go back to your character logs and choose a report to
                        analyze.
                    </p>
                    <div className="mt-8">
                        <Link to="/analyze">
                            <Button>Back to logs</Button>
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
                    Select Fight
                </p>
                <h1 className="mt-3 text-3xl font-semibold text-white">
                    Fight and character selection
                </h1>
                <p className="mt-4 max-w-2xl text-slate-300">
                    Select a fight and character from your chosen report to
                    analyze.
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
                            {selectedFight && charactersFightId !== selectedFight.id
                                ? "Loading characters..."
                                : "Generating report..."}
                        </p>
                    </div>
                )}

                {!isLoading && !error && (
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
                                            checked={selectedFight?.id === fight.id}
                                            onChange={() =>
                                                setSelectedFight(fight)
                                            }
                                            className="text-sky-400 focus:ring-sky-500"
                                        />
                                        <span className="text-slate-200">
                                            {fight.name} ({fight.difficulty},{" "}
                                            {fight.kill ? "Kill" : "Wipe"}) -{" "}
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
                                <div className="mt-4 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                                    {characters.map((character) => (
                                        <label
                                            key={character.id}
                                            className={`flex cursor-pointer items-start gap-3 rounded-3xl border-2 p-5 transition ${getClassBorderClasses(
                                                character.class,
                                                selectedCharacter?.id ===
                                                    character.id,
                                            )}`}
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
                                                className="mt-1 text-sky-400 focus:ring-sky-500"
                                            />
                                            <span className="min-w-0">
                                                <span className="block text-lg font-bold text-white">
                                                    {character.name}
                                                </span>
                                                <span className="mt-1 block text-xs text-slate-400">
                                                    {character.serverName ||
                                                        "Unknown server"}
                                                </span>
                                                <span className="mt-2 block text-sm text-slate-200">
                                                    {character.class} -{" "}
                                                    {character.spec || "Unknown Spec"}
                                                </span>
                                            </span>
                                        </label>
                                    ))}
                                </div>
                            </div>
                        )}
                    </>
                )}

                <div className="mt-8 flex flex-col gap-3 sm:flex-row">
                    <Link to="/analyze">
                        <Button variant="secondary">Back to logs</Button>
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
