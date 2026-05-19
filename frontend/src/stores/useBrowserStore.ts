import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { BrowserState, CharacterReportsCacheEntry } from "../types";

const initialState = {
    auth: null,
    authCachedAt: null,
    characters: [],
    selectedCharacter: null,
    reportCacheByCharacter: {} as Record<number, CharacterReportsCacheEntry>,
    reports: [],
    reportsCachedAt: null,
    nextCursor: null,
    hasMoreReports: false,
    isAuthLoading: false,
    isCharactersLoading: false,
    isReportsLoading: false,
    error: null,
};

export const useBrowserStore = create<BrowserState>()(
    persist(
        (set) => ({
            ...initialState,

            setAuth: (auth) => set({ auth, authCachedAt: auth !== null ? Date.now() : null }),

            setCharacters: (characters) => set({ characters }),

            finishCharactersLoad: (characters) =>
                set((state) => ({
                    characters,
                    selectedCharacter: state.selectedCharacter
                        ? characters.find(
                              (character) =>
                                  character.id === state.selectedCharacter?.id,
                          ) ?? null
                        : null,
                    isCharactersLoading: false,
                    error: null,
                })),

            setSelectedCharacter: (selectedCharacter) =>
                set((state) => {
                    const cached = selectedCharacter
                        ? (state.reportCacheByCharacter[selectedCharacter.id] ?? null)
                        : null;
                    return {
                        selectedCharacter,
                        reports: cached?.reports ?? [],
                        reportsCachedAt: cached?.cachedAt ?? null,
                        nextCursor: cached?.nextCursor ?? null,
                        hasMoreReports: cached?.hasMoreReports ?? false,
                        error: null,
                    };
                }),

            resetReports: () =>
                set((state) => {
                    const id = state.selectedCharacter?.id;
                    const newCache = id
                        ? { ...state.reportCacheByCharacter }
                        : state.reportCacheByCharacter;
                    if (id) delete newCache[id];
                    return {
                        reports: [],
                        reportsCachedAt: null,
                        nextCursor: null,
                        hasMoreReports: false,
                        reportCacheByCharacter: newCache,
                    };
                }),

            appendReports: (page) =>
                set((state) => {
                    const newReports = [...state.reports, ...page.reports];
                    const cachedAt = Date.now();
                    const id = state.selectedCharacter?.id;
                    const newCache = id
                        ? {
                              ...state.reportCacheByCharacter,
                              [id]: {
                                  reports: newReports,
                                  cachedAt,
                                  nextCursor: page.nextCursor,
                                  hasMoreReports: page.hasMore,
                              },
                          }
                        : state.reportCacheByCharacter;
                    return {
                        reports: newReports,
                        reportsCachedAt: cachedAt,
                        nextCursor: page.nextCursor,
                        hasMoreReports: page.hasMore,
                        reportCacheByCharacter: newCache,
                        isReportsLoading: false,
                        error: null,
                    };
                }),

            setLoadingState: (key, value) =>
                set({ [key]: value } as Pick<
                    BrowserState,
                    "isAuthLoading" | "isCharactersLoading" | "isReportsLoading"
                >),

            setError: (error) =>
                set({
                    error,
                    isAuthLoading: false,
                    isCharactersLoading: false,
                    isReportsLoading: false,
                }),

            reset: () => set(initialState),
        }),
        {
            name: "wow-log-browser",
            partialize: (state) => ({
                auth: state.auth,
                authCachedAt: state.authCachedAt,
                characters: state.characters,
                selectedCharacter: state.selectedCharacter,
                reportCacheByCharacter: state.reportCacheByCharacter,
            }),
            merge: (persistedState, currentState) => {
                const persisted = (persistedState ?? {}) as Partial<BrowserState>;

                const TTL_MS = 7 * 24 * 60 * 60 * 1000;
                const now = Date.now();

                const authStale =
                    !persisted.authCachedAt ||
                    now - persisted.authCachedAt > TTL_MS;

                const rawCache = persisted.reportCacheByCharacter ?? {};
                const freshCache: Record<number, CharacterReportsCacheEntry> = {};
                for (const [idStr, entry] of Object.entries(rawCache)) {
                    if (entry.cachedAt && now - entry.cachedAt <= TTL_MS) {
                        freshCache[Number(idStr)] = entry;
                    }
                }

                const selectedId = persisted.selectedCharacter?.id ?? null;
                const cachedForSelected = selectedId ? (freshCache[selectedId] ?? null) : null;

                return {
                    ...currentState,
                    auth: authStale ? null : (persisted.auth ?? null),
                    authCachedAt: authStale ? null : (persisted.authCachedAt ?? null),
                    characters: persisted.characters ?? currentState.characters,
                    selectedCharacter:
                        persisted.selectedCharacter ?? currentState.selectedCharacter,
                    reportCacheByCharacter: freshCache,
                    reports: cachedForSelected?.reports ?? [],
                    reportsCachedAt: cachedForSelected?.cachedAt ?? null,
                    nextCursor: cachedForSelected?.nextCursor ?? null,
                    hasMoreReports: cachedForSelected?.hasMoreReports ?? false,
                };
            },
        },
    ),
);
