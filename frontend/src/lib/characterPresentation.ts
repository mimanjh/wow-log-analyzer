export function getCharacterCardClasses(
    characterClass: string,
    selected = false,
) {
    const palette: Record<string, string> = {
        "Death Knight": selected
            ? "border-red-500 bg-red-950/20"
            : "border-red-700/60 bg-slate-950/80 hover:border-red-500/70",
        "Demon Hunter": selected
            ? "border-violet-500 bg-violet-950/20"
            : "border-violet-700/60 bg-slate-950/80 hover:border-violet-500/70",
        Druid: selected
            ? "border-orange-500 bg-orange-950/20"
            : "border-orange-700/60 bg-slate-950/80 hover:border-orange-500/70",
        Evoker: selected
            ? "border-emerald-500 bg-emerald-950/20"
            : "border-emerald-700/60 bg-slate-950/80 hover:border-emerald-500/70",
        Hunter: selected
            ? "border-lime-500 bg-lime-950/20"
            : "border-lime-700/60 bg-slate-950/80 hover:border-lime-500/70",
        Mage: selected
            ? "border-sky-400 bg-sky-950/20"
            : "border-sky-700/60 bg-slate-950/80 hover:border-sky-400/70",
        Monk: selected
            ? "border-teal-500 bg-teal-950/20"
            : "border-teal-700/60 bg-slate-950/80 hover:border-teal-500/70",
        Paladin: selected
            ? "border-pink-400 bg-pink-950/20"
            : "border-pink-700/60 bg-slate-950/80 hover:border-pink-400/70",
        Priest: selected
            ? "border-stone-300 bg-stone-950/20"
            : "border-stone-600/70 bg-slate-950/80 hover:border-stone-300/70",
        Rogue: selected
            ? "border-amber-400 bg-amber-950/20"
            : "border-amber-700/60 bg-slate-950/80 hover:border-amber-400/70",
        Shaman: selected
            ? "border-blue-500 bg-blue-950/20"
            : "border-blue-700/60 bg-slate-950/80 hover:border-blue-500/70",
        Warlock: selected
            ? "border-fuchsia-500 bg-fuchsia-950/20"
            : "border-fuchsia-700/60 bg-slate-950/80 hover:border-fuchsia-500/70",
        Warrior: selected
            ? "border-yellow-700 bg-yellow-950/20"
            : "border-yellow-800/80 bg-slate-950/80 hover:border-yellow-700/80",
    };

    return (
        palette[characterClass] ??
        (selected
            ? "border-sky-500 bg-sky-950/20"
            : "border-slate-700 bg-slate-950/80 hover:border-slate-500")
    );
}
