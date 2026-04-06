const REPORT_ID_REGEX = /\/reports\/([A-Za-z0-9]+)/i
const HOST_REGEX = /^https?:\/\/(?:www\.)?warcraftlogs\.com\/reports\//i

export interface UrlValidationResult {
  isValid: boolean
  reportId?: string
  fightId?: number
  error?: string
}

export function validateWarcraftLogsUrl(value: string): UrlValidationResult {
  const trimmed = value.trim()

  if (!trimmed) {
    return { isValid: false, error: "URL is required" }
  }

  if (!HOST_REGEX.test(trimmed)) {
    return { isValid: false, error: "URL must be from warcraftlogs.com/reports/" }
  }

  const match = trimmed.match(REPORT_ID_REGEX)
  if (!match || !match[1]) {
    return { isValid: false, error: "URL must contain a valid report code" }
  }

  const reportId = match[1]
  if (reportId.length < 6) {
    return { isValid: false, error: "Report code appears too short" }
  }

  if (reportId.length > 16) {
    return { isValid: false, error: "Report code appears too long" }
  }

  const fightId = extractFightId(trimmed)

  return { isValid: true, reportId, fightId }
}

export function isValidWarcraftLogsUrl(value: string): boolean {
  return validateWarcraftLogsUrl(value).isValid
}

export function extractReportId(value: string): string | undefined {
  return validateWarcraftLogsUrl(value).reportId
}

export function extractFightId(value: string): number | undefined {
  try {
    const parsed = new URL(value.trim())
    const queryFightId = parseFightId(parsed.searchParams.get("fight"))
    if (queryFightId) {
      return queryFightId
    }

    const hash = parsed.hash.startsWith("#") ? parsed.hash.slice(1) : parsed.hash
    const hashParams = new URLSearchParams(hash)
    return parseFightId(hashParams.get("fight"))
  } catch {
    return undefined
  }
}

function parseFightId(value: string | null): number | undefined {
  if (!value) {
    return undefined
  }

  const parsed = Number.parseInt(value, 10)
  if (Number.isNaN(parsed) || parsed <= 0) {
    return undefined
  }

  return parsed
}
