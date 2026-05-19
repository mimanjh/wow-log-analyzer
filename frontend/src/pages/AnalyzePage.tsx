import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import {
    createReportJob,
    getAuthStatus,
    getBrowserCharacters,
    getCharacterReports,
    getCharacters,
    getFights,
} from "../lib/api";
import { matchFightCharacter } from "../lib/characterMatching";
import { usePageTitle } from "../hooks/usePageTitle";
import { getCharacterCardClasses } from "../lib/characterPresentation";
import { useAnalyzeStore } from "../stores/useAnalyzeStore";
import { useBrowserStore } from "../stores/useBrowserStore";

export function AnalyzePage() {
    usePageTitle("Select");
    const navigate = useNavigate();
    const hasFetchFailedRef = useRef(false);
    const [toastMessage, setToastMessage] = useState<string | null>(null);
    const [collapsedReports, setCollapsedReports] = useState<Set<string>>(
        new Set(),
    );
    const [creatingFightId, setCreatingFightId] = useState<number | null>(null);
    const [pendingFight, setPendingFight] = useState<{
        reportCode: string;
        fight: NonNullable<(typeof reports)[0]["fights"]>[0];
    } | null>(null);
    const {
        auth,
        characters,
        selectedCharacter,
        reports,
        nextCursor,
        hasMoreReports,
        isAuthLoading,
        isCharactersLoading,
        isReportsLoading,
        error,
        setAuth,
        finishCharactersLoad,
        setSelectedCharacter,
        resetReports,
        appendReports,
        setLoadingState,
        setError,
    } = useBrowserStore();
    const { setReportUrl, setReportJob } = useAnalyzeStore();

    useEffect(() => {
        if (!toastMessage) {
            return;
        }

        const timeout = window.setTimeout(() => {
            setToastMessage(null);
        }, 2800);

        return () => window.clearTimeout(timeout);
    }, [toastMessage]);

    useEffect(() => {
        let cancelled = false;
        async function load() {
            try {
                let status = auth;
                if (status === null) {
                    setLoadingState("isAuthLoading", true);
                    status = await getAuthStatus();
                    if (cancelled) {
                        return;
                    }

                    setAuth(status);
                    setLoadingState("isAuthLoading", false);
                }

                if (!status.authenticated) {
                    finishCharactersLoad([]);
                    return;
                }

                if (characters.length > 0) {
                    return;
                }

                setLoadingState("isCharactersLoading", true);
                const nextCharacters = await getBrowserCharacters();
                if (cancelled) {
                    return;
                }

                finishCharactersLoad(nextCharacters);
            } catch (err) {
                if (cancelled) {
                    return;
                }

                setError(
                    err instanceof Error
                        ? err.message
                        : "Failed to load Warcraft Logs browser state",
                );
            }
        }

        void load();

        return () => {
            cancelled = true;
        };
    }, [
        auth,
        characters.length,
        finishCharactersLoad,
        setAuth,
        setError,
        setLoadingState,
    ]);

    useEffect(() => {
        if (!selectedCharacter || reports.length > 0) {
            return;
        }

        const selectedCharacterId = selectedCharacter.id;
        let cancelled = false;

        async function loadInitialReports() {
            try {
                setLoadingState("isReportsLoading", true);
                const page = await getCharacterReports(selectedCharacterId);
                if (cancelled) {
                    return;
                }

                appendReports(page);
                if (page.reports.length > 1) {
                    setCollapsedReports(
                        new Set(page.reports.slice(1).map((r) => r.code)),
                    );
                }
                setToastMessage("Recent logs loaded.");
            } catch (err) {
                if (cancelled) {
                    return;
                }

                hasFetchFailedRef.current = true;
                setError(
                    err instanceof Error
                        ? err.message
                        : "Failed to load character reports",
                );
            }
        }

        void loadInitialReports();

        return () => {
            cancelled = true;
        };
    }, [
        appendReports,
        reports.length,
        selectedCharacter,
        setError,
        setLoadingState,
    ]);

    async function loadMoreReports() {
        if (
            !selectedCharacter ||
            isReportsLoading ||
            hasFetchFailedRef.current
        ) {
            return;
        }

        try {
            setLoadingState("isReportsLoading", true);
            const page = await getCharacterReports(
                selectedCharacter.id,
                nextCursor,
            );
            appendReports(page);
            setCollapsedReports((prev) => {
                const next = new Set(prev);
                for (const r of page.reports) next.add(r.code);
                return next;
            });
        } catch (err) {
            hasFetchFailedRef.current = true;
            setError(
                err instanceof Error
                    ? err.message
                    : "Failed to load more reports",
            );
        }
    }

    async function handleSelectFight(reportCode: string, fightId: number) {
        if (creatingFightId !== null) return;
        setCreatingFightId(fightId);
        setError(null);

        const reportUrl = `https://www.warcraftlogs.com/reports/${reportCode}?fight=${fightId}`;
        setReportUrl(reportUrl);

        try {
            const fights = await getFights(reportCode, fightId);
            const fight = fights.find((f) => f.id === fightId) ?? fights[0];
            if (!fight) {
                throw new Error("Fight not found in report");
            }

            const fightCharacters = await getCharacters(reportCode, fight.id);
            const character = matchFightCharacter(
                selectedCharacter,
                fightCharacters,
            );
            if (!character) {
                throw new Error(
                    "Your character was not found among the fight participants",
                );
            }

            const job = await createReportJob(reportCode, fight, character);
            setReportJob(job);
            navigate("/report");
        } catch (err) {
            setError(
                err instanceof Error ? err.message : "Failed to start report",
            );
        } finally {
            setCreatingFightId(null);
        }
    }

    async function refreshReports() {
        if (!selectedCharacter) {
            return;
        }

        hasFetchFailedRef.current = false;
        setLoadingState("isReportsLoading", true);
        setError(null);

        try {
            const page = await getCharacterReports(selectedCharacter.id);
            resetReports();
            appendReports(page);
            setToastMessage(
                `${selectedCharacter.name}'s recent log reports refreshed.`,
            );
        } catch (err) {
            hasFetchFailedRef.current = true;
            setError(
                err instanceof Error
                    ? err.message
                    : "Failed to refresh character reports",
            );
        }
    }

    async function refreshCharacters() {
        setLoadingState("isCharactersLoading", true);
        setError(null);

        try {
            const nextCharacters = await getBrowserCharacters();
            finishCharactersLoad(nextCharacters);
            setToastMessage("Characters refreshed.");
        } catch (err) {
            setError(
                err instanceof Error
                    ? err.message
                    : "Failed to refresh characters",
            );
        }
    }

    function handleCharacterPick(characterId: number) {
        const character =
            characters.find((entry) => entry.id === characterId) ?? null;
        hasFetchFailedRef.current = false;
        setCollapsedReports(new Set());
        setSelectedCharacter(character);
        resetReports();
    }

    function toggleReportCollapse(code: string) {
        setCollapsedReports((prev) => {
            const next = new Set(prev);
            if (next.has(code)) {
                next.delete(code);
            } else {
                next.add(code);
            }
            return next;
        });
    }

    const reportGroups = reports.filter(
        (report) => report.fights && report.fights.length > 0,
    );

    function groupFightsByBoss(
        fights: NonNullable<(typeof reports)[0]["fights"]>,
    ) {
        const order: string[] = [];
        const map = new Map<string, typeof fights>();
        for (const fight of fights) {
            const key = fight.name.trim() || "Unknown";
            if (!map.has(key)) {
                order.push(key);
                map.set(key, []);
            }
            map.get(key)!.push(fight);
        }
        return order.map((bossName) => ({
            bossName,
            fights: map.get(bossName)!,
        }));
    }

    if (isAuthLoading) {
        return (
            <section className="space-y-8">
                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <p className="text-sm text-slate-400">
                        Checking Warcraft Logs session...
                    </p>
                </div>
            </section>
        );
    }

    if (!auth?.authenticated) {
        return (
            <section className="space-y-8">
                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                        Select
                    </p>
                    <h1 className="mt-3 text-3xl font-semibold text-white">
                        Login required
                    </h1>
                    <p className="mt-4 max-w-2xl text-slate-300">
                        Sign in with Warcraft Logs first so the app can load
                        your available characters and recent reports.
                    </p>
                    <div className="mt-8 flex gap-3">
                        <a href="/api/auth/login">
                            <Button>Log in with Warcraft Logs</Button>
                        </a>
                        <Link to="/">
                            <Button variant="secondary">Back to home</Button>
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
                    Select
                </p>
                <h1 className="mt-3 text-3xl font-semibold text-white">
                    Select a character and log
                </h1>
                <p className="mt-4 max-w-2xl text-slate-300">
                    Choose a character, then select one of their recent reports
                    to continue. The app loads the next 10 reports as you
                    continue down the list.
                </p>
            </div>

            {error && (
                <div className="rounded-3xl border border-rose-800 bg-rose-950/20 p-6">
                    <p className="text-sm text-rose-400">{error}</p>
                </div>
            )}

            <div className="grid gap-6 lg:grid-cols-[320px_1fr]">
                <aside className="rounded-3xl border border-slate-800 bg-slate-900/80 p-6">
                    <div className="flex items-center justify-between gap-3">
                        <h2 className="text-lg font-semibold text-white">
                            Characters
                        </h2>
                        <Button
                            type="button"
                            variant="secondary"
                            onClick={() => void refreshCharacters()}
                            disabled={isCharactersLoading}
                        >
                            {isCharactersLoading ? "SYNCING..." : "SYNC"}
                        </Button>
                    </div>
                    {isCharactersLoading ? (
                        <p className="mt-4 text-sm text-slate-400">
                            Loading characters...
                        </p>
                    ) : (
                        <div className="mt-4 space-y-3">
                            {characters.map((character) => (
                                <button
                                    key={`${character.id}-${character.serverSlug ?? character.serverName}`}
                                    type="button"
                                    onClick={() =>
                                        handleCharacterPick(character.id)
                                    }
                                    className={`w-full rounded-3xl border-2 p-5 text-left transition ${getCharacterCardClasses(
                                        character.class,
                                        selectedCharacter?.id === character.id,
                                    )} active:scale-[0.99]`}
                                >
                                    <p className="text-lg font-bold text-white">
                                        {character.name}
                                    </p>
                                    <p className="mt-1 text-xs text-slate-400">
                                        {character.serverName} |{" "}
                                        {character.serverRegion}
                                    </p>
                                    <p className="mt-2 text-sm text-slate-200">
                                        {character.class}
                                    </p>
                                </button>
                            ))}
                        </div>
                    )}
                </aside>

                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-6">
                    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                        <h2 className="text-lg font-semibold text-white">
                            {selectedCharacter
                                ? `${selectedCharacter.name}'s recent log reports`
                                : "Select a character"}
                        </h2>
                        {selectedCharacter && (
                            <Button
                                type="button"
                                variant="secondary"
                                onClick={() => void refreshReports()}
                                disabled={isReportsLoading}
                            >
                                {isReportsLoading ? "REFRESHING..." : "REFRESH"}
                            </Button>
                        )}
                    </div>

                    {!selectedCharacter ? (
                        <p className="mt-4 text-sm text-slate-400">
                            Choose a character from the left to start loading
                            reports.
                        </p>
                    ) : (
                        <div className="mt-6 space-y-4">
                            {reportGroups.map((report) => {
                                const isCollapsed = collapsedReports.has(
                                    report.code,
                                );
                                return (
                                    <article
                                        key={report.code}
                                        className="rounded-3xl border border-slate-800 bg-slate-950/80 p-5"
                                    >
                                        <button
                                            type="button"
                                            onClick={() =>
                                                toggleReportCollapse(
                                                    report.code,
                                                )
                                            }
                                            className="flex w-full items-center justify-between gap-3 text-left"
                                        >
                                            <div>
                                                <span className="text-sm font-medium text-slate-300">
                                                    {new Date(
                                                        report.startTime,
                                                    ).toLocaleString()}
                                                </span>
                                                {report.bossNames &&
                                                    report.bossNames.length >
                                                        0 && (
                                                        <div className="mt-1 flex flex-wrap gap-1">
                                                            {report.bossNames.map(
                                                                (name) => (
                                                                    <span
                                                                        key={
                                                                            name
                                                                        }
                                                                        className="rounded-full bg-emerald-950/60 px-2 py-0.5 text-xs text-emerald-400 ring-1 ring-emerald-700/40"
                                                                    >
                                                                        {name}
                                                                    </span>
                                                                ),
                                                            )}
                                                        </div>
                                                    )}
                                            </div>
                                            <svg
                                                className={`h-4 w-4 shrink-0 text-slate-400 transition-transform ${isCollapsed ? "-rotate-90" : ""}`}
                                                viewBox="0 0 16 16"
                                                fill="currentColor"
                                                aria-hidden="true"
                                            >
                                                <path d="M8 10.94L2.47 5.41l1.06-1.06L8 8.82l4.47-4.47 1.06 1.06z" />
                                            </svg>
                                        </button>

                                        {!isCollapsed && (
                                            <div className="mt-4 space-y-2">
                                                {groupFightsByBoss(
                                                    report.fights!,
                                                ).map((group, groupIndex) => (
                                                    <div
                                                        key={group.bossName}
                                                        className={`space-y-3 ${groupIndex > 0 ? "border-t border-slate-800 pt-4" : ""}`}
                                                    >
                                                        <h3 className="text-base font-semibold text-slate-100">
                                                            {group.bossName}
                                                        </h3>
                                                        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
                                                            {group.fights
                                                                .slice(0, 4)
                                                                .map(
                                                                    (fight) => {
                                                                        const isCreating =
                                                                            creatingFightId ===
                                                                            fight.id;
                                                                        const cardClasses =
                                                                            fight.kill
                                                                                ? "border-emerald-500/50 bg-emerald-950/20 hover:border-emerald-400/70"
                                                                                : "border-slate-800 bg-slate-900/50 hover:border-slate-700";
                                                                        return (
                                                                            <button
                                                                                key={
                                                                                    fight.id
                                                                                }
                                                                                type="button"
                                                                                disabled={
                                                                                    creatingFightId !==
                                                                                    null
                                                                                }
                                                                                onClick={() =>
                                                                                    setPendingFight(
                                                                                        {
                                                                                            reportCode:
                                                                                                report.code,
                                                                                            fight,
                                                                                        },
                                                                                    )
                                                                                }
                                                                                className={`flex min-h-16 w-full cursor-pointer items-center gap-3 rounded-2xl border p-4 text-left transition ${cardClasses} disabled:opacity-60`}
                                                                            >
                                                                                <span className="min-w-0 flex-1">
                                                                                    <span className="mt-1 block text-sm text-slate-400">
                                                                                        {isCreating ? (
                                                                                            "Loading..."
                                                                                        ) : (
                                                                                            <>
                                                                                                {
                                                                                                    fight.difficulty
                                                                                                }{" "}
                                                                                                {Math.floor(
                                                                                                    fight.killTime /
                                                                                                        60,
                                                                                                )}
                                                                                                :
                                                                                                {(
                                                                                                    fight.killTime %
                                                                                                    60
                                                                                                )
                                                                                                    .toString()
                                                                                                    .padStart(
                                                                                                        2,
                                                                                                        "0",
                                                                                                    )}
                                                                                            </>
                                                                                        )}
                                                                                    </span>
                                                                                </span>
                                                                                <span
                                                                                    className={`shrink-0 rounded-full px-3 py-1 text-xs font-semibold uppercase ${
                                                                                        fight.kill
                                                                                            ? "bg-emerald-400/15 text-emerald-200 ring-1 ring-emerald-400/40"
                                                                                            : "bg-slate-800 text-slate-300 ring-1 ring-slate-700"
                                                                                    }`}
                                                                                >
                                                                                    {fight.kill
                                                                                        ? "Kill"
                                                                                        : "Wipe"}
                                                                                </span>
                                                                            </button>
                                                                        );
                                                                    },
                                                                )}
                                                        </div>
                                                    </div>
                                                ))}
                                            </div>
                                        )}
                                    </article>
                                );
                            })}

                            {isReportsLoading && (
                                <p className="text-sm text-slate-400">
                                    Loading reports...
                                </p>
                            )}

                            {!isReportsLoading && reports.length === 0 && (
                                <p className="text-sm text-slate-400">
                                    No recent reports were found for this
                                    character.
                                </p>
                            )}

                            {!isReportsLoading &&
                                reports.length > 0 &&
                                reportGroups.length === 0 && (
                                    <p className="text-sm text-slate-400">
                                        No recent raid boss fights were found
                                        for this character.
                                    </p>
                                )}

                            {hasMoreReports && !hasFetchFailedRef.current && (
                                <div className="flex justify-center pt-2">
                                    <Button
                                        type="button"
                                        variant="secondary"
                                        onClick={() => void loadMoreReports()}
                                        disabled={isReportsLoading}
                                    >
                                        {isReportsLoading
                                            ? "LOADING..."
                                            : "LOAD MORE"}
                                    </Button>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>

            {toastMessage && (
                <div className="pointer-events-none fixed bottom-6 right-6 z-50">
                    <div className="rounded-2xl border border-emerald-500/40 bg-emerald-950/90 px-4 py-3 shadow-2xl shadow-emerald-950/30 backdrop-blur">
                        <p className="text-sm font-medium text-emerald-100">
                            {toastMessage}
                        </p>
                    </div>
                </div>
            )}

            {pendingFight && (
                <div
                    className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
                    onClick={() => setPendingFight(null)}
                >
                    <div
                        className="w-full max-w-sm rounded-3xl border border-slate-700 bg-slate-900 p-6 shadow-2xl"
                        onClick={(e) => e.stopPropagation()}
                    >
                        <h2 className="text-lg font-semibold text-white">
                            Confirm analysis
                        </h2>
                        <div className="mt-4 space-y-1">
                            <p className="text-sm text-slate-400">Character</p>
                            <p className="font-medium text-white">
                                {selectedCharacter?.name}
                                {selectedCharacter?.serverName && (
                                    <span className="ml-1 font-normal text-slate-400">
                                        — {selectedCharacter.serverName}
                                    </span>
                                )}
                            </p>
                        </div>
                        <div className="mt-3 space-y-1">
                            <p className="text-sm text-slate-400">Fight</p>
                            <p className="font-medium text-white">
                                {pendingFight.fight.name}
                            </p>
                            <p className="text-sm text-slate-400">
                                {pendingFight.fight.difficulty} &middot;{" "}
                                {pendingFight.fight.kill ? "Kill" : "Wipe"}{" "}
                                &middot;{" "}
                                {Math.floor(pendingFight.fight.killTime / 60)}:
                                {(pendingFight.fight.killTime % 60)
                                    .toString()
                                    .padStart(2, "0")}
                            </p>
                        </div>
                        <div className="mt-6 flex gap-3">
                            <Button
                                onClick={() => {
                                    const { reportCode, fight } = pendingFight;
                                    setPendingFight(null);
                                    void handleSelectFight(
                                        reportCode,
                                        fight.id,
                                    );
                                }}
                                disabled={creatingFightId !== null}
                            >
                                {creatingFightId !== null
                                    ? "Loading..."
                                    : "Analyze"}
                            </Button>
                            <Button
                                variant="secondary"
                                onClick={() => setPendingFight(null)}
                                disabled={creatingFightId !== null}
                            >
                                Cancel
                            </Button>
                        </div>
                    </div>
                </div>
            )}
        </section>
    );
}
