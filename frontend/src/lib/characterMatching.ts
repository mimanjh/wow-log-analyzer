import type { BrowserCharacter, Character } from "../types";

export function matchFightCharacter(
    browserCharacter: BrowserCharacter | null,
    fightCharacters: Character[],
) {
    if (!browserCharacter) {
        return null;
    }

    const browserName = browserCharacter.name.trim().toLowerCase();
    const browserServer = browserCharacter.serverName.trim().toLowerCase();

    return (
        fightCharacters.find((character) => {
            const sameName =
                character.name.trim().toLowerCase() === browserName;
            const sameServer =
                !character.serverName ||
                character.serverName.trim().toLowerCase() === browserServer;

            return sameName && sameServer;
        }) ?? null
    );
}
