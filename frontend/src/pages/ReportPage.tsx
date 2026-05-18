import { type ReactNode, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useAnalyzeStore } from "../stores/useAnalyzeStore";
import { useBrowserStore } from "../stores/useBrowserStore";
import { usePageTitle } from "../hooks/usePageTitle";
import { Button } from "../components/ui/Button";
import {
    getAbilityTimeline,
    getReportJob,
    getResourceTimeline,
} from "../lib/api";
import {
    formatKillTime,
    formatSigned,
    getStageCompletionState,
    getStageLabel,
    matchesFilter,
    sortByTrackedPriority,
    type ComparisonFilter,
} from "../lib/reportPresentation";
import type {
    AbilityTimelineResponse,
    AbilityTimelineSeries,
    Character,
    Fight,
    ReportJob,
    ResourceTimelineResponse,
    ResourceTimelineSeries,
} from "../types";

const reportStages = [
    { key: "player-data", label: "Fetch Player Data" },
    { key: "rankings", label: "Find Ranking Elites" },
    { key: "cohort", label: "Load Top Ranked Fights" },
    { key: "analyzing", label: "Run Deterministic Analysis" },
    { key: "insights", label: "Generate Insights" },
    { key: "completed", label: "Complete" },
] as const;

const trackedSpecPriorities: Record<
    string,
    { abilities: string[]; buffs: string[] }
