import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { BrowserState } from "../types";

const initialState = {
    auth: null,
    characters: [],
    selectedCharacter: null,
    reports: [],
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
                set({
                    characters,
                    isCharactersLoading: false,
                    error: null,
                }),

            setSelectedCharacter: (selectedCharacter) =>
                set({
                    selectedCharacter,
                    reports: [],
                    nextCursor: null,
                    hasMoreReports: false,
                    error: null,
                }),

            resetReports: () =>
                set({
                    reports: [],
                    nextCursor: null,
                    hasMoreReports: false,
                }),

            appendReports: (page) =>
                set((state) => ({
                    reports: [...state.reports, ...page.reports],
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
                characters: state.characters,
                selectedCharacter: state.selectedCharacter,
            }),
        },
    ),
);
