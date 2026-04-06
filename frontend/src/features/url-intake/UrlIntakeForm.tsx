import { useState } from "react";
import type { FormEvent } from "react";
import { Button } from "../../components/ui/Button";
import { validateWarcraftLogsUrl } from "../../lib/validateUrl";

interface UrlIntakeFormProps {
    onSubmit: (url: string) => void;
}

export function UrlIntakeForm({ onSubmit }: UrlIntakeFormProps) {
    const [url, setUrl] = useState("");
    const [error, setError] = useState("");

    function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        const validation = validateWarcraftLogsUrl(url);
        if (!validation.isValid) {
            setError(validation.error || "Invalid URL");
            return;
        }

        setError("");
        onSubmit(url.trim());
    }

    return (
        <form className="space-y-6" onSubmit={handleSubmit}>
            <div>
                <label
                    htmlFor="reportUrl"
                    className="mb-2 block text-sm font-medium text-slate-200"
                >
                    Warcraft Logs report URL
                </label>
                <input
                    id="reportUrl"
                    type="url"
                    value={url}
                    onChange={(event) => {
                        setUrl(event.target.value);
                        if (error) setError(""); // Clear error on change
                    }}
                    placeholder="https://www.warcraftlogs.com/reports/abc123"
                    className="w-full rounded-2xl border border-slate-700 bg-slate-950/90 px-4 py-3 text-sm text-slate-100 outline-none transition focus:border-sky-400 focus:ring-2 focus:ring-sky-500/20"
                />
                {error ? (
                    <p className="mt-2 text-sm text-rose-400">{error}</p>
                ) : (
                    <p className="mt-2 text-xs text-slate-400">
                        Example: https://www.warcraftlogs.com/reports/abc123
                    </p>
                )}
            </div>

            <div className="flex justify-end">
                <Button type="submit">Start analysis</Button>
            </div>
        </form>
    );
}
