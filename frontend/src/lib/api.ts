import type {
  AbilityTimelineResponse,
  AuthStatus,
  BrowserCharacter,
  BuffTimelineResponse,
  Character,
  CharacterReportsPage,
  Fight,
  ReportJob,
  ResourceTimelineResponse,
} from '../types'

export interface AnalyzeResponse {
  reportId: string
  preferredFightId?: number
  fights: Fight[]
  characters: Character[]
}

export interface FightCharacterFilter {
  name: string
  serverName?: string
  serverSlug?: string
  className?: string
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

export async function getFights(
  reportId: string,
  preferredFightId?: number | null,
  character?: FightCharacterFilter | null,
): Promise<Fight[]> {
  const params = new URLSearchParams({
    reportId,
  })

  if (preferredFightId) {
    params.set('preferredFightId', String(preferredFightId))
  }
  if (character?.name) {
    params.set('characterName', character.name)
  }
  if (character?.serverName) {
    params.set('serverName', character.serverName)
  }
  if (character?.serverSlug) {
    params.set('serverSlug', character.serverSlug)
  }
  if (character?.className) {
    params.set('className', character.className)
  }

  const response = await fetch(`/api/analyze/fights?${params.toString()}`)

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to load fights')
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

export async function getAbilityTimeline(
  jobId: string,
  abilityId: number,
): Promise<AbilityTimelineResponse> {
  const params = new URLSearchParams({
    abilityId: String(abilityId),
  })

  const response = await fetch(
    `/api/report/jobs/${encodeURIComponent(jobId)}/ability-timeline?${params.toString()}`,
  )

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to load ability timeline')
  }

  return response.json()
}

export async function getBuffTimeline(
  jobId: string,
  abilityId: number,
): Promise<BuffTimelineResponse> {
  const params = new URLSearchParams({
    abilityId: String(abilityId),
  })

  const response = await fetch(
    `/api/report/jobs/${encodeURIComponent(jobId)}/buff-timeline?${params.toString()}`,
  )

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to load buff timeline')
  }

  return response.json()
}

export async function getResourceTimeline(
  jobId: string,
  resourceTypeId: number,
): Promise<ResourceTimelineResponse> {
  const params = new URLSearchParams({
    resourceTypeId: String(resourceTypeId),
  })

  const response = await fetch(
    `/api/report/jobs/${encodeURIComponent(jobId)}/resource-timeline?${params.toString()}`,
  )

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to load resource timeline')
  }

  return response.json()
}

export interface BillingStatus {
  tier: string
  subscription?: {
    status: string
    currentPeriodEnd?: string
  }
}

export async function getBillingStatus(): Promise<BillingStatus> {
  const response = await fetch('/api/billing/status', { credentials: 'include' })
  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to load billing status')
  }
  return response.json()
}

export async function createCheckoutSession(): Promise<string> {
  const response = await fetch('/api/billing/checkout', {
    method: 'POST',
    credentials: 'include',
  })
  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to create checkout session')
  }
  const data = await response.json() as { url: string }
  return data.url
}

export async function createPortalSession(): Promise<string> {
  const response = await fetch('/api/billing/portal', {
    method: 'POST',
    credentials: 'include',
  })
  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to create portal session')
  }
  const data = await response.json() as { url: string }
  return data.url
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

export interface SavedReport {
  reportId: string
  fightId: number
  characterId: number
  encounterName: string
  difficulty: string
  characterName: string
  characterClass: string
  characterSpec: string
  analyzedAt: string
  cached: boolean
}

export async function listSavedReports(): Promise<SavedReport[]> {
  const response = await fetch('/api/reports', {
    credentials: 'include',
  })

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(errorText || 'Failed to load saved reports')
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