> = {
    "Blood Death Knight": {
        abilities: [
            "Dancing Rune Weapon",
            "Vampiric Blood",
            "Anti-Magic Zone",
            "Icebound Fortitude",
            "Lichborne",
            "Anti-Magic Shell",
        ],
        buffs: [
            "Dancing Rune Weapon",
            "Vampiric Blood",
            "Anti-Magic Zone",
            "Icebound Fortitude",
            "Lichborne",
            "Anti-Magic Shell",
        ],
    },
    "Vengeance Demon Hunter": {
        abilities: [
            "Fiery Brand",
            "Immolation Aura",
            "Fel Devastation",
            "Sigil of Flame",
            "Darkness",
            "Demon Spikes",
        ],
        buffs: [
            "Fiery Brand",
            "Immolation Aura",
            "Fel Devastation",
            "Sigil of Flame",
            "Darkness",
            "Demon Spikes",
        ],
    },
    "Guardian Druid": {
        abilities: [
            "Incarnation: Guardian of Ursoc",
            "Convoke the Spirits",
            "Lunar Beam",
            "Barkskin",
            "Survival Instincts",
            "Frenzied Regeneration",
        ],
        buffs: [
            "Incarnation: Guardian of Ursoc",
            "Convoke the Spirits",
            "Lunar Beam",
            "Barkskin",
            "Survival Instincts",
            "Frenzied Regeneration",
        ],
    },
    "Brewmaster Monk": {
        abilities: [
            "Invoke Niuzao, the Black Ox",
            "Exploding Keg",
            "Celestial Infusion",
            "Touch of Death",
            "Fortifying Brew",
        ],
        buffs: [
            "Invoke Niuzao, the Black Ox",
            "Exploding Keg",
            "Celestial Infusion",
            "Touch of Death",
            "Fortifying Brew",
        ],
    },
    "Protection Paladin": {
        abilities: [
            "Avenging Wrath",
            "Divine Toll",
            "Ardent Defender",
            "Guardian of Ancient Kings",
            "Divine Shield",
        ],
        buffs: [
            "Avenging Wrath",
            "Divine Toll",
            "Ardent Defender",
            "Guardian of Ancient Kings",
            "Divine Shield",
        ],
    },
    "Protection Warrior": {
        abilities: [
            "Avatar",
            "Ravager",
            "Thunder Blast",
            "Shield Charge",
            "Rend",
            "Shield Wall",
            "Demoralizing Shout",
            "Spell Reflection",
            "Rallying Cry",
        ],
        buffs: [
            "Avatar",
            "Ravager",
            "Thunder Blast",
            "Shield Charge",
            "Rend",
            "Shield Wall",
            "Demoralizing Shout",
            "Spell Reflection",
            "Rallying Cry",
        ],
    },
    "Frost Death Knight": {
        abilities: [
            "Pillar of Frost",
            "Frostwyrm's Fury",
            "Breath of Sindragosa",
            "Icebound Fortitude",
            "Lichborne",
            "Anti-Magic Shell",
        ],
        buffs: [
            "Pillar of Frost",
            "Frostwyrm's Fury",
            "Breath of Sindragosa",
            "Icebound Fortitude",
            "Lichborne",
            "Anti-Magic Shell",
        ],
    },
    "Unholy Death Knight": {
        abilities: [
            "Army of the Dead",
            "Raise Abomination",
            "Dark Transformation",
            "Putrefy",
            "Soul Reaper",
            "Icebound Fortitude",
            "Lichborne",
            "Anti-Magic Shell",
        ],
        buffs: [
            "Army of the Dead",
            "Raise Abomination",
            "Dark Transformation",
            "Putrefy",
            "Soul Reaper",
            "Icebound Fortitude",
            "Lichborne",
            "Anti-Magic Shell",
        ],
    },
    "Havoc Demon Hunter": {
        abilities: [
            "Metamorphosis",
            "Eye Beam",
            "Essence Break",
            "The Hunt",
            "Darkness",
        ],
        buffs: [
            "Metamorphosis",
            "Eye Beam",
            "Essence Break",
            "The Hunt",
            "Darkness",
        ],
    },
    "Feral Druid": {
        abilities: [
            "Berserk",
            "Tiger's Fury",
            "Feral Frenzy",
            "Convoke the Spirits",
            "Barkskin",
            "Survival Instincts",
            "Frenzied Regeneration",
        ],
        buffs: [
            "Berserk",
            "Tiger's Fury",
            "Feral Frenzy",
            "Convoke the Spirits",
            "Barkskin",
            "Survival Instincts",
            "Frenzied Regeneration",
        ],
    },
    "Survival Hunter": {
        abilities: [
            "Aspect of the Eagle",
            "Takedown",
            "Boomstick",
            "Exhilaration",
            "Survival of the Fittest",
            "Aspect of the Turtle",
        ],
        buffs: [
            "Aspect of the Eagle",
            "Takedown",
            "Boomstick",
            "Exhilaration",
            "Survival of the Fittest",
            "Aspect of the Turtle",
        ],
    },
    "Windwalker Monk": {
        abilities: [
            "Zenith",
            "Touch of Death",
            "Touch of Karma",
            "Fortifying Brew",
        ],
        buffs: [
            "Zenith",
            "Touch of Death",
            "Touch of Karma",
            "Fortifying Brew",
        ],
    },
    "Retribution Paladin": {
        abilities: [
            "Avenging Wrath",
            "Divine Toll",
            "Wake of Ashes",
            "Execution Sentence",
            "Divine Protection",
            "Divine Shield",
        ],
        buffs: [
            "Avenging Wrath",
            "Divine Toll",
            "Wake of Ashes",
            "Execution Sentence",
            "Divine Protection",
            "Divine Shield",
        ],
    },
    "Assassination Rogue": {
        abilities: [
            "Kingsbane",
            "Shiv",
            "Deathmark",
            "Vanish",
            "Feint",
            "Evasion",
            "Cloak of Shadows",
            "Crimson Vial",
        ],
        buffs: [
            "Kingsbane",
            "Shiv",
            "Deathmark",
            "Vanish",
            "Feint",
            "Evasion",
            "Cloak of Shadows",
            "Crimson Vial",
        ],
    },
    "Outlaw Rogue": {
        abilities: [
            "Adrenaline Rush",
            "Keep It Rolling",
            "Vanish",
            "Feint",
            "Evasion",
            "Cloak of Shadows",
            "Crimson Vial",
        ],
        buffs: [
            "Adrenaline Rush",
            "Keep It Rolling",
            "Vanish",
            "Feint",
            "Evasion",
            "Cloak of Shadows",
            "Crimson Vial",
        ],
    },
    "Subtlety Rogue": {
        abilities: [
            "Shadow Blades",
            "Shadow Dance",
            "Secret Technique",
            "Vanish",
            "Feint",
            "Evasion",
            "Cloak of Shadows",
            "Crimson Vial",
        ],
        buffs: [
            "Shadow Blades",
            "Shadow Dance",
            "Secret Technique",
            "Vanish",
            "Feint",
            "Evasion",
            "Cloak of Shadows",
            "Crimson Vial",
        ],
    },
    "Enhancement Shaman": {
        abilities: [
            "Sundering",
            "Doom Winds",
            "Primordial Storm",
            "Astral Shift",
        ],
        buffs: ["Sundering", "Doom Winds", "Primordial Storm", "Astral Shift"],
    },
    "Arms Warrior": {
        abilities: [
            "Colossus Smash",
            "Ravager",
            "Avatar",
            "Bladestorm",
            "Sweeping Strikes",
            "Execute",
            "Rend",
            "Spell Reflection",
            "Rallying Cry",
        ],
        buffs: [
            "Colossus Smash",
            "Ravager",
            "Avatar",
            "Bladestorm",
            "Sweeping Strikes",
            "Execute",
            "Rend",
            "Spell Reflection",
            "Rallying Cry",
        ],
    },
    "Fury Warrior": {
        abilities: [
            "Recklessness",
            "Avatar",
            "Odyn's Fury",
            "Bladestorm",
            "Rend",
            "Enraged Regeneration",
            "Spell Reflection",
            "Rallying Cry",
        ],
        buffs: [
            "Recklessness",
            "Avatar",
            "Odyn's Fury",
            "Bladestorm",
            "Rend",
            "Enraged Regeneration",
            "Spell Reflection",
            "Rallying Cry",
        ],
    },
    "Devourer Demon Hunter": {
        abilities: ["Void Metamorphosis", "Collapsing Star", "Darkness"],
        buffs: ["Void Metamorphosis", "Collapsing Star", "Darkness"],
    },
    "Balance Druid": {
        abilities: [
            "Celestial Alignment",
            "Force of Nature",
            "Fury of Elune",
            "Full Moon",
            "Convoke the Spirits",
            "Lunar Eclipse",
            "Solar Eclipse",
            "Barkskin",
            "Survival Instincts",
            "Frenzied Regeneration",
        ],
        buffs: [
            "Celestial Alignment",
            "Force of Nature",
            "Fury of Elune",
            "Full Moon",
            "Convoke the Spirits",
            "Lunar Eclipse",
            "Solar Eclipse",
            "Barkskin",
            "Survival Instincts",
            "Frenzied Regeneration",
        ],
    },
    "Augmentation Evoker": {
        abilities: [
            "Ebon Might",
            "Breath of Eons",
            "Time Skip",
            "Spatial Paradox",
            "Time Spiral",
            "Zephyr",
            "Tip the Scales",
            "Obsidian Scales",
        ],
        buffs: [
            "Ebon Might",
            "Breath of Eons",
            "Time Skip",
            "Spatial Paradox",
            "Time Spiral",
            "Zephyr",
            "Tip the Scales",
            "Obsidian Scales",
        ],
    },
    "Devastation Evoker": {
        abilities: [
            "Dragonrage",
            "Deep Breath",
            "Spatial Paradox",
            "Time Spiral",
            "Zephyr",
            "Tip the Scales",
            "Obsidian Scales",
        ],
        buffs: [
            "Dragonrage",
            "Deep Breath",
            "Spatial Paradox",
            "Time Spiral",
            "Zephyr",
            "Tip the Scales",
            "Obsidian Scales",
        ],
    },
    "Beast Mastery Hunter": {
        abilities: [
            "Bestial Wrath",
            "Barbed Shot",
            "Exhilaration",
            "Survival of the Fittest",
            "Aspect of the Turtle",
        ],
        buffs: [
            "Bestial Wrath",
            "Barbed Shot",
            "Exhilaration",
            "Survival of the Fittest",
            "Aspect of the Turtle",
        ],
    },
    "Marksmanship Hunter": {
        abilities: [
            "Trueshot",
            "Volley",
            "Exhilaration",
            "Survival of the Fittest",
            "Aspect of the Turtle",
        ],
        buffs: [
            "Trueshot",
            "Volley",
            "Exhilaration",
            "Survival of the Fittest",
            "Aspect of the Turtle",
        ],
    },
    "Arcane Mage": {
        abilities: [
            "Arcane Surge",
            "Touch of the Magi",
            "Alter Time",
            "Mirror Image",
            "Ice Cold",
        ],
        buffs: [
            "Arcane Surge",
            "Touch of the Magi",
            "Alter Time",
            "Mirror Image",
            "Ice Cold",
        ],
    },
    "Fire Mage": {
        abilities: [
            "Combustion",
            "Pyroblast",
            "Fire Blast",
            "Flamestrike",
            "Meteor",
            "Scorch",
            "Alter Time",
            "Mirror Image",
            "Ice Cold",
        ],
        buffs: [
            "Combustion",
            "Pyroblast",
            "Fire Blast",
            "Flamestrike",
            "Meteor",
            "Scorch",
            "Alter Time",
            "Mirror Image",
            "Ice Cold",
        ],
    },
    "Frost Mage": {
        abilities: [
            "Ray of Frost",
            "Frozen Orb",
            "Alter Time",
            "Mirror Image",
            "Ice Cold",
        ],
        buffs: [
            "Ray of Frost",
            "Frozen Orb",
            "Alter Time",
            "Mirror Image",
            "Ice Cold",
        ],
    },
    "Shadow Priest": {
        abilities: [
            "Voidform",
            "Halo",
            "Dispersion",
            "Vampiric Embrace",
            "Desperate Prayer",
            "Fade",
        ],
        buffs: [
            "Voidform",
            "Halo",
            "Dispersion",
            "Vampiric Embrace",
            "Desperate Prayer",
            "Fade",
        ],
    },
    "Elemental Shaman": {
        abilities: [
            "Ascendance",
            "Stormkeeper",
            "Astral Shift",
            "Spiritwalker's Grace",
        ],
        buffs: [
            "Ascendance",
            "Stormkeeper",
            "Astral Shift",
            "Spiritwalker's Grace",
        ],
    },
    "Affliction Warlock": {
        abilities: [
            "Summon Darkglare",
            "Dark Harvest",
            "Unending Resolve",
            "Dark Pact",
        ],
        buffs: [
            "Summon Darkglare",
            "Dark Harvest",
            "Unending Resolve",
            "Dark Pact",
        ],
    },
    "Demonology Warlock": {
        abilities: [
            "Summon Demonic Tyrant",
            "Grimoire: Imp Lord",
            "Grimoire: Fel Ravager",
            "Unending Resolve",
            "Dark Pact",
        ],
        buffs: [
            "Summon Demonic Tyrant",
            "Grimoire: Imp Lord",
            "Grimoire: Fel Ravager",
            "Unending Resolve",
            "Dark Pact",
        ],
    },
    "Destruction Warlock": {
        abilities: [
            "Summon Infernal",
            "Havoc",
            "Unending Resolve",
            "Dark Pact",
        ],
        buffs: ["Summon Infernal", "Havoc", "Unending Resolve", "Dark Pact"],
    },
};

