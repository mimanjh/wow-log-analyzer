import { create } from 'zustand'
import type { AnalyzeState } from '../types'

const initialState = {
  reportUrl: '',
  reportId: null,
  fights: [],
  characters: [],
  charactersFightId: null,
  selectedFight: null,
  selectedCharacter: null,
  reportResult: null,
  isLoading: false,
  error: null,
}

export const useAnalyzeStore = create<AnalyzeState>((set) => ({
  ...initialState,

  setReportUrl: (reportUrl) =>
    set({
      ...initialState,
      reportUrl,
    }),

  setReportData: ({ reportId, fights, characters }) => set({
    reportId,
    fights,
    characters,
    charactersFightId: fights[0]?.id ?? null,
    isLoading: false,
    error: null,
  }),

  setCharactersForFight: (fightId, characters) => set({
    charactersFightId: fightId,
    characters,
    isLoading: false,
    error: null,
  }),

  setSelectedFight: (selectedFight) => set({ selectedFight }),

  setSelectedCharacter: (selectedCharacter) => set({ selectedCharacter }),

  setReportResult: (reportResult) => set({ reportResult }),

  setLoading: (isLoading) => set({ isLoading }),

  setError: (error) => set({ error, isLoading: false }),

  reset: () => set(initialState),
}))
