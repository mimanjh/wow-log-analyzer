interface MetricCardProps {
    title: string;
    description: string;
    playerValue: number;
    cohortValue: number;
    delta: number;
    confidence: "high" | "medium" | "low";
    caution?: string;
    unit?: string;
}

export function MetricCard({
    title,
    description,
    playerValue,
    cohortValue,
    delta,
    confidence,
    caution,
    unit = "",
}: MetricCardProps) {
    const getConfidenceColor = (metricConfidence: string) => {
        switch (metricConfidence) {
            case "high":
                return "text-emerald-400";
            case "medium":
                return "text-amber-400";
            case "low":
                return "text-rose-400";
            default:
                return "text-slate-400";
        }
    };

    const formatValue = (value: number) => {
        if (typeof value === "number" && value % 1 !== 0) {
            return value.toFixed(1);
        }

        return value.toString();
    };

    return (
        <div className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
            <div className="flex items-start justify-between">
                <div>
                    <h3 className="text-lg font-semibold text-white">
                        {title}
                    </h3>
                    <p className="mt-1 text-sm text-slate-400">{description}</p>
                </div>
                <div
                    className={`text-sm font-medium ${getConfidenceColor(confidence)}`}
                >
                    {confidence.toUpperCase()}
                </div>
            </div>

            <div className="mt-4 grid grid-cols-3 gap-4">
                <div className="text-center">
                    <p className="text-sm text-slate-400">You</p>
                    <p className="mt-1 text-xl font-semibold text-white">
                        {formatValue(playerValue)}
                        {unit}
                    </p>
                </div>
                <div className="text-center">
                    <p className="text-sm text-slate-400">Elites</p>
                    <p className="mt-1 text-xl font-semibold text-slate-300">
                        {formatValue(cohortValue)}
                        {unit}
                    </p>
                </div>
                <div className="text-center">
                    <p className="text-sm text-slate-400">Delta</p>
                    <p
                        className={`mt-1 text-xl font-semibold ${
                            delta > 0
                                ? "text-emerald-400"
                                : delta < 0
                                  ? "text-rose-400"
                                  : "text-slate-400"
                        }`}
                    >
                        {delta > 0 ? "+" : ""}
                        {formatValue(delta)}
                        {unit}
                    </p>
                </div>
            </div>

            <div className="mt-4 flex items-center justify-end gap-4">
                {caution && (
                    <div className="flex items-center space-x-1">
                        <span className="text-xs text-amber-400">Caution:</span>
                        <span className="text-xs text-amber-400">
                            {caution}
                        </span>
                    </div>
                )}
            </div>
        </div>
    );
}
