import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import { Button } from "../../components/ui/Button";
import { validateWarcraftLogsUrl } from "../../lib/validateUrl";

interface UrlIntakeFormProps {
    onSubmit: (url: string) => void | Promise<void>;
    isSubmitting?: boolean;
    submitError?: string | null;
    initialUrl?: string;
}

export function UrlIntakeForm({
    onSubmit,
    isSubmitting = false,
    submitError = null,
    initialUrl = "",
}: UrlIntakeFormProps) {
    const [url, setUrl] = useState(initialUrl);
    const [validationError, setValidationError] = useState("");

    useEffect(() => {
        setUrl(initialUrl);
    }, [initialUrl]);

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        const validation = validateWarcraftLogsUrl(url);
        if (!validation.isValid) {
            setValidationError(validation.error || "Invalid URL");
            return;
        }

        setValidationError("");
        await onSubmit(url.trim());
    }

    const errorMessage = validationError || submitError || "";

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
                        if (validationError) setValidationError("");
                    }}
                    disabled={isSubmitting}
                    placeholder="https://www.warcraftlogs.com/reports/abc123"
                    className="w-full rounded-2xl border border-slate-700 bg-slate-950/90 px-4 py-3 text-sm text-slate-100 outline-none transition focus:border-sky-400 focus:ring-2 focus:ring-sky-500/20"
                />
                {errorMessage ? (
                    <p className="mt-2 text-sm text-rose-400">{errorMessage}</p>
                ) : (
                    <p className="mt-2 text-xs text-slate-400">
                        Example: https://www.warcraftlogs.com/reports/abc123
                    </p>
                )}
            </div>

            <div className="flex justify-end">
                <Button type="submit" disabled={isSubmitting}>
                    {isSubmitting ? "Loading report..." : "Start analysis"}
                </Button>
            </div>
        </form>
    );
}
