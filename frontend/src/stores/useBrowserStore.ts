import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { BrowserState } from "../types";

const REPORTS_TTL_MS = 24 * 60 * 60 * 1000;

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
                          ) ?? state.selectedCharacter
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
                reports: state.reports,
                reportsCachedAt: state.reportsCachedAt,
                nextCursor: state.nextCursor,
                hasMoreReports: state.hasMoreReports,
            }),
            merge: (persistedState, currentState) => {
                const persisted = {
                    ...(persistedState as Partial<BrowserState>),
                };
                delete (persisted as Partial<BrowserState>).characters;

                const mergedState = {
                    ...currentState,
                    ...persisted,
                };
                const reportsCachedAt = mergedState.reportsCachedAt;

                if (
                    reportsCachedAt !== null &&
                    Date.now() - reportsCachedAt > REPORTS_TTL_MS
                ) {
                    return {
                        ...mergedState,
                        reports: [],
                        reportsCachedAt: null,
                        nextCursor: null,
                        hasMoreReports: false,
                    };
                }

                return mergedState;
            },
        },
    ),
);
