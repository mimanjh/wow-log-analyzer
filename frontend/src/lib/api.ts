import type {
  AuthStatus,
  BrowserCharacter,
  Character,
  CharacterReportsPage,
  Fight,
  ReportJob,
} from '../types'

export interface AnalyzeResponse {
  reportId: string
  preferredFightId?: number
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

export async function createReportJob(reportId: string, fight: Fight, character: Character): Promise<ReportJob> {
  const response = await fetch('/api/report/jobs', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ reportId, fight, character }),
  })

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to create report job')
  }

  return response.json()
}

export async function getReportJob(jobId: string): Promise<ReportJob> {
  const response = await fetch(`/api/report/jobs/${encodeURIComponent(jobId)}`)

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to load report job')
  }

  return response.json()
}

export async function getAuthStatus(): Promise<AuthStatus> {
  const response = await fetch('/api/auth/status', {
    credentials: 'include',
  })

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to load auth status')
  }

  return response.json()
}

export async function logout(): Promise<void> {
  const response = await fetch('/api/auth/logout', {
    method: 'POST',
    credentials: 'include',
  })

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to log out')
  }
}

export async function getBrowserCharacters(): Promise<BrowserCharacter[]> {
  const response = await fetch('/api/browser/characters', {
    credentials: 'include',
  })

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to load characters')
  }

  return response.json()
}

export async function getCharacterReports(
  characterId: number,
  cursor?: string | null,
  limit = 10,
): Promise<CharacterReportsPage> {
  const params = new URLSearchParams({
    limit: String(limit),
  })

  if (cursor) {
    params.set('cursor', cursor)
  }

  const response = await fetch(`/api/browser/characters/${characterId}/reports?${params.toString()}`, {
    credentials: 'include',
  })

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to load character reports')
  }

  return response.json()
}
