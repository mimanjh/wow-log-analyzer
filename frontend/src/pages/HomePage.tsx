import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { usePageTitle } from "../hooks/usePageTitle";
import { getCharacterCardClasses } from "../lib/characterPresentation";
import { getAuthStatus, getBrowserCharacters } from "../lib/api";
import { useBrowserStore } from "../stores/useBrowserStore";

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
                    After you log in, choose one of your characters and select
                    one of their recent reports without pasting URLs manually.
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
                            Pick a character to select from recent logs.
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
                                        className={`rounded-3xl border-2 p-6 text-left transition ${getCharacterCardClasses(
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
