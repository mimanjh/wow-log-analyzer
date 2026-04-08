import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import {
    analyzeReport,
    getAuthStatus,
    getBrowserCharacters,
    getCharacterReports,
} from "../lib/api";
import { usePageTitle } from "../hooks/usePageTitle";
import { useAnalyzeStore } from "../stores/useAnalyzeStore";
import { useBrowserStore } from "../stores/useBrowserStore";

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

export function AnalyzePage() {
    usePageTitle("Analyze");
    const navigate = useNavigate();
    const loadMoreRef = useRef<HTMLDivElement | null>(null);
    const [toastMessage, setToastMessage] = useState<string | null>(null);
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
    const { setReportUrl, setReportData } = useAnalyzeStore();

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
    }, [auth, finishCharactersLoad, setAuth, setError, setLoadingState]);

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
                setToastMessage("Recent logs loaded.");
            } catch (err) {
                if (cancelled) {
                    return;
                }

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

    useEffect(() => {
        const node = loadMoreRef.current;
        if (
            !node ||
            !selectedCharacter ||
            !hasMoreReports ||
            isReportsLoading
        ) {
            return;
        }

        const selectedCharacterId = selectedCharacter.id;
        const observer = new IntersectionObserver(
            (entries) => {
                const entry = entries[0];
                if (!entry?.isIntersecting) {
                    return;
                }

                void (async () => {
                    try {
                        setLoadingState("isReportsLoading", true);
                        const page = await getCharacterReports(
                            selectedCharacterId,
                            nextCursor,
                        );
                        appendReports(page);
                    } catch (err) {
                        setError(
                            err instanceof Error
                                ? err.message
                                : "Failed to load more reports",
                        );
                    }
                })();
            },
            {
                rootMargin: "200px",
            },
        );

        observer.observe(node);
        return () => observer.disconnect();
    }, [
        appendReports,
        hasMoreReports,
        isReportsLoading,
        nextCursor,
        selectedCharacter,
        setError,
        setLoadingState,
    ]);

    async function handleSelectReport(reportCode: string) {
        const reportUrl = `https://www.warcraftlogs.com/reports/${reportCode}`;
        setReportUrl(reportUrl);
        setLoadingState("isReportsLoading", true);
        setError(null);

        try {
            const data = await analyzeReport(reportUrl);
            setReportData(data);
            navigate("/select");
        } catch (err) {
            setError(
                err instanceof Error
                    ? err.message
                    : "Failed to load report details",
            );
        } finally {
            setLoadingState("isReportsLoading", false);
        }
    }

    async function refreshReports() {
        if (!selectedCharacter) {
            return;
        }

        setLoadingState("isReportsLoading", true);
        setError(null);

        try {
            const page = await getCharacterReports(selectedCharacter.id);
            resetReports();
            appendReports(page);
            setToastMessage(`${selectedCharacter.name}'s recent logs refreshed.`);
        } catch (err) {
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
        setSelectedCharacter(character);
        resetReports();
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
                        Analyze
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
                    Analyze
                </p>
                <h1 className="mt-3 text-3xl font-semibold text-white">
                    Browse your character logs
                </h1>
                <p className="mt-4 max-w-2xl text-slate-300">
                    Select a character, then scroll through recent reports. The
                    app loads the next 10 reports as you continue down the list.
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
                                    className={`w-full rounded-3xl border-2 p-5 text-left transition ${getClassBorderClasses(
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
                                ? `${selectedCharacter.name}'s recent logs`
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
                            {reports.map((report) => (
                                <article
                                    key={report.code}
                                    className="rounded-3xl border border-slate-800 bg-slate-950/80 p-5"
                                >
                                    <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                                        <div>
                                            <h3 className="text-lg font-semibold text-white">
                                                {report.title}
                                            </h3>
                                            <p className="mt-2 text-sm text-slate-300">
                                                {report.zoneName ||
                                                    "Unknown zone"}
                                            </p>
                                            {report.bossNames &&
                                                report.bossNames.length > 0 && (
                                                    <p className="mt-2 text-sm text-slate-400">
                                                        {report.bossNames.join(
                                                            ", ",
                                                        )}
                                                    </p>
                                                )}
                                            <p className="mt-2 text-xs text-slate-400">
                                                {new Date(
                                                    report.startTime,
                                                ).toLocaleString()}
                                            </p>
                                        </div>
                                        <Button
                                            type="button"
                                            onClick={() =>
                                                handleSelectReport(report.code)
                                            }
                                        >
                                            SELECT
                                        </Button>
                                    </div>
                                </article>
                            ))}

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

                            {hasMoreReports && (
                                <div
                                    ref={loadMoreRef}
                                    className="h-6 w-full"
                                    aria-hidden="true"
                                />
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
        </section>
    );
}
