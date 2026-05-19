import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useAnalyzeStore } from "../stores/useAnalyzeStore";
import { useBrowserStore } from "../stores/useBrowserStore";
import { usePageTitle } from "../hooks/usePageTitle";
import { Button } from "../components/ui/Button";
import { createReportJob, getCharacters, getFights } from "../lib/api";
import { matchFightCharacter } from "../lib/characterMatching";
import { getCharacterCardClasses } from "../lib/characterPresentation";
import type { Character, Fight } from "../types";

const FIGHT_DISCOVERY_BATCH_SIZE = 10;

export function FightSelectPage() {
    usePageTitle("Select Fight");
    const navigate = useNavigate();
    const charactersByFightRef = useRef<Map<number, Character[]>>(new Map());
    const loadingFightIdsRef = useRef<Set<number>>(new Set());
    const loadedFightsReportIdRef = useRef<string | null>(null);
    const isLoadingFightsRef = useRef(false);
    const [isDiscoveringFights, setIsDiscoveringFights] = useState(false);
    const [discoveryProgress, setDiscoveryProgress] = useState({
        checked: 0,
        total: 0,
    });
    const browserSelectedCharacter = useBrowserStore(
        (state) => state.selectedCharacter,
    );
    const {
        reportUrl,
        reportId,
        preferredFightId,
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
        setFightsForReport,
        appendFightForReport,
    } = useAnalyzeStore();

    useEffect(() => {
        charactersByFightRef.current.clear();
        loadingFightIdsRef.current.clear();
        loadedFightsReportIdRef.current = null;
        isLoadingFightsRef.current = false;
        setDiscoveryProgress({ checked: 0, total: 0 });
    }, [reportId]);

    useEffect(() => {
        if (
            !reportId ||
            loadedFightsReportIdRef.current === reportId ||
            isLoadingFightsRef.current
        ) {
            return;
        }

        let cancelled = false;
        isLoadingFightsRef.current = true;
        setIsDiscoveringFights(true);
        setDiscoveryProgress({ checked: 0, total: 0 });
        setFightsForReport([]);
        setError(null);

        getFights(reportId, preferredFightId)
            .then(async (nextFights) => {
                if (cancelled) {
                    return;
                }
                loadedFightsReportIdRef.current = reportId;
                setDiscoveryProgress({ checked: 0, total: nextFights.length });

                if (!browserSelectedCharacter) {
                    setFightsForReport(nextFights);
                    setDiscoveryProgress({
                        checked: nextFights.length,
                        total: nextFights.length,
                    });
                    return;
                }

                for (
                    let index = 0;
                    index < nextFights.length;
                    index += FIGHT_DISCOVERY_BATCH_SIZE
                ) {
                    if (cancelled) {
                        return;
                    }

                    const fightBatch = nextFights.slice(
                        index,
                        index + FIGHT_DISCOVERY_BATCH_SIZE,
                    );
                    const batchResults = await Promise.all(
                        fightBatch.map(async (fight) => {
                            const fightCharacters = await getCharacters(
                                reportId,
                                fight.id,
                            );
                            return {
                                fight,
                                fightCharacters,
                                matchedCharacter: matchFightCharacter(
                                    browserSelectedCharacter,
                                    fightCharacters,
                                ),
                            };
                        }),
                    );

                    if (cancelled) {
                        return;
                    }

                    for (const result of batchResults) {
                        charactersByFightRef.current.set(
                            result.fight.id,
                            result.fightCharacters,
                        );
                        if (result.matchedCharacter) {
                            appendFightForReport(result.fight);
                        }
                    }
                    setDiscoveryProgress({
                        checked: Math.min(
                            index + FIGHT_DISCOVERY_BATCH_SIZE,
                            nextFights.length,
                        ),
                        total: nextFights.length,
                    });
                }
            })
            .catch((err) => {
                if (cancelled) {
                    return;
                }
                setError(
                    err instanceof Error ? err.message : "Failed to load fights",
                );
            })
            .finally(() => {
                isLoadingFightsRef.current = false;
                if (!cancelled) {
                    setIsDiscoveringFights(false);
                }
            });

        return () => {
            cancelled = true;
        };
    }, [
        reportId,
        preferredFightId,
        browserSelectedCharacter,
        appendFightForReport,
        setFightsForReport,
        setError,
    ]);

    useEffect(() => {
        if (!reportId || !selectedFight) {
            return;
        }

        const selectedFightId = selectedFight.id;
        if (charactersFightId === selectedFightId) {
            charactersByFightRef.current.set(selectedFightId, characters);

            const matchedCharacter = matchFightCharacter(
                browserSelectedCharacter,
                characters,
            );
            if (selectedCharacter?.id !== matchedCharacter?.id) {
                setSelectedCharacter(matchedCharacter);
            }
            return;
        }

        const cachedCharacters = charactersByFightRef.current.get(selectedFightId);
        if (cachedCharacters) {
            setCharactersForFight(selectedFightId, cachedCharacters);
            setSelectedCharacter(
                matchFightCharacter(browserSelectedCharacter, cachedCharacters),
            );
            return;
        }

        if (loadingFightIdsRef.current.has(selectedFightId)) {
            return;
        }

        loadingFightIdsRef.current.add(selectedFightId);
        setLoading(true);
        setError(null);

        getCharacters(reportId, selectedFightId)
            .then((nextCharacters) => {
                charactersByFightRef.current.set(selectedFightId, nextCharacters);
                setCharactersForFight(selectedFightId, nextCharacters);
                setSelectedCharacter(
                    matchFightCharacter(browserSelectedCharacter, nextCharacters),
                );
            })
            .catch((err) => {
                setError(
                    err instanceof Error
                        ? err.message
                        : "Failed to load characters",
                );
            })
            .finally(() => {
                loadingFightIdsRef.current.delete(selectedFightId);
            });
    }, [
        reportId,
        selectedFight,
        charactersFightId,
        characters,
        selectedCharacter,
        setCharactersForFight,
        setSelectedCharacter,
        setLoading,
        setError,
        browserSelectedCharacter,
    ]);

    useEffect(() => {
        if (!selectedFight || charactersFightId !== selectedFight.id) {
            return;
        }
        if (!browserSelectedCharacter) {
            return;
        }

        const matchedCharacter = matchFightCharacter(
            browserSelectedCharacter,
            characters,
        );
        if (!matchedCharacter) {
            return;
        }
        if (selectedCharacter?.id === matchedCharacter.id) {
            return;
        }

        setSelectedCharacter(matchedCharacter);
    }, [
        characters,
        charactersFightId,
        selectedCharacter,
        selectedFight,
        browserSelectedCharacter,
        setSelectedCharacter,
    ]);

    function getFightGroups() {
        const groups: Array<{ bossName: string; fights: Fight[] }> = [];
        const groupByBossName = new Map<string, Fight[]>();

        for (const fight of fights) {
            const bossName = fight.name.trim() || "Unknown encounter";
            const group = groupByBossName.get(bossName);
            if (group) {
                group.push(fight);
                continue;
            }

            const nextGroup = [fight];
            groupByBossName.set(bossName, nextGroup);
            groups.push({ bossName, fights: nextGroup });
        }

        return groups;
    }

    function renderFightOption(fight: Fight) {
        const isSelected = selectedFight?.id === fight.id;
        const rowClasses = fight.kill
            ? "border-emerald-500/50 bg-emerald-950/20 hover:border-emerald-400/70"
            : "border-slate-800 bg-slate-900/50 hover:border-slate-700";

        return (
            <label
                key={fight.id}
                className={`flex min-h-24 cursor-pointer items-center gap-3 rounded-2xl border p-4 transition ${rowClasses} ${
                    isSelected ? "ring-2 ring-sky-400/70" : ""
                }`}
            >
                <input
                    type="radio"
                    name="fight"
                    value={fight.id}
                    checked={isSelected}
                    onChange={() => setSelectedFight(fight)}
                    className="sr-only"
                />
                <span className="w-5 shrink-0" aria-hidden="true" />
                <span className="min-w-0 flex-1">
                    <span className="mt-1 block text-sm text-slate-400">
                        {fight.difficulty} - {Math.floor(fight.killTime / 60)}:
                        {(fight.killTime % 60).toString().padStart(2, "0")}
                    </span>
                </span>
                <span
                    className={`shrink-0 rounded-full px-3 py-1 text-xs font-semibold uppercase ${
                        fight.kill
                            ? "bg-emerald-400/15 text-emerald-200 ring-1 ring-emerald-400/40"
                            : "bg-slate-800 text-slate-300 ring-1 ring-slate-700"
                    }`}
                >
                    {fight.kill ? "Kill" : "Wipe"}
                </span>
            </label>
        );
    }

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
                        select.
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
                    Select a fight
                </h1>
                <p className="mt-4 max-w-2xl text-slate-300">
                    Choose the fight from your selected report and continue
                    with the character you already picked.
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
                            {fights.length === 0
                                ? "Loading fights..."
                                : selectedFight && charactersFightId !== selectedFight.id
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
                            {isDiscoveringFights && (
                                <p className="mt-2 text-sm text-slate-400">
                                    Loading matching fights{" "}
                                    {discoveryProgress.total > 0
                                        ? `(${discoveryProgress.checked}/${discoveryProgress.total})`
                                        : "..."}
                                </p>
                            )}
                            <div className="mt-4 space-y-2">
                                {getFightGroups().map((group, index) => (
                                    <div
                                        key={group.bossName}
                                        className={`space-y-3 ${
                                            index > 0
                                                ? "border-t border-slate-800 pt-4"
                                                : ""
                                        }`}
                                    >
                                        <h3 className="text-base font-semibold text-slate-100">
                                            {group.bossName}
                                        </h3>
                                        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                                            {group.fights
                                                .slice(0, 4)
                                                .map((fight) =>
                                                    renderFightOption(fight),
                                                )}
                                        </div>
                                    </div>
                                ))}
                                {!isDiscoveringFights && fights.length === 0 && (
                                    <p className="text-sm text-slate-400">
                                        No fights for the selected character were
                                        found in this report.
                                    </p>
                                )}
                            </div>
                        </div>

                        {selectedFight && (
                            <div className="mt-6 rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                                <h2 className="text-lg font-semibold text-white">
                                    Selected Character
                                </h2>
                                {selectedCharacter ? (
                                    <div
                                        className={`mt-4 rounded-3xl border-2 p-5 ${getCharacterCardClasses(
                                            selectedCharacter.class,
                                            true,
                                        )}`}
                                    >
                                        <p className="text-lg font-bold text-white">
                                            {selectedCharacter.name}
                                        </p>
                                        <p className="mt-1 text-xs text-slate-400">
                                            {selectedCharacter.serverName ||
                                                browserSelectedCharacter?.serverName ||
                                                "Unknown server"}
                                        </p>
                                        <p className="mt-2 text-sm text-slate-200">
                                            {selectedCharacter.class} -{" "}
                                            {selectedCharacter.spec ||
                                                "Unknown Spec"}
                                        </p>
                                    </div>
                                ) : (
                                    <p className="mt-4 text-sm text-slate-400">
                                        The selected character could not be
                                        matched in this fight. Go back and pick
                                        a different report if needed.
                                    </p>
                                )}
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
                            {isLoading ? "Generating report..." : "Generate report"}
                        </Button>
                    )}
                </div>
            </div>
        </section>
    );
}
