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

export interface Fight {
  id: number
  name: string
  difficulty: string
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
}

export interface ComparisonResult {
  playerMetrics: PlayerFightMetrics
  cohortStats: CohortStatistics
  deltas: MetricDeltas
  rankings: MetricRankings
}

export interface ReportResult {
  fight: Fight
  character: Character
  comparison: ComparisonResult
  ai: AISection
}

export interface ReportJob {
  jobId: string
  status: "queued" | "running" | "completed" | "failed"
  stage: string
  message: string
  error?: string
  result?: ReportResult
  createdAt: string
  updatedAt: string
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
  percentile: number
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
