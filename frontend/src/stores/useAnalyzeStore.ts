import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { AnalyzeState } from '../types'

const initialState = {
  reportUrl: '',
  reportId: null,
  preferredFightId: null,
  fights: [],
  characters: [],
  charactersFightId: null,
  selectedFight: null,
  selectedCharacter: null,
  reportJob: null,
  reportResult: null,
  isLoading: false,
  error: null,
}

export const useAnalyzeStore = create<AnalyzeState>()(
  persist(
    (set) => ({
      ...initialState,

      setReportUrl: (reportUrl) =>
        set({
          ...initialState,
          reportUrl,
        }),

        setReportData: ({ reportId, preferredFightId, fights, characters }) => {
        const initialFight =
          fights.find((fight) => fight.id === preferredFightId) ?? fights[0] ?? null

        set({
          reportId,
          preferredFightId: preferredFightId ?? null,
          fights,
          characters,
          charactersFightId: characters.length > 0 ? initialFight?.id ?? null : null,
          selectedFight: initialFight,
          selectedCharacter: null,
          reportJob: null,
          reportResult: null,
          isLoading: false,
          error: null,
        })
      },

      setCharactersForFight: (fightId, characters) => set({
        charactersFightId: fightId,
        characters,
        isLoading: false,
        error: null,
      }),

      setSelectedFight: (selectedFight) => set({ selectedFight }),

      setSelectedCharacter: (selectedCharacter) => set({ selectedCharacter }),

      setReportJob: (reportJob) => set({ reportJob }),

      setReportResult: (reportResult) => set({ reportResult }),

      setLoading: (isLoading) => set({ isLoading }),

      setError: (error) => set({ error, isLoading: false }),

      reset: () => set(initialState),
    }),
    {
      name: 'wow-log-analyzer-analyze',
      partialize: (state) => ({
        reportUrl: state.reportUrl,
      }),
    },
  ),
)
