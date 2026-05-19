export interface AnalyzeState {
  reportUrl: string
  reportId: string | null
  reportJob: ReportJob | null
  reportResult: ReportResult | null
  error: string | null
  setReportUrl: (reportUrl: string) => void
  setReportJob: (job: ReportJob | null) => void
  setReportResult: (result: ReportResult | null) => void
  setError: (error: string | null) => void
  reset: () => void
}

export interface CharacterReportsCacheEntry {
  reports: CharacterReportSummary[]
  cachedAt: number
  nextCursor: string | null
  hasMoreReports: boolean
}

export interface BrowserState {
  auth: AuthStatus | null
  authCachedAt: number | null
  characters: BrowserCharacter[]
  selectedCharacter: BrowserCharacter | null
  reportCacheByCharacter: Record<number, CharacterReportsCacheEntry>
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

export interface CharacterFightSummary {
  id: number
  name: string
  difficulty: string
  kill: boolean
  killTime: number
  encounterId: number
}

export interface CharacterReportSummary {
  code: string
  title: string
  zoneName?: string
  bossNames?: string[]
  startTime: string
  endTime: string
  fights?: CharacterFightSummary[]
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
  friendlyPlayers?: FightParticipant[]
}

export interface FightParticipant {
  id: number
  name: string
  serverName?: string
  class?: string
}

export interface Character {
  id: number
  name: string
  class: string
  spec: string
  role: string
  serverName?: string
  talentImportCode?: string
  talentCalculatorUrl?: string
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
  warnings?: ReportWarning[]
  ai: AISection
}

export interface ReportWarning {
  kind: string
  title: string
  message: string
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

export interface ResourceTimelineResponse {
  resourceTypeId: number
  resourceType: string
  fightDurationMs: number
  player: ResourceTimelineSeries
  elite: ResourceTimelineSeries[]
}

export interface BuffTimelineResponse {
  abilityId: number
  abilityName: string
  fightDurationMs: number
  player: BuffTimelineSeries
  elite: BuffTimelineSeries[]
}

export interface AbilityTimelineSeries {
  label: string
  subtitle?: string
  reportUrl?: string
  castsMs: number[]
}

export interface ResourceTimelineSeries {
  label: string
  subtitle?: string
  reportUrl?: string
  durationMs: number
  samples: ResourceTimelineSample[]
  wasteMarkersMs?: number[]
}

export interface ResourceTimelineSample {
  timestampMs: number
  value: number
  maxValue?: number
  waste?: number
}

export interface BuffTimelineSeries {
  label: string
  subtitle?: string
  reportUrl?: string
  windows: BuffTimelineWindow[]
}

export interface BuffTimelineWindow {
  startMs: number
  endMs: number
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
  caution?: string
}

export interface MajorCDCountMetric {
  value: number
  totalCooldowns: number
  caution?: string
}

export interface MajorCDDriftMetric {
  value: number
  totalDrift: number
  cooldownCount: number
  caution?: string
}

export interface BuffUptimeMetric {
  value: number
  totalUptime: number
  fightDuration: number
  caution?: string
}

export interface DowntimePercentageMetric {
  value: number
  totalDowntime: number
  fightDuration: number
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
  caution?: string
}

export interface ResourceUsageComparison {
  resourceTypeId: number
  resourceType: string
  playerSampleCount: number
  cohortMedianSampleCount: number
  sampleCountDelta: number
  playerFullMarkerCount: number
  cohortMedianFullMarkerCount: number
  fullMarkerDelta: number
  playerFullWindowSeconds: number
  cohortMedianFullWindowSeconds: number
  fullWindowDeltaSeconds: number
  playerAveragePct: number
  cohortMedianAveragePct: number
  averagePctDelta: number
  playerTimeAtMaxSeconds: number
  cohortMedianTimeAtMaxSeconds: number
  timeAtMaxDeltaSeconds: number
  playerSpent: number
  cohortMedianSpent: number
  spentDelta: number
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
  caution?: string
}
