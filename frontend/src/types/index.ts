export interface AnalyzeState {
  reportUrl: string
  setReportUrl: (reportUrl: string) => void
  reportId: string | null
  preferredFightId: number | null
  fights: Fight[]
  characters: Character[]
  charactersFightId: number | null
  selectedFight: Fight | null
  selectedCharacter: Character | null
  reportJob: ReportJob | null
  reportResult: ReportResult | null
  isLoading: boolean
  error: string | null
  setReportData: (data: { reportId: string | null; preferredFightId?: number | null; fights: Fight[]; characters: Character[] }) => void
  setCharactersForFight: (fightId: number, characters: Character[]) => void
  setSelectedFight: (fight: Fight | null) => void
  setSelectedCharacter: (character: Character | null) => void
  setReportJob: (job: ReportJob | null) => void
  setReportResult: (result: ReportResult | null) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  reset: () => void
}

export interface BrowserState {
  auth: AuthStatus | null
  characters: BrowserCharacter[]
  selectedCharacter: BrowserCharacter | null
  reports: CharacterReportSummary[]
  reportsCachedAt: number | null
  nextCursor: string | null
  hasMoreReports: boolean
  isAuthLoading: boolean
  isCharactersLoading: boolean
  isReportsLoading: boolean
  error: string | null
  setAuth: (auth: AuthStatus | null) => void
  setCharacters: (characters: BrowserCharacter[]) => void
  finishCharactersLoad: (characters: BrowserCharacter[]) => void
  setSelectedCharacter: (character: BrowserCharacter | null) => void
  resetReports: () => void
  appendReports: (page: CharacterReportsPage) => void
  setLoadingState: (key: "isAuthLoading" | "isCharactersLoading" | "isReportsLoading", value: boolean) => void
  setError: (error: string | null) => void
  reset: () => void
}

export interface AuthStatus {
  authenticated: boolean
  user?: AuthUser
}

export interface AuthUser {
  id: number
  name: string
  avatar?: string
  battleTag?: string
}

export interface BrowserCharacter {
  id: number
  name: string
  class: string
  serverName: string
  serverRegion: string
  serverSlug?: string
}

export interface CharacterReportSummary {
  code: string
  title: string
  zoneName?: string
  bossNames?: string[]
  startTime: string
  endTime: string
}

export interface CharacterReportsPage {
  reports: CharacterReportSummary[]
  nextCursor: string | null
  hasMore: boolean
}

export interface Fight {
  id: number
  name: string
  difficulty: string
  kill: boolean
  killTime: number
  encounterId: number
  startTime: string
  endTime: string
  bossPercent?: number
}

export interface Character {
  id: number
  name: string
  class: string
  spec: string
  role: string
  serverName?: string
}

export interface ComparisonResult {
  playerMetrics: PlayerFightMetrics
  cohortStats: CohortStatistics
  deltas: MetricDeltas
  rankings: MetricRankings
  abilityUsage: AbilityUsageComparison[]
  buffUptimes: BuffUptimeComparison[]
  resourceUsage: ResourceUsageComparison[]
}

export interface ReportResult {
  fight: Fight
  character: Character
  cohort: CohortEntry[]
  comparison: ComparisonResult
  ai: AISection
}

export interface CohortEntry {
  name: string
  class: string
  spec: string
  server?: string
  serverRegion?: string
  reportId: string
  fightId: number
  rankValue: number
  durationMs: number
  reportUrl: string
}

export interface ReportJob {
  jobId: string
  status: "queued" | "running" | "completed" | "failed"
  stage: string
  message: string
  fight: Fight
  character: Character
  progress: ReportJobProgress
  error?: string
  result?: ReportResult
  createdAt: string
  updatedAt: string
}

export interface AbilityTimelineResponse {
  abilityId: number
  abilityName: string
  fightDurationMs: number
  player: AbilityTimelineSeries
  elite: AbilityTimelineSeries[]
}

