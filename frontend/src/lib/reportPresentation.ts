export const reportStages = [
    { key: "player-data", label: "Fetch Player Data" },
    { key: "rankings", label: "Find Ranking Elites" },
    { key: "cohort", label: "Load Top Ranked Fights" },
    { key: "analyzing", label: "Run Deterministic Analysis" },
    { key: "insights", label: "Generate Insights" },
    { key: "completed", label: "Complete" },
] as const;

export type ComparisonFilter = "all" | "behind" | "ahead";

export function normalizeTrackedName(value: string) {
    return value.toLowerCase().replace(/[^a-z0-9]+/g, "");
}

export function sortByTrackedPriority<T extends { abilityName: string }>(
    values: T[],
    trackedNames: string[],
) {
    const priorityByName = new Map(
        trackedNames.map((name, index) => [normalizeTrackedName(name), index]),
    );

    return [...values].sort((left, right) => {
        const leftPriority = priorityByName.get(
            normalizeTrackedName(left.abilityName),
        );
        const rightPriority = priorityByName.get(
            normalizeTrackedName(right.abilityName),
        );

        if (leftPriority !== undefined && rightPriority !== undefined) {
            return leftPriority - rightPriority;
        }
        if (leftPriority !== undefined) {
            return -1;
        }
        if (rightPriority !== undefined) {
            return 1;
        }

        return left.abilityName.localeCompare(right.abilityName);
    });
}

export function formatKillTime(seconds: number) {
    return `${Math.floor(seconds / 60)}:${(seconds % 60)
        .toString()
        .padStart(2, "0")}`;
}

export function formatSigned(value: number, unit = "", digits = 1) {
    const prefix = value > 0 ? "+" : "";
    return `${prefix}${value.toFixed(digits)}${unit}`;
}

export function confidenceColor(confidence: string) {
    switch (confidence) {
        case "high":
            return "text-emerald-400";
        case "medium":
            return "text-amber-400";
        default:
            return "text-rose-400";
    }
}

export function matchesFilter(delta: number, filter: ComparisonFilter) {
    if (filter === "behind") {
        return delta < 0;
    }
    if (filter === "ahead") {
        return delta > 0;
    }
    return true;
}

export function getStageLabel(stage: string) {
    return reportStages.find((entry) => entry.key === stage)?.label ?? stage;
}

export function getStageCompletionState(
    stage: string,
    status: string,
    key: string,
) {
    const stageIndex = reportStages.findIndex((entry) => entry.key === stage);
    const keyIndex = reportStages.findIndex((entry) => entry.key === key);

    if (status === "completed") {
        return "complete";
    }
    if (status === "failed") {
        if (keyIndex < stageIndex) {
            return "complete";
        }
        if (key === stage) {
            return "failed";
        }
        return "pending";
    }
    if (keyIndex < stageIndex) {
        return "complete";
    }
    if (key === stage) {
        return "active";
    }
    return "pending";
}

export function mapRegionToBlizzardSlug(region?: string) {
    const normalized = region?.trim().toLowerCase() ?? "";
    switch (normalized) {
        case "united states":
        case "us":
            return "us";
        case "europe":
        case "eu":
            return "eu";
        case "korea":
        case "kr":
            return "kr";
        case "taiwan":
        case "tw":
            return "tw";
        case "china":
        case "cn":
            return "cn";
        default:
            return "";
    }
}

export function slugifyServerName(serverName?: string) {
    return (serverName ?? "")
        .trim()
        .toLowerCase()
        .replace(/['’]/g, "")
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "");
}

export function formatTimelineTimestamp(durationMs: number) {
    const totalSeconds = Math.max(0, Math.round(durationMs / 1000));
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}
