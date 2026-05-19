import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { BrowserState } from "../types";

const initialState = {
    auth: null,
    characters: [],
    selectedCharacter: null,
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

            setAuth: (auth) => set({ auth }),

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
                set({
                    selectedCharacter,
                    reports: [],
                    reportsCachedAt: null,
                    nextCursor: null,
                    hasMoreReports: false,
                    error: null,
                }),

            resetReports: () =>
                set({
                    reports: [],
                    reportsCachedAt: null,
                    nextCursor: null,
                    hasMoreReports: false,
                }),

            appendReports: (page) =>
                set((state) => ({
                    reports: [...state.reports, ...page.reports],
                    reportsCachedAt: Date.now(),
                    nextCursor: page.nextCursor,
                    hasMoreReports: page.hasMore,
                    isReportsLoading: false,
                    error: null,
                })),

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
                selectedCharacter: state.selectedCharacter,
            }),
            merge: (persistedState, currentState) => {
                const persisted = (persistedState ?? {}) as Partial<BrowserState>;

                return {
                    ...currentState,
                    auth: persisted.auth ?? currentState.auth,
                    selectedCharacter:
                        persisted.selectedCharacter ??
                        currentState.selectedCharacter,
                    reports: [],
                    reportsCachedAt: null,
                    nextCursor: null,
                    hasMoreReports: false,
                };
            },
        },
    ),
);
