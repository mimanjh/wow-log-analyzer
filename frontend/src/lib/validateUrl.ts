const REPORT_ID_REGEX = /\/reports\/([A-Za-z0-9]+)/i
const HOST_REGEX = /^https?:\/\/(?:www\.)?warcraftlogs\.com\/reports\//i

export interface UrlValidationResult {
  isValid: boolean
  reportId?: string
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

  return { isValid: true, reportId }
}

export function isValidWarcraftLogsUrl(value: string): boolean {
  return validateWarcraftLogsUrl(value).isValid
}

export function extractReportId(value: string): string | undefined {
  return validateWarcraftLogsUrl(value).reportId
}