export interface AbilityTimelineSeries {
  label: string
  subtitle?: string
  reportUrl?: string
  castsMs: number[]
}

export interface ReportJobProgress {
  current: number
  total: number
}

export interface AISection {
  available: boolean
  fallbackUsed: boolean
  model?: string
  warning?: string
  insights: AIInsight[]
  focusRecommendation?: FocusRecommendation
}

export interface AIInsight {
  metricKey: string
  title: string
  summary: string
  confidence: string
  caution?: string
}

export interface FocusRecommendation {
  metricKey: string
  title: string
  recommendation: string
  reasoning: string
}

export interface PlayerFightMetrics {
  playerId: number
  fightId: number
  fightStart: string
  fightEnd: string
  duration: number // in milliseconds
  castsPerMin: CastsPerMinuteMetric
  majorCdCount: MajorCDCountMetric
  majorCdDrift: MajorCDDriftMetric
  buffUptime: BuffUptimeMetric
  downtimePct: DowntimePercentageMetric
}

export interface CohortStatistics {
  sampleSize: number
  castsPerMin: CohortMetricStats
  majorCdCount: CohortMetricStats
  majorCdDrift: CohortMetricStats
  buffUptime: CohortMetricStats
  downtimePct: CohortMetricStats
}

export interface CohortMetricStats {
  mean: number
  median: number
  stdDev: number
  min: number
  max: number
  p25: number
  p75: number
  p95: number
}

export interface MetricDeltas {
  castsPerMin: MetricDelta
  majorCdCount: MetricDelta
  majorCdDrift: MetricDelta
  buffUptime: MetricDelta
  downtimePct: MetricDelta
}

export interface MetricDelta {
  playerValue: number
  cohortValue: number
  difference: number
  confidence: string
  caution?: string
}

export interface MetricRankings {
  castsPerMin: number
  majorCdCount: number
  majorCdDrift: number
  buffUptime: number
  downtimePct: number
}

export interface CastsPerMinuteMetric {
  value: number
  totalCasts: number
  fightDuration: number
  confidence: string
  caution?: string
}

export interface MajorCDCountMetric {
  value: number
  totalCooldowns: number
  confidence: string
  caution?: string
}

export interface MajorCDDriftMetric {
  value: number
  totalDrift: number
  cooldownCount: number
  confidence: string
  caution?: string
}

export interface BuffUptimeMetric {
  value: number
  totalUptime: number
  fightDuration: number
  confidence: string
  caution?: string
}

export interface DowntimePercentageMetric {
  value: number
  totalDowntime: number
  fightDuration: number
  confidence: string
  caution?: string
}

export interface AbilityUsageComparison {
  abilityId: number
  abilityName: string
  playerCount: number
  playerCastsPerMinute: number
  playerFirstUseSeconds?: number
  cohortMedianCount: number
  cohortMedianPerMinute: number
  cohortMedianFirstUseSeconds?: number
  countDelta: number
  perMinuteDelta: number
  firstUseDeltaSeconds?: number
  sampleSize: number
  confidence: string
  caution?: string
}

export interface BuffUptimeComparison {
  abilityId: number
  abilityName: string
  playerUptimePct: number
  playerFirstApplySeconds?: number
  cohortMedianUptimePct: number
  cohortMedianFirstApplySeconds?: number
  uptimeDelta: number
  firstApplyDeltaSeconds?: number
  sampleSize: number
  confidence: string
  caution?: string
}

export interface ResourceUsageComparison {
  resourceTypeId: number
  resourceType: string
  playerGeneratedPerMinute: number
  cohortMedianGeneratedPerMinute: number
  generatedDelta: number
  playerWastePerMinute: number
  cohortMedianWastePerMinute: number
  wasteDelta: number
  playerWastePct: number
  cohortMedianWastePct: number
  wastePctDelta: number
  sampleSize: number
  confidence: string
  caution?: string
}
