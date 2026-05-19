import { describe, expect, it } from "vitest";
import { matchFightCharacter } from "../lib/characterMatching";
import type { BrowserCharacter, Character } from "../types";

const browserCharacter: BrowserCharacter = {
    id: 1,
    name: "Jaicher",
    class: "Death Knight",
    serverName: "Area 52",
    serverRegion: "US",
    serverSlug: "area-52",
};

function fightCharacter(overrides: Partial<Character> = {}): Character {
    return {
        id: 10,
        name: "Jaicher",
        class: "Death Knight",
        spec: "Blood",
        role: "Tank",
        serverName: "Area-52",
        ...overrides,
    };
}

describe("matchFightCharacter", () => {
    it("matches realm names with formatting differences", () => {
        const matched = matchFightCharacter(browserCharacter, [
            fightCharacter(),
        ]);

        expect(matched?.id).toBe(10);
    });

    it("falls back to a unique same-name same-class character", () => {
        const matched = matchFightCharacter(browserCharacter, [
            fightCharacter({ serverName: "Unknown Realm" }),
        ]);

        expect(matched?.id).toBe(10);
    });

    it("does not guess between multiple same-name same-class characters", () => {
        const matched = matchFightCharacter(browserCharacter, [
            fightCharacter({ id: 10, serverName: "Unknown Realm" }),
            fightCharacter({ id: 11, serverName: "Other Realm" }),
        ]);

        expect(matched).toBeNull();
    });
});
