import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { AnalyzeState } from '../types'

const initialState = {
  reportUrl: '',
  reportId: null,
  reportJob: null,
  reportResult: null,
  error: null,
}

export const useAnalyzeStore = create<AnalyzeState>()(
  persist(
    (set) => ({
      ...initialState,

      setReportUrl: (reportUrl) => {
        const match = reportUrl.match(/\/reports\/([^?#/]+)/)
        const reportId = match ? match[1] : null
        set({ ...initialState, reportUrl, reportId })
      },

      setReportJob: (reportJob) => set({ reportJob }),

      setReportResult: (reportResult) => set({ reportResult }),

      setError: (error) => set({ error }),

      reset: () => set(initialState),
    }),
    {
      name: 'wow-log-analyzer-analyze',
      partialize: (state) => ({
        reportUrl: state.reportUrl,
        reportId: state.reportId,
      }),
    },
  ),
)
