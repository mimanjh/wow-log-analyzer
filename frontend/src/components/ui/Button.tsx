import type { ButtonHTMLAttributes } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: "primary" | "secondary";
}

export function Button({
    variant = "primary",
    className = "",
    ...rest
}: ButtonProps) {
    const baseStyles =
        "inline-flex cursor-pointer items-center justify-center rounded-lg px-4 py-2 text-sm font-semibold transition duration-150 focus:outline-none focus:ring-2 focus:ring-sky-500 active:translate-y-px active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50";
    const variantStyles =
        variant === "secondary"
            ? "bg-slate-800 text-slate-100 hover:bg-slate-700"
            : "bg-sky-500 text-slate-950 hover:bg-sky-400";

    return (
        <button
            className={`${baseStyles} ${variantStyles} ${className}`}
            {...rest}
        />
    );
}