function renderSummaryCard(
    title: string,
    content: Array<{
        label: string;
        value: ReactNode;
        href?: string;
        fullWidth?: boolean;
    }>,
    href?: string,
) {
    const card = (
        <div
            className={`rounded-3xl border border-slate-800 bg-slate-950/80 p-6 transition hover:border-sky-500/60 ${href ? "cursor-pointer" : ""}`}
        >
            <h2 className="text-lg font-semibold text-white">{title}</h2>
            <div className="mt-4 grid gap-3 sm:grid-cols-2">
                {content.map((item) => (
                    <div
                        key={item.label}
                        className={item.fullWidth ? "sm:col-span-2" : ""}
                    >
                        <p className="text-sm text-slate-400">{item.label}</p>
                        {item.href ? (
                            <a
                                href={item.href}
                                target="_blank"
                                rel="noreferrer"
                                className="text-sky-300 underline-offset-4 hover:text-sky-200 hover:underline"
                            >
                                {item.value}
                            </a>
                        ) : (
                            <p className="text-white">{item.value}</p>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );

    if (!href) {
        return card;
    }

    return (
        <a href={href} target="_blank" rel="noreferrer" className="block">
            {card}
        </a>
    );
}

function mapRegionToBlizzardSlug(region?: string) {
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

function slugifyServerName(serverName?: string) {
    return (serverName ?? "")
        .trim()
        .toLowerCase()
        .replace(/['’]/g, "")
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "");
}

function formatTimelineTimestamp(durationMs: number) {
    const totalSeconds = Math.max(0, Math.round(durationMs / 1000));
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

function formatSeconds(value: number) {
    const rounded = Math.max(0, Math.round(value));
    const minutes = Math.floor(rounded / 60);
    const seconds = rounded % 60;
    if (minutes === 0) {
        return `${seconds}s`;
    }
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

function formatSignedSeconds(value: number) {
    const sign = value > 0 ? "+" : value < 0 ? "-" : "";
    return `${sign}${formatSeconds(Math.abs(value))}`;
}

function resourceMaxValue(resourceTypeId: number, resourceType: string) {
    switch (resourceTypeId) {
        case 1:
        case 2:
        case 3:
        case 6:
        case 8:
        case 13:
        case 18:
            return 100;
        case 4:
        case 7:
        case 9:
            return 5;
        case 5:
        case 12:
        case 19:
            return 6;
        case 11:
            return 10;
        case 16:
            return 4;
        case 17:
            return 120;
        default:
            break;
    }

    switch (resourceType.trim().toLowerCase()) {
        case "rage":
        case "focus":
        case "energy":
        case "runic power":
        case "lunar power":
        case "insanity":
        case "pain":
            return 100;
        case "fury":
            return 120;
        case "combo points":
        case "soul shards":
        case "holy power":
            return 5;
        case "runes":
        case "chi":
        case "essence":
            return 6;
        case "maelstrom":
            return 10;
        case "arcane charges":
            return 4;
        default:
            return 1;
    }
}

function TimelineRow({
    series,
    durationMs,
    toneClass,
}: {
    series: AbilityTimelineSeries;
    durationMs: number;
    toneClass: string;
}) {
    return (
        <div className="grid gap-3 lg:grid-cols-[220px_minmax(0,1fr)] lg:items-center">
            <div>
                {series.reportUrl ? (
                    <a
                        href={series.reportUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="font-medium text-white transition hover:text-sky-300"
                    >
                        {series.label} ({series.castsMs.length})
                    </a>
                ) : (
                    <p className="font-medium text-white">
                        {series.label} ({series.castsMs.length})
                    </p>
                )}
                {series.subtitle && (
                    <p className="mt-1 text-sm text-slate-400">
                        {series.subtitle}
                    </p>
                )}
            </div>
            <div className="relative rounded-2xl border border-slate-800 bg-slate-950/80 px-3 pb-5 pt-3">
                <div className="relative h-16 overflow-visible rounded-xl bg-slate-900/90">
                    {series.castsMs.map((castMs, index) => {
                        const left = Math.min(
                            100,
                            Math.max(0, (castMs / durationMs) * 100),
                        );

                        return (
                            <div
                                key={`${series.label}-${castMs}-${index}`}
                                className="absolute inset-y-0"
                                style={{ left: `${left}%` }}
                            >
                                <div className="-translate-x-1/2">
                                    <div
                                        className={`rounded-md px-1.5 py-0.5 text-[10px] font-semibold text-white shadow-sm ${toneClass}`}
                                    >
                                        {formatTimelineTimestamp(castMs)}
                                    </div>
                                    <div
                                        className={`mx-auto mt-1 h-9 w-0.5 ${toneClass.replace("bg-", "bg-").replace("/20", "/90")}`}
                                    />
                                </div>
                            </div>
                        );
                    })}
                </div>
                <div className="mt-2 flex justify-between text-xs text-slate-500">
                    <span>0:00</span>
                    <span>{formatTimelineTimestamp(durationMs)}</span>
                </div>
            </div>
        </div>
    );
}

function AbilityTimelineModal({
    timeline,
    loading,
    error,
    onClose,
}: {
    timeline: AbilityTimelineResponse | null;
    loading: boolean;
    error: string | null;
    onClose: () => void;
}) {
    if (!loading && !timeline && !error) {
        return null;
    }

    const durationMs = timeline?.fightDurationMs ?? 1;

    return (
        <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 px-4 py-8 backdrop-blur-sm"
            onClick={onClose}
        >
            <div
                className="max-h-[90vh] w-full max-w-6xl overflow-y-auto rounded-3xl border border-slate-800 bg-slate-900 p-6 shadow-2xl"
                onClick={(event) => event.stopPropagation()}
            >
                <div className="flex items-start justify-between gap-4">
                    <div>
                        <p className="text-sm uppercase tracking-[0.2em] text-sky-400">
                            Ability Timeline
                        </p>
                        <h2 className="mt-2 text-2xl font-semibold text-white">
                            {timeline?.abilityName ?? "Loading timeline"}
                        </h2>
                        <p className="mt-2 text-sm text-slate-400">
                            Compare your cast timing against the elite fight
                            timelines on the same boss.
                        </p>
                    </div>
                    <Button variant="secondary" onClick={onClose}>
                        Close
                    </Button>
                </div>

                {loading && (
                    <div className="mt-8 rounded-3xl border border-slate-800 bg-slate-950/80 p-6 text-sm text-slate-300">
                        Loading ability timeline...
                    </div>
                )}

                {error && (
                    <div className="mt-8 rounded-3xl border border-rose-500/30 bg-rose-950/20 p-6 text-sm text-rose-200">
                        {error}
                    </div>
                )}

                {timeline && (
                    <div className="mt-8 space-y-8">
                        <div className="space-y-4">
                            <h3 className="text-lg font-semibold text-white">
                                You
                            </h3>
                            <TimelineRow
                                series={timeline.player}
                                durationMs={durationMs}
                                toneClass="bg-sky-500/90"
                            />
                        </div>

                        <div className="space-y-4">
                            <h3 className="text-lg font-semibold text-white">
                                Elite
                            </h3>
                            <div className="space-y-4">
                                {timeline.elite.map((series) => (
                                    <TimelineRow
                                        key={`${series.label}-${series.reportUrl ?? "elite"}`}
                                        series={series}
                                        durationMs={durationMs}
                                        toneClass="bg-amber-500/90"
                                    />
                                ))}
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}

function renderProgressView(
    reportJob: ReportJob,
    fight: Fight,
    character: Character,
) {
    const progressPercent =
        reportJob.progress.total > 0
            ? Math.min(
                  100,
                  Math.round(
                      (reportJob.progress.current / reportJob.progress.total) *
                          100,
                  ),
              )
            : 0;
    const isCohortStage =
        reportJob.stage === "cohort" && reportJob.progress.total > 0;

    return (
        <>
            <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                    Analysis Progress
                </p>
                <h1 className="mt-3 text-3xl font-semibold text-white">
                    {reportJob.status === "failed"
                        ? "Analysis failed"
                        : reportJob.status === "completed"
                          ? "Analysis complete"
                          : "Analyzing fight"}
                </h1>
                <p className="mt-4 max-w-2xl text-slate-300">
                    {reportJob.message}
                </p>

                <div className="mt-6 rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                        <div>
                            <p className="text-sm uppercase tracking-[0.2em] text-sky-400">
                                {getStageLabel(reportJob.stage)}
                            </p>
                            <p className="mt-2 text-sm text-slate-300">
                                Status: {reportJob.status}
                            </p>
                        </div>
                        <div className="text-sm text-slate-400">
                            {isCohortStage
                                ? `Elite member ${reportJob.progress.current} of ${reportJob.progress.total}`
                                : `Step ${Math.min(
                                      reportJob.progress.current,
                                      reportJob.progress.total || 1,
                                  )} of ${reportJob.progress.total || 1}`}
                        </div>
                    </div>

                    <div className="mt-4 h-3 overflow-hidden rounded-full bg-slate-800">
                        <div
                            className={`h-full rounded-full transition-all ${
                                reportJob.status === "failed"
                                    ? "bg-rose-500"
                                    : reportJob.status === "completed"
                                      ? "bg-emerald-500"
                                      : "bg-sky-500"
                            }`}
                            style={{ width: `${progressPercent}%` }}
                        />
                    </div>

                    {reportJob.error && (
                        <p className="mt-4 text-sm text-rose-400">
                            {reportJob.error}
                        </p>
                    )}
                </div>
            </div>

            <div className="grid gap-6 lg:grid-cols-2">
                {renderSummaryCard("Fight Summary", [
                    { label: "Encounter", value: fight.name },
                    { label: "Difficulty", value: fight.difficulty },
                    { label: "Result", value: fight.kill ? "Kill" : "Wipe" },
                    {
                        label: "Combat Time",
                        value: formatKillTime(fight.killTime),
                    },
                ])}

                {renderSummaryCard("Character Summary", [
                    { label: "Name", value: character.name },
                    {
                        label: "Server",
                        value: character.serverName || "Unknown server",
                    },
                    {
                        label: "Class & Spec",
                        value: `${character.class} - ${character.spec}`,
                    },
                    { label: "Role", value: character.role },
                    {
                        label: "Talent",
                        value: character.talentCalculatorUrl
                            ? "Open in Wowhead"
                            : "Unavailable",
                        href: character.talentCalculatorUrl,
                        fullWidth: true,
                    },
                ])}
            </div>

            <div className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                <h2 className="text-lg font-semibold text-white">Pipeline</h2>
                <div className="mt-4 grid gap-3 lg:grid-cols-3">
                    {reportStages.map((step) => {
                        const state = getStageCompletionState(
                            reportJob.stage,
                            reportJob.status,
                            step.key,
                        );
                        const stateClasses =
                            state === "complete"
                                ? "border-emerald-700/60 bg-emerald-950/20 text-emerald-200"
                                : state === "active"
                                  ? "border-sky-500/60 bg-sky-950/20 text-sky-100"
                                  : state === "failed"
                                    ? "border-rose-700/60 bg-rose-950/20 text-rose-200"
                                    : "border-slate-800 bg-slate-900/80 text-slate-400";

                        return (
                            <div
                                key={step.key}
                                className={`rounded-3xl border p-4 ${stateClasses}`}
                            >
                                <p className="text-xs uppercase tracking-[0.2em]">
                                    {state}
                                </p>
                                <p className="mt-2 font-medium">{step.label}</p>
                            </div>
                        );
                    })}
                </div>
            </div>

            <div className="flex flex-col gap-3 sm:flex-row">
                <Link to="/select">
                    <Button variant="secondary">Back to selection</Button>
                </Link>
                <Link to="/analyze">
                    <Button variant="secondary">Back to logs</Button>
                </Link>
            </div>
        </>
    );
}

function ResourceTimelineRow({
    series,
    durationMs,
    resourceTypeId,
    resourceType,
    toneClass,
}: {
    series: ResourceTimelineSeries;
    durationMs: number;
    resourceTypeId: number;
    resourceType: string;
    toneClass: string;
}) {
    const rowDurationMs = series.durationMs || durationMs;
    const maxValue = resourceMaxValue(resourceTypeId, resourceType);
    const points = series.samples
        .map((sample) => {
            const x = Math.min(
                100,
                Math.max(0, (sample.timestampMs / rowDurationMs) * 100),
            );
            const y =
                100 -
                Math.min(100, Math.max(0, (sample.value / maxValue) * 100));
            return `${x},${y}`;
        })
        .join(" ");

    return (
        <div className="grid gap-3 lg:grid-cols-[220px_minmax(0,1fr)] lg:items-center">
            <div>
                {series.reportUrl ? (
                    <a
                        href={series.reportUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="font-medium text-white transition hover:text-sky-300"
                    >
                        {series.label}
                    </a>
                ) : (
                    <p className="font-medium text-white">{series.label}</p>
                )}
                {series.subtitle && (
                    <p className="mt-1 text-sm text-slate-400">
                        {series.subtitle}
                    </p>
                )}
                <p className="mt-1 text-xs text-slate-500">
                    {series.samples.length} samples
                    {series.wasteMarkersMs?.length
                        ? `, ${series.wasteMarkersMs.length} waste/full markers`
                        : ""}
                </p>
            </div>
            <div className="relative rounded-2xl border border-slate-800 bg-slate-950/80 px-3 pb-5 pt-3">
                <div className="relative h-24 overflow-visible rounded-xl bg-slate-900/90">
                    <svg
                        viewBox="0 0 100 100"
                        preserveAspectRatio="none"
                        className="absolute inset-0 h-full w-full"
                    >
                        <polyline
                            points={points}
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2"
                            className={toneClass}
                            vectorEffect="non-scaling-stroke"
                        />
                    </svg>
                    {(series.wasteMarkersMs ?? []).map((markerMs, index) => {
                        const left = Math.min(
                            100,
                            Math.max(0, (markerMs / rowDurationMs) * 100),
                        );
                        return (
                            <div
                                key={`${series.label}-waste-${markerMs}-${index}`}
                                className="absolute inset-y-0 w-[1.5%] -translate-x-1/2 rounded-sm bg-rose-500/50"
                                style={{ left: `${left}%` }}
                                title={`Waste/full marker at ${formatTimelineTimestamp(markerMs)}`}
                            />
                        );
                    })}
                </div>
                <div className="mt-2 flex justify-between text-xs text-slate-500">
                    <span>0:00</span>
                    <span>{formatTimelineTimestamp(rowDurationMs)}</span>
                </div>
            </div>
        </div>
    );
}

function ResourceTimelineModal({
    timeline,
    loading,
    error,
    onClose,
}: {
    timeline: ResourceTimelineResponse | null;
    loading: boolean;
    error: string | null;
    onClose: () => void;
}) {
    if (!loading && !timeline && !error) {
        return null;
    }

    const durationMs = timeline?.fightDurationMs ?? 1;

    return (
        <div
            className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 px-4 py-8 backdrop-blur-sm"
            onClick={onClose}
        >
            <div
                className="max-h-[90vh] w-full max-w-6xl overflow-y-auto rounded-3xl border border-slate-800 bg-slate-900 p-6 shadow-2xl"
                onClick={(event) => event.stopPropagation()}
            >
                <div className="flex items-start justify-between gap-4">
                    <div>
                        <p className="text-sm uppercase tracking-[0.2em] text-sky-400">
                            Resource Timeline
                        </p>
                        <h2 className="mt-2 text-2xl font-semibold text-white">
                            {timeline?.resourceType ?? "Loading timeline"}
                        </h2>
                        <p className="mt-2 text-sm text-slate-400">
                            Compare resource level/change samples against elite
                            fights. Red markers indicate wasted resource or
                            full-resource samples.
                        </p>
                    </div>
                    <Button variant="secondary" onClick={onClose}>
                        Close
                    </Button>
                </div>

                {loading && (
                    <div className="mt-8 rounded-3xl border border-slate-800 bg-slate-950/80 p-6 text-sm text-slate-300">
                        Loading resource timeline...
                    </div>
                )}

                {error && (
                    <div className="mt-8 rounded-3xl border border-rose-500/30 bg-rose-950/20 p-6 text-sm text-rose-200">
                        {error}
                    </div>
                )}

                {timeline && (
                    <div className="mt-8 space-y-8">
                        <div className="space-y-4">
                            <h3 className="text-lg font-semibold text-white">
                                You
                            </h3>
                            <ResourceTimelineRow
                                series={timeline.player}
                                durationMs={durationMs}
                                resourceTypeId={timeline.resourceTypeId}
                                resourceType={timeline.resourceType}
                                toneClass="text-sky-400"
                            />
                        </div>

                        <div className="space-y-4">
                            <h3 className="text-lg font-semibold text-white">
                                Elite
                            </h3>
                            <div className="space-y-4">
                                {timeline.elite.map((series) => (
                                    <ResourceTimelineRow
                                        key={`${series.label}-${series.reportUrl ?? "elite"}`}
                                        series={series}
                                        durationMs={durationMs}
                                        resourceTypeId={timeline.resourceTypeId}
                                        resourceType={timeline.resourceType}
                                        toneClass="text-amber-400"
                                    />
                                ))}
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}

export function ReportPage() {
    usePageTitle("Report");
    const {
        reportId,
        reportJob,
        reportResult,
        setReportJob,
        setReportResult,
        setError,
    } = useAnalyzeStore();
    const browserSelectedCharacter = useBrowserStore(
        (state) => state.selectedCharacter,
    );
    const [abilitySearch, setAbilitySearch] = useState("");
    const [abilityFilter, setAbilityFilter] = useState<ComparisonFilter>("all");
    const [buffSearch, setBuffSearch] = useState("");
    const [buffFilter, setBuffFilter] = useState<ComparisonFilter>("all");
    const [resourceSearch, setResourceSearch] = useState("");
    const [resourceFilter, setResourceFilter] =
        useState<ComparisonFilter>("all");
    const [showAllAbilities, setShowAllAbilities] = useState(false);
    const [showAllBuffs, setShowAllBuffs] = useState(false);
    const [showAllResources, setShowAllResources] = useState(false);
    const [timelineAbilityId, setTimelineAbilityId] = useState<number | null>(
        null,
    );
    const [timelineCache, setTimelineCache] = useState<
        Record<number, AbilityTimelineResponse>
    >({});
    const [timelineLoading, setTimelineLoading] = useState(false);
    const [timelineError, setTimelineError] = useState<string | null>(null);
    const [resourceTimelineTypeId, setResourceTimelineTypeId] = useState<
        number | null
    >(null);
    const [resourceTimelineCache, setResourceTimelineCache] = useState<
        Record<number, ResourceTimelineResponse>
    >({});
    const [resourceTimelineLoading, setResourceTimelineLoading] =
        useState(false);
    const [resourceTimelineError, setResourceTimelineError] = useState<
        string | null
    >(null);
    const [dismissedWarnings, setDismissedWarnings] = useState<
        Record<string, boolean>
    >({});
    const reportJobId = reportJob?.jobId;
    const reportJobStatus = reportJob?.status;

    useEffect(() => {
        if (
            !reportJobId ||
            reportJobStatus === "completed" ||
            reportJobStatus === "failed"
        ) {
            return;
        }

        let cancelled = false;

        const refreshJob = async () => {
            try {
                const nextJob = await getReportJob(reportJobId);
                if (cancelled) {
                    return;
                }

                setReportJob(nextJob);
                if (nextJob.result) {
                    setReportResult(nextJob.result);
                }
                if (nextJob.status === "failed" && nextJob.error) {
                    setError(nextJob.error);
                }
            } catch (err) {
                if (cancelled) {
                    return;
                }

                setError(
                    err instanceof Error
                        ? err.message
                        : "Failed to refresh report status",
                );
            }
        };

        void refreshJob();
        const intervalId = window.setInterval(() => {
            void refreshJob();
        }, 5000);

        return () => {
            cancelled = true;
            window.clearInterval(intervalId);
        };
    }, [reportJobId, reportJobStatus, setError, setReportJob, setReportResult]);

    useEffect(() => {
        setTimelineAbilityId(null);
        setTimelineCache({});
        setTimelineLoading(false);
        setTimelineError(null);
        setResourceTimelineTypeId(null);
        setResourceTimelineCache({});
        setResourceTimelineLoading(false);
        setResourceTimelineError(null);
        setDismissedWarnings({});
    }, [reportJobId]);

    if (!reportJob && !reportResult) {
        return (
            <section className="space-y-8">
                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                        Report
                    </p>
                    <h1 className="mt-3 text-3xl font-semibold text-white">
                        No analysis data available
                    </h1>
                    <p className="mt-4 max-w-2xl text-slate-300">
                        Please go back to the analyze page and select a fight
                        and character.
                    </p>
                    <div className="mt-8 flex flex-col gap-3 sm:flex-row">
                        <Link to="/analyze">
                            <Button variant="secondary">Back to analyze</Button>
                        </Link>
                        <Link to="/">
                            <Button variant="secondary">Back to home</Button>
                        </Link>
                    </div>
                </div>
            </section>
        );
    }

    if (!reportResult && reportJob) {
        return (
            <section className="space-y-8">
                {renderProgressView(
                    reportJob,
                    reportJob.fight,
                    reportJob.character,
                )}
            </section>
        );
    }

    const { fight, character, comparison, warnings = [], ai } = reportResult!;
    const visibleWarnings = warnings.filter(
        (warning) => !dismissedWarnings[`${warning.kind}-${warning.title}`],
    );
    const showAIWarning =
        Boolean(ai.warning) && !dismissedWarnings["ai-warning"];
    const { cohortStats } = comparison;
    const specKey = `${character.spec} ${character.class}`;
    const trackedPriority = trackedSpecPriorities[specKey] ?? {
        abilities: [],
        buffs: [],
    };
    const filteredAbilityUsage = comparison.abilityUsage.filter((entry) => {
        const searchMatches = entry.abilityName
            .toLowerCase()
            .includes(abilitySearch.trim().toLowerCase());
        return searchMatches && matchesFilter(entry.countDelta, abilityFilter);
    });
    const filteredBuffUptimes = comparison.buffUptimes.filter((entry) => {
        const searchMatches = entry.abilityName
            .toLowerCase()
            .includes(buffSearch.trim().toLowerCase());
        return searchMatches && matchesFilter(entry.uptimeDelta, buffFilter);
    });
    const filteredResourceUsage = comparison.resourceUsage.filter((entry) => {
        const searchMatches = entry.resourceType
            .toLowerCase()
            .includes(resourceSearch.trim().toLowerCase());
        return (
            searchMatches &&
            matchesFilter(entry.fullWindowDeltaSeconds * -1, resourceFilter)
        );
    });
    const orderedAbilityUsage = sortByTrackedPriority(
        filteredAbilityUsage,
        trackedPriority.abilities,
    );
    const orderedBuffUptimes = sortByTrackedPriority(
        filteredBuffUptimes,
        trackedPriority.buffs,
    );
    const orderedResourceUsage = [...filteredResourceUsage].sort(
        (left, right) => {
            if (left.fullWindowDeltaSeconds === right.fullWindowDeltaSeconds) {
                if (left.fullMarkerDelta === right.fullMarkerDelta) {
                    if (left.spentDelta !== right.spentDelta) {
                        return (
                            Math.abs(right.spentDelta) -
                            Math.abs(left.spentDelta)
                        );
                    }
                    return left.resourceType.localeCompare(right.resourceType);
                }
                return right.fullMarkerDelta - left.fullMarkerDelta;
            }
            return right.fullWindowDeltaSeconds - left.fullWindowDeltaSeconds;
        },
    );
    const visibleAbilityUsage = showAllAbilities
        ? orderedAbilityUsage
        : orderedAbilityUsage.slice(0, 10);
    const visibleBuffUptimes = showAllBuffs
        ? orderedBuffUptimes
        : orderedBuffUptimes.slice(0, 10);
    const visibleResourceUsage = showAllResources
        ? orderedResourceUsage
        : orderedResourceUsage.slice(0, 10);
    const matchingBrowserCharacter =
        browserSelectedCharacter &&
        browserSelectedCharacter.name === character.name &&
        browserSelectedCharacter.serverName === character.serverName
            ? browserSelectedCharacter
            : null;
    const fightUrl = reportId
        ? `https://www.warcraftlogs.com/reports/${reportId}#fight=${fight.id}`
        : undefined;
    const blizzardRegion = mapRegionToBlizzardSlug(
        matchingBrowserCharacter?.serverRegion,
    );
    const serverSlug =
        matchingBrowserCharacter?.serverSlug ||
        slugifyServerName(character.serverName);
    const characterUrl =
        blizzardRegion && serverSlug && character.name
            ? `https://worldofwarcraft.blizzard.com/en-us/character/${blizzardRegion}/${serverSlug}/${character.name.toLowerCase()}`
            : undefined;
    const selectedTimeline =
        timelineAbilityId !== null ? timelineCache[timelineAbilityId] : null;
    const selectedResourceTimeline =
        resourceTimelineTypeId !== null
            ? resourceTimelineCache[resourceTimelineTypeId]
            : null;

    const openAbilityTimeline = async (abilityId: number) => {
        if (!reportJobId) {
            setTimelineError("Timeline data is not available for this report.");
            return;
        }

        setTimelineAbilityId(abilityId);
        setTimelineError(null);

        if (timelineCache[abilityId]) {
            return;
        }

        setTimelineLoading(true);
        try {
            const response = await getAbilityTimeline(reportJobId, abilityId);
            setTimelineCache((current) => ({
                ...current,
                [abilityId]: response,
            }));
        } catch (error) {
            setTimelineError(
                error instanceof Error
                    ? error.message
                    : "Failed to load ability timeline",
            );
        } finally {
            setTimelineLoading(false);
        }
    };

    const openResourceTimeline = async (resourceTypeId: number) => {
        if (!reportJobId) {
            setResourceTimelineError(
                "Timeline data is not available for this report.",
            );
            return;
        }

        setResourceTimelineTypeId(resourceTypeId);
        setResourceTimelineError(null);

        if (resourceTimelineCache[resourceTypeId]) {
            return;
        }

        setResourceTimelineLoading(true);
        try {
            const response = await getResourceTimeline(
                reportJobId,
                resourceTypeId,
            );
            setResourceTimelineCache((current) => ({
                ...current,
                [resourceTypeId]: response,
            }));
        } catch (error) {
            setResourceTimelineError(
                error instanceof Error
                    ? error.message
                    : "Failed to load resource timeline",
            );
        } finally {
            setResourceTimelineLoading(false);
        }
    };

    return (
        <>
            <section className="space-y-8">
                {reportJob &&
                    reportJob.status !== "completed" &&
                    renderProgressView(
                        reportJob,
                        reportJob.fight,
                        reportJob.character,
                    )}

                <div className="rounded-3xl border border-slate-800 bg-slate-900/80 p-8">
                    <p className="text-sm uppercase tracking-[0.25em] text-sky-400">
                        Report
                    </p>
                    <h1 className="mt-3 text-3xl font-semibold text-white">
                        Analysis Results
                    </h1>
                    <p className="mt-4 max-w-2xl text-slate-300">
                        The analyzer assumes you have done the basic preparation
                        for a raid: enchantments & gems & food & flask.
                        Comparison against {cohortStats.sampleSize} elite
                        players.
                    </p>
                </div>

                <div className="grid gap-6 lg:grid-cols-2">
                    {renderSummaryCard(
                        "Fight Summary",
                        [
                            { label: "Encounter", value: fight.name },
                            { label: "Difficulty", value: fight.difficulty },
                            {
                                label: "Result",
                                value: fight.kill ? "Kill" : "Wipe",
                            },
                            {
                                label: "Kill Time",
                                value: formatKillTime(fight.killTime),
                            },
                        ],
                        fightUrl,
                    )}

                    {renderSummaryCard("Character Summary", [
                        {
                            label: "Name",
                            value: character.name,
                            href: characterUrl,
                        },
                        {
                            label: "Server",
                            value: character.serverName || "Unknown server",
                        },
                        {
                            label: "Class & Spec",
                            value: `${character.class} ${character.spec}`,
                        },
                        { label: "Role", value: character.role },
                        {
                            label: "Talent",
                            value: character.talentCalculatorUrl
                                ? "Open in Wowhead"
                                : "Unavailable",
                            href: character.talentCalculatorUrl,
                            fullWidth: true,
                        },
                    ])}
                </div>

                {visibleWarnings.map((warning) => {
                    const warningKey = `${warning.kind}-${warning.title}`;
                    return (
                        <div
                            key={warningKey}
                            className="relative rounded-3xl border border-amber-700/60 bg-amber-950/20 p-6 pr-14"
                        >
                            <button
                                type="button"
                                aria-label="Dismiss warning"
                                className="absolute right-4 top-4 flex h-8 w-8 items-center justify-center rounded-full border border-amber-700/60 text-amber-200 transition hover:border-amber-400 hover:text-white"
                                onClick={() =>
                                    setDismissedWarnings((current) => ({
                                        ...current,
                                        [warningKey]: true,
                                    }))
                                }
                            >
                                x
                            </button>
                            <h2 className="text-lg font-semibold text-amber-100">
                                {warning.title}
                            </h2>
                            <p className="mt-2 text-sm text-amber-300">
                                {warning.message}
                            </p>
                        </div>
                    );
                })}

                {showAIWarning && ai.warning && (
                    <div className="relative rounded-3xl border border-amber-700/60 bg-amber-950/20 p-6 pr-14">
                        <button
                            type="button"
                            aria-label="Dismiss warning"
                            className="absolute right-4 top-4 flex h-8 w-8 items-center justify-center rounded-full border border-amber-700/60 text-amber-200 transition hover:border-amber-400 hover:text-white"
                            onClick={() =>
                                setDismissedWarnings((current) => ({
                                    ...current,
                                    "ai-warning": true,
                                }))
                            }
                        >
                            x
                        </button>
                        <p className="text-sm text-amber-300">{ai.warning}</p>
                    </div>
                )}

                <div className="space-y-6">
                    <h2 className="text-xl font-semibold text-white">
                        Insights
                    </h2>

                    {ai.insights.length > 0 ? (
                        <div className="grid gap-6 lg:grid-cols-3">
                            {ai.insights.map((insight) => (
                                <article
                                    key={insight.metricKey}
                                    className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6"
                                >
                                    <h3 className="text-lg font-semibold text-white">
                                        {insight.title}
                                    </h3>
                                    <p className="mt-3 text-sm text-slate-300">
                                        {insight.summary}
                                    </p>
                                    {insight.caution && (
                                        <p className="mt-3 text-xs text-amber-300">
                                            {insight.caution}
                                        </p>
                                    )}
                                </article>
                            ))}
                        </div>
                    ) : (
                        <div className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                            <p className="text-sm text-slate-300">
                                AI insights are not available for this report,
                                but the deterministic comparison metrics below
                                are still valid.
                            </p>
                        </div>
                    )}

                    {ai.focusRecommendation && (
                        <div className="rounded-3xl border border-sky-500/30 bg-sky-950/20 p-6">
                            <p className="text-sm uppercase tracking-[0.2em] text-sky-400">
                                Focus Recommendation
                            </p>
                            <h3 className="mt-3 text-xl font-semibold text-white">
                                {ai.focusRecommendation.title}
                            </h3>
                            <p className="mt-3 text-slate-200">
                                {ai.focusRecommendation.recommendation}
                            </p>
                            <p className="mt-3 text-sm text-slate-400">
                                {ai.focusRecommendation.reasoning}
                            </p>
                        </div>
                    )}
                </div>

                {reportResult!.cohort.length > 0 && (
                    <div className="space-y-6">
                        <h2 className="text-xl font-semibold text-white">
                            Elite Logs
                        </h2>
                        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
                            {reportResult!.cohort.map((entry) => (
                                <a
                                    key={`${entry.reportId}-${entry.fightId}-${entry.name}`}
                                    href={entry.reportUrl}
                                    target="_blank"
                                    rel="noreferrer"
                                    className="rounded-3xl border border-slate-800 bg-slate-950/80 p-5 transition hover:border-sky-500/60"
                                >
                                    <p className="text-lg font-semibold text-white">
                                        {entry.name}
                                    </p>
                                    <p className="mt-1 text-sm text-slate-300">
                                        {entry.class} {entry.spec}
                                    </p>
                                    <p className="mt-1 text-xs text-slate-400">
                                        {entry.server || "Unknown server"}
                                        {entry.serverRegion
                                            ? ` | ${entry.serverRegion}`
                                            : ""}
                                    </p>
                                </a>
                            ))}
                        </div>
                    </div>
                )}

                <div className="space-y-6">
                    <h2 className="text-xl font-semibold text-white">
                        Ability Usage
                    </h2>
                    <div className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                        <div className="mb-6 flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                            <p className="text-xs text-slate-400">
                                * all combat data is normalized by fight length
                            </p>
                            <div className="flex flex-col gap-3 md:flex-row">
                                <input
                                    type="text"
                                    value={abilitySearch}
                                    onChange={(event) =>
                                        setAbilitySearch(event.target.value)
                                    }
                                    placeholder="Filter abilities"
                                    className="w-full rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-sm text-white outline-none transition focus:border-sky-500 md:max-w-sm"
                                />
                                <select
                                    value={abilityFilter}
                                    onChange={(event) =>
                                        setAbilityFilter(
                                            event.target
                                                .value as ComparisonFilter,
                                        )
                                    }
                                    className="rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-sm text-white outline-none transition focus:border-sky-500"
                                >
                                    <option value="all">All abilities</option>
                                    <option value="behind">Only behind</option>
                                    <option value="ahead">Only ahead</option>
                                </select>
                            </div>
                        </div>
                        {orderedAbilityUsage.length > 0 ? (
                            <div className="overflow-x-auto">
                                <table className="min-w-full text-left text-sm">
                                    <thead className="text-slate-400">
                                        <tr className="border-b border-slate-800">
                                            <th className="pb-3 pr-4 font-medium">
                                                Ability
                                            </th>
                                            <th className="pb-3 pr-4 font-medium">
                                                You
                                            </th>
                                            <th className="pb-3 pr-4 font-medium">
                                                Elite (MED)
                                            </th>
                                            <th className="pb-3 pr-4 font-medium">
                                                Difference
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {visibleAbilityUsage.map((entry) => (
                                            <tr
                                                key={entry.abilityId}
                                                className="cursor-pointer border-b border-slate-900 align-top transition hover:bg-slate-900/60 last:border-b-0"
                                                onClick={() => {
                                                    void openAbilityTimeline(
                                                        entry.abilityId,
                                                    );
                                                }}
                                            >
                                                <td className="py-4 pr-4">
                                                    <p className="font-medium text-white">
                                                        {entry.abilityName}
                                                    </p>
                                                </td>
                                                <td className="py-4 pr-4 text-slate-200">
                                                    <p
                                                        title={`Total casts: ${entry.playerCount}`}
                                                    >
                                                        {entry.playerCastsPerMinute.toFixed(
                                                            2,
                                                        )}
                                                    </p>
                                                </td>
                                                <td className="py-4 pr-4 text-slate-300">
                                                    <p
                                                        title={`Elite median total casts: ${Number.isInteger(entry.cohortMedianCount) ? entry.cohortMedianCount : entry.cohortMedianCount.toFixed(1)}`}
                                                    >
                                                        {entry.cohortMedianPerMinute.toFixed(
                                                            2,
                                                        )}
                                                    </p>
                                                </td>
                                                <td className="py-4 pr-4">
                                                    <p
                                                        title={`Raw total delta: ${formatSigned(entry.countDelta, "", 1)}`}
                                                        className={
                                                            entry.perMinuteDelta >=
                                                            0
                                                                ? "text-emerald-400"
                                                                : "text-rose-400"
                                                        }
                                                    >
                                                        {formatSigned(
                                                            entry.perMinuteDelta,
                                                            "",
                                                            2,
                                                        )}
                                                    </p>
                                                </td>
                                                <td className="py-4 pr-4">
                                                    {entry.caution && (
                                                        <p className="text-xs text-amber-300">
                                                            {entry.caution}
                                                        </p>
                                                    )}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                                {orderedAbilityUsage.length > 10 && (
                                    <div className="mt-4 flex justify-center">
                                        <Button
                                            type="button"
                                            variant="secondary"
                                            onClick={() =>
                                                setShowAllAbilities(
                                                    (current) => !current,
                                                )
                                            }
                                        >
                                            {showAllAbilities
                                                ? "Show fewer abilities"
                                                : `Show all abilities (${orderedAbilityUsage.length})`}
                                        </Button>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <p className="text-sm text-slate-300">
                                No ability usage comparisons matched this
                                filter.
                            </p>
                        )}
                    </div>
                </div>

                <div className="space-y-6">
                    <h2 className="text-xl font-semibold text-white">
                        Buff Uptimes
                    </h2>
                    <div className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                        <div className="mb-6 flex flex-col gap-3 md:flex-row">
                            <input
                                type="text"
                                value={buffSearch}
                                onChange={(event) =>
                                    setBuffSearch(event.target.value)
                                }
                                placeholder="Filter buffs"
                                className="w-full rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-sm text-white outline-none transition focus:border-sky-500 md:max-w-sm"
                            />
                            <select
                                value={buffFilter}
                                onChange={(event) =>
                                    setBuffFilter(
                                        event.target.value as ComparisonFilter,
                                    )
                                }
                                className="rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-sm text-white outline-none transition focus:border-sky-500"
                            >
                                <option value="all">All buffs</option>
                                <option value="behind">Only behind</option>
                                <option value="ahead">Only ahead</option>
                            </select>
                        </div>
                        {orderedBuffUptimes.length > 0 ? (
                            <div className="overflow-x-auto">
                                <table className="min-w-full text-left text-sm">
                                    <thead className="text-slate-400">
                                        <tr className="border-b border-slate-800">
                                            <th className="pb-3 pr-4 font-medium">
                                                Buff
                                            </th>
                                            <th className="pb-3 pr-4 font-medium">
                                                You
                                            </th>
                                            <th className="pb-3 pr-4 font-medium">
                                                Cohort
                                            </th>
                                            <th className="pb-3 pr-4 font-medium">
                                                Delta
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {visibleBuffUptimes.map((entry) => (
                                            <tr
                                                key={entry.abilityId}
                                                className="border-b border-slate-900 align-top last:border-b-0"
                                            >
                                                <td className="py-4 pr-4">
                                                    <p className="font-medium text-white">
                                                        {entry.abilityName}
                                                    </p>
                                                </td>
                                                <td className="py-4 pr-4 text-slate-200">
                                                    <p>
                                                        {entry.playerUptimePct.toFixed(
                                                            1,
                                                        )}
                                                        %
                                                    </p>
                                                </td>
                                                <td className="py-4 pr-4 text-slate-300">
                                                    <p>
                                                        {entry.cohortMedianUptimePct.toFixed(
                                                            1,
                                                        )}
                                                        %
                                                    </p>
                                                </td>
                                                <td className="py-4 pr-4">
                                                    <p
                                                        className={
                                                            entry.uptimeDelta >=
                                                            0
                                                                ? "text-emerald-400"
                                                                : "text-rose-400"
                                                        }
                                                    >
                                                        {formatSigned(
                                                            entry.uptimeDelta,
                                                            "%",
                                                            1,
                                                        )}
                                                    </p>
                                                </td>
                                                <td className="py-4 pr-4">
                                                    {entry.caution && (
                                                        <p className="text-xs text-amber-300">
                                                            {entry.caution}
                                                        </p>
                                                    )}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                                {orderedBuffUptimes.length > 10 && (
                                    <div className="mt-4 flex justify-center">
                                        <Button
                                            type="button"
                                            variant="secondary"
                                            onClick={() =>
                                                setShowAllBuffs(
                                                    (current) => !current,
                                                )
                                            }
                                        >
                                            {showAllBuffs
                                                ? "Show fewer buffs"
                                                : `Show all buffs (${orderedBuffUptimes.length})`}
                                        </Button>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <p className="text-sm text-slate-300">
                                No buff uptime comparisons matched this filter.
                            </p>
                        )}
                    </div>
                </div>

                <div className="space-y-6">
                    <h2 className="text-xl font-semibold text-white">
                        Resource Usage
                    </h2>
                    <div className="rounded-3xl border border-slate-800 bg-slate-950/80 p-6">
                        <div className="mb-6 flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                            <p className="text-xs text-slate-400">
                                * resource rows summarize the same timeline
                                data: full/waste markers and total time spent in
                                full windows
                            </p>
                            <div className="flex flex-col gap-3 md:flex-row">
                                <input
                                    type="text"
                                    value={resourceSearch}
                                    onChange={(event) =>
                                        setResourceSearch(event.target.value)
                                    }
                                    placeholder="Filter resources"
                                    className="w-full rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-sm text-white outline-none transition focus:border-sky-500 md:max-w-sm"
                                />
                                <select
                                    value={resourceFilter}
                                    onChange={(event) =>
                                        setResourceFilter(
                                            event.target
                                                .value as ComparisonFilter,
                                        )
                                    }
                                    className="rounded-2xl border border-slate-700 bg-slate-900 px-4 py-3 text-sm text-white outline-none transition focus:border-sky-500"
                                >
                                    <option value="all">All resources</option>
                                    <option value="behind">
                                        More full time than elites
                                    </option>
                                    <option value="ahead">
                                        Less full time than elites
                                    </option>
                                </select>
                            </div>
                        </div>
                        {orderedResourceUsage.length > 0 ? (
                            <div className="overflow-x-auto">
                                <table className="min-w-full text-left text-sm">
                                    <thead className="text-slate-400">
                                        <tr className="border-b border-slate-800">
                                            <th className="pb-3 pr-4 font-medium">
                                                Resource
                                            </th>
                                            <th className="pb-3 pr-4 font-medium">
                                                Used
                                            </th>
                                            <th className="pb-3 pr-4 font-medium">
                                                Full Markers
                                            </th>
                                            <th className="pb-3 pr-4 font-medium">
                                                Full Window Time
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {visibleResourceUsage.map((entry) => (
                                            <tr
                                                key={entry.resourceTypeId}
                                                className="cursor-pointer border-b border-slate-900 align-top transition hover:bg-slate-900/60 last:border-b-0"
                                                onClick={() => {
                                                    void openResourceTimeline(
                                                        entry.resourceTypeId,
                                                    );
                                                }}
                                            >
                                                <td className="py-4 pr-4">
                                                    <p className="font-medium text-white">
                                                        {entry.resourceType}
                                                    </p>
                                                    <p className="mt-1 text-xs text-slate-500">
                                                        {entry.sampleSize} elite
                                                        comparisons
                                                    </p>
                                                </td>
                                                <td className="py-4 pr-4 text-slate-200">
                                                    <p>
                                                        {entry.playerSpent.toFixed(
                                                            0,
                                                        )}
                                                    </p>
                                                    <p
                                                        className={
                                                            entry.spentDelta >=
                                                            0
                                                                ? "mt-1 text-xs text-emerald-400"
                                                                : "mt-1 text-xs text-rose-400"
                                                        }
                                                    >
                                                        {formatSigned(
                                                            entry.spentDelta,
                                                            "",
                                                            0,
                                                        )}{" "}
                                                        vs elite
                                                    </p>
                                                </td>
                                                <td className="py-4 pr-4 text-slate-200">
                                                    <p>
                                                        {
                                                            entry.playerFullMarkerCount
                                                        }
                                                    </p>
                                                    <p
                                                        className={
                                                            entry.fullMarkerDelta <=
                                                            0
                                                                ? "mt-1 text-xs text-emerald-400"
                                                                : "mt-1 text-xs text-rose-400"
                                                        }
                                                    >
                                                        {formatSigned(
                                                            entry.fullMarkerDelta,
                                                            "",
                                                            0,
                                                        )}{" "}
                                                        vs elite
                                                    </p>
                                                </td>
                                                <td className="py-4 pr-4">
                                                    <p className="text-slate-200">
                                                        {formatSeconds(
                                                            entry.playerFullWindowSeconds,
                                                        )}
                                                    </p>
                                                    <p
                                                        className={
                                                            entry.fullWindowDeltaSeconds <=
                                                            0
                                                                ? "mt-1 text-xs text-emerald-400"
                                                                : "mt-1 text-xs text-rose-400"
                                                        }
                                                    >
                                                        {formatSignedSeconds(
                                                            entry.fullWindowDeltaSeconds,
                                                        )}{" "}
                                                        vs elite
                                                    </p>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                                {orderedResourceUsage.length > 10 && (
                                    <div className="mt-4 flex justify-center">
                                        <Button
                                            type="button"
                                            variant="secondary"
                                            onClick={() =>
                                                setShowAllResources(
                                                    (current) => !current,
                                                )
                                            }
                                        >
                                            {showAllResources
                                                ? "Show fewer resources"
                                                : `Show all resources (${orderedResourceUsage.length})`}
                                        </Button>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <p className="text-sm text-slate-300">
                                No resource usage comparisons matched this
                                filter.
                            </p>
                        )}
                    </div>
                </div>

                <div className="mt-8 flex flex-col gap-3 sm:flex-row">
                    <Link to="/analyze">
                        <Button variant="secondary">
                            Analyze another fight
                        </Button>
                    </Link>
                    <Link to="/">
                        <Button variant="secondary">Back to home</Button>
                    </Link>
                </div>
            </section>
            <AbilityTimelineModal
                timeline={selectedTimeline}
                loading={timelineLoading}
                error={timelineError}
                onClose={() => {
                    setTimelineAbilityId(null);
                    setTimelineError(null);
                    setTimelineLoading(false);
                }}
            />
            <ResourceTimelineModal
                timeline={selectedResourceTimeline}
                loading={resourceTimelineLoading}
                error={resourceTimelineError}
                onClose={() => {
                    setResourceTimelineTypeId(null);
                    setResourceTimelineError(null);
                    setResourceTimelineLoading(false);
                }}
            />
        </>
    );
}
