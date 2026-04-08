import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { usePageTitle } from "../hooks/usePageTitle";
import { getAuthStatus, getBrowserCharacters } from "../lib/api";
import { useBrowserStore } from "../stores/useBrowserStore";

function getClassBorderClasses(characterClass: string) {
    const palette: Record<string, string> = {
        "Death Knight":
            "border-red-700/60 bg-slate-950/80 hover:border-red-500/70",
        "Demon Hunter":
            "border-violet-700/60 bg-slate-950/80 hover:border-violet-500/70",
        Druid: "border-orange-700/60 bg-slate-950/80 hover:border-orange-500/70",
        Evoker:
            "border-emerald-700/60 bg-slate-950/80 hover:border-emerald-500/70",
        Hunter: "border-lime-700/60 bg-slate-950/80 hover:border-lime-500/70",
        Mage: "border-sky-700/60 bg-slate-950/80 hover:border-sky-400/70",
        Monk: "border-teal-700/60 bg-slate-950/80 hover:border-teal-500/70",
        Paladin: "border-pink-700/60 bg-slate-950/80 hover:border-pink-400/70",
        Priest:
            "border-stone-600/70 bg-slate-950/80 hover:border-stone-300/70",
        Rogue: "border-amber-700/60 bg-slate-950/80 hover:border-amber-400/70",
        Shaman: "border-blue-700/60 bg-slate-950/80 hover:border-blue-500/70",
        Warlock:
            "border-fuchsia-700/60 bg-slate-950/80 hover:border-fuchsia-500/70",
        Warrior:
            "border-yellow-800/80 bg-slate-950/80 hover:border-yellow-700/80",
    };

    return (
        palette[characterClass] ??
        "border-slate-700 bg-slate-950/80 hover:border-slate-500"
    );
}

export function HomePage() {
    usePageTitle("Home");
    const navigate = useNavigate();
    const {
        auth,
        characters,
        isAuthLoading,
        isCharactersLoading,
        error,
        setAuth,
        finishCharactersLoad,
        setSelectedCharacter,
        setLoadingState,
        setError,
    } = useBrowserStore();

    useEffect(() => {
        let cancelled = false;

        async function load() {
            try {
                setError(null);
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
                        : "Failed to load Warcraft Logs account state",
                );
            }
        }

        void load();

        return () => {
            cancelled = true;
        };
    }, [
        auth,
        finishCharactersLoad,
        setAuth,
        setError,
        setLoadingState,
    ]);

    function handleCharacterClick(characterId: number) {
        const selected =
            characters.find((character) => character.id === characterId) ??
            null;
        setSelectedCharacter(selected);
        navigate("/analyze");
    }

    return (
        <section className="space-y-8">
            <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8 shadow-xl shadow-slate-950/20">
                <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                    Get started
                </p>
                <h1 className="mt-3 text-4xl font-semibold text-white sm:text-5xl">
                    Sign in with Warcraft Logs.
                </h1>
                <p className="mt-4 max-w-2xl text-slate-300">
                    After you log in, choose one of your characters and browse
                    recent reports without pasting URLs manually.
                </p>
            </div>

            {error && (
                <div className="rounded-3xl border border-rose-800 bg-rose-950/20 p-6">
                    <p className="text-sm text-rose-400">{error}</p>
                </div>
            )}

            {!auth?.authenticated ? (
                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <h2 className="text-2xl font-semibold text-white">
                        Connect your account
                    </h2>
                    <p className="mt-4 max-w-2xl text-slate-300">
                        Use Warcraft Logs OAuth to load your claimed
                        characters and browse recent personal logs.
                    </p>
                    <div className="mt-8">
                        <a href="/api/auth/login">
                            <Button disabled={isAuthLoading}>
                                {isAuthLoading
                                    ? "Checking session..."
                                    : "Log in with Warcraft Logs"}
                            </Button>
                        </a>
                    </div>
                </div>
            ) : (
                <div className="space-y-6">
                    <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-6">
                        <p className="text-sm uppercase tracking-[0.2em] text-sky-400">
                            Connected
                        </p>
                        <h2 className="mt-2 text-2xl font-semibold text-white">
                            {auth.user?.name ?? "Warcraft Logs account"}
                        </h2>
                        {auth.user?.battleTag && (
                            <p className="mt-2 text-sm text-slate-400">
                                {auth.user.battleTag}
                            </p>
                        )}
                    </div>

                    <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                        <h2 className="text-2xl font-semibold text-white">
                            Your characters
                        </h2>
                        <p className="mt-4 text-slate-300">
                            Pick a character to browse recent logs.
                        </p>

                        {isCharactersLoading ? (
                            <p className="mt-6 text-sm text-slate-400">
                                Loading characters...
                            </p>
                        ) : characters.length === 0 ? (
                            <p className="mt-6 text-sm text-slate-400">
                                No claimed Warcraft Logs characters were found
                                for this account.
                            </p>
                        ) : (
                            <div className="mt-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                                {characters.map((character) => (
                                    <button
                                        key={`${character.id}-${character.serverSlug ?? character.serverName}`}
                                        type="button"
                                        onClick={() =>
                                            handleCharacterClick(character.id)
                                        }
                                        className={`rounded-3xl border-2 p-6 text-left transition ${getClassBorderClasses(
                                            character.class,
                                        )}`}
                                    >
                                        <p className="text-xl font-bold text-white">
                                            {character.name}
                                        </p>
                                        <p className="mt-2 text-xs text-slate-400">
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
                    </div>
                </div>
            )}
        </section>
    );
}
