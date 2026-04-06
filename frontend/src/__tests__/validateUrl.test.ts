import { describe, expect, it } from 'vitest'
import { validateWarcraftLogsUrl, isValidWarcraftLogsUrl, extractFightId, extractReportId } from '../lib/validateUrl'

describe('validateWarcraftLogsUrl', () => {
  it('validates a correct Warcraft Logs URL', () => {
    const result = validateWarcraftLogsUrl('https://www.warcraftlogs.com/reports/abc123')
    expect(result.isValid).toBe(true)
    expect(result.reportId).toBe('abc123')
    expect(result.error).toBeUndefined()
  })

  it('validates URL without www', () => {
    const result = validateWarcraftLogsUrl('https://warcraftlogs.com/reports/def456')
    expect(result.isValid).toBe(true)
    expect(result.reportId).toBe('def456')
  })

  it('validates HTTP URL', () => {
    const result = validateWarcraftLogsUrl('http://www.warcraftlogs.com/reports/ghi789')
    expect(result.isValid).toBe(true)
    expect(result.reportId).toBe('ghi789')
  })

  it('rejects empty string', () => {
    const result = validateWarcraftLogsUrl('')
    expect(result.isValid).toBe(false)
    expect(result.error).toBe('URL is required')
  })

  it('rejects whitespace only', () => {
    const result = validateWarcraftLogsUrl('   ')
    expect(result.isValid).toBe(false)
    expect(result.error).toBe('URL is required')
  })

  it('rejects non-Warcraft Logs URL', () => {
    const result = validateWarcraftLogsUrl('https://google.com')
    expect(result.isValid).toBe(false)
    expect(result.error).toBe('URL must be from warcraftlogs.com/reports/')
  })

  it('rejects Warcraft Logs URL without reports path', () => {
    const result = validateWarcraftLogsUrl('https://www.warcraftlogs.com/')
    expect(result.isValid).toBe(false)
    expect(result.error).toBe('URL must be from warcraftlogs.com/reports/')
  })

  it('rejects URL with invalid report code', () => {
    const result = validateWarcraftLogsUrl('https://www.warcraftlogs.com/reports/')
    expect(result.isValid).toBe(false)
    expect(result.error).toBe('URL must contain a valid report code')
  })

  it('rejects report code that is too short', () => {
    const result = validateWarcraftLogsUrl('https://www.warcraftlogs.com/reports/abc')
    expect(result.isValid).toBe(false)
    expect(result.error).toBe('Report code appears too short')
  })

  it('rejects report code that is too long', () => {
    const result = validateWarcraftLogsUrl('https://www.warcraftlogs.com/reports/abcdefghijklmnopqrstuv')
    expect(result.isValid).toBe(false)
    expect(result.error).toBe('Report code appears too long')
  })

  it('handles URLs with extra path segments', () => {
    const result = validateWarcraftLogsUrl('https://www.warcraftlogs.com/reports/abc123/fight/1')
    expect(result.isValid).toBe(true)
    expect(result.reportId).toBe('abc123')
  })

  it('handles URLs with query parameters', () => {
    const result = validateWarcraftLogsUrl('https://www.warcraftlogs.com/reports/abc123?ref=raidbots')
    expect(result.isValid).toBe(true)
    expect(result.reportId).toBe('abc123')
  })

  it('extracts a preferred fight id from query parameters', () => {
    const result = validateWarcraftLogsUrl('https://www.warcraftlogs.com/reports/abc123?fight=4')
    expect(result.isValid).toBe(true)
    expect(result.reportId).toBe('abc123')
    expect(result.fightId).toBe(4)
  })

  it('extracts a preferred fight id from hash parameters', () => {
    const result = validateWarcraftLogsUrl('https://www.warcraftlogs.com/reports/abc123#fight=7')
    expect(result.isValid).toBe(true)
    expect(result.reportId).toBe('abc123')
    expect(result.fightId).toBe(7)
  })

  it('handles mixed case report codes', () => {
    const result = validateWarcraftLogsUrl('https://www.warcraftlogs.com/reports/AbC123')
    expect(result.isValid).toBe(true)
    expect(result.reportId).toBe('AbC123')
  })
})

describe('isValidWarcraftLogsUrl', () => {
  it('returns true for valid URLs', () => {
    expect(isValidWarcraftLogsUrl('https://www.warcraftlogs.com/reports/abc123')).toBe(true)
  })

  it('returns false for invalid URLs', () => {
    expect(isValidWarcraftLogsUrl('https://google.com')).toBe(false)
    expect(isValidWarcraftLogsUrl('')).toBe(false)
  })
})

describe('extractReportId', () => {
  it('extracts report ID from valid URLs', () => {
    expect(extractReportId('https://www.warcraftlogs.com/reports/abc123')).toBe('abc123')
  })

  it('returns undefined for invalid URLs', () => {
    expect(extractReportId('https://google.com')).toBeUndefined()
    expect(extractReportId('')).toBeUndefined()
  })
})

describe('extractFightId', () => {
  it('extracts fight id from query parameters', () => {
    expect(extractFightId('https://www.warcraftlogs.com/reports/abc123?fight=4')).toBe(4)
  })

  it('extracts fight id from hash parameters', () => {
    expect(extractFightId('https://www.warcraftlogs.com/reports/abc123#fight=8')).toBe(8)
  })

  it('returns undefined when no fight id is present', () => {
    expect(extractFightId('https://www.warcraftlogs.com/reports/abc123')).toBeUndefined()
  })
})
