import type { BrowserCharacter, Character } from "../types";

export function matchFightCharacter(
    browserCharacter: BrowserCharacter | null,
    fightCharacters: Character[],
) {
    if (!browserCharacter) {
        return null;
    }

    const browserName = normalizeName(browserCharacter.name);
    const browserClass = normalizeName(browserCharacter.class);
    const browserServerKeys = [
        normalizeRealm(browserCharacter.serverName),
        normalizeRealm(browserCharacter.serverSlug ?? ""),
    ].filter(Boolean);

    const sameNameCharacters = fightCharacters.filter(
        (character) => normalizeName(character.name) === browserName,
    );

    const serverMatch =
        sameNameCharacters.find((character) => {
            const characterServer = normalizeRealm(character.serverName ?? "");

            return (
                characterServer === "" ||
                browserServerKeys.includes(characterServer)
            );
        }) ?? null;

    if (serverMatch) {
        return serverMatch;
    }

    const sameClassCharacters = sameNameCharacters.filter(
        (character) => normalizeName(character.class) === browserClass,
    );
    if (sameClassCharacters.length === 1) {
        return sameClassCharacters[0];
    }

    if (sameNameCharacters.length === 1) {
        return sameNameCharacters[0];
    }

    return null;
}

function normalizeName(value: string) {
    return value.trim().toLowerCase();
}

function normalizeRealm(value: string) {
    return value.trim().toLowerCase().replace(/[^a-z0-9]/g, "");
}
