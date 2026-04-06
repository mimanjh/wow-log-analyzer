import type { Fight, Character, ReportResult } from '../types'

export interface AnalyzeResponse {
  reportId: string
  fights: Fight[]
  characters: Character[]
}

export async function analyzeReport(url: string): Promise<AnalyzeResponse> {
  const response = await fetch('/api/analyze/intake', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ url }),
  })

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to analyze report')
  }

  return response.json()
}

export async function getCharacters(reportId: string, fightId: number): Promise<Character[]> {
  const response = await fetch(`/api/analyze/characters?reportId=${encodeURIComponent(reportId)}&fightId=${fightId}`)

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to load characters')
  }

  return response.json()
}

export async function generateReport(reportId: string, fight: Fight, character: Character): Promise<ReportResult> {
  const response = await fetch('/api/report/generate', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ reportId, fight, character }),
  })

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to generate report')
  }

  return response.json()
}
