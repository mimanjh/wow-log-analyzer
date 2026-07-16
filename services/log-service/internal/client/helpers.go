package client

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wow-log-analyzer/services/log-service/internal/types"
)

func matchRankingToCharacter(characters []types.CharacterOption, ranking WCLRankingEntry) (types.CharacterOption, bool) {
	rankingName := strings.TrimSpace(ranking.Name)
	rankingClass := normalizeClass(ranking.Class)
	rankingSpec := normalizeSpec(ranking.Spec)

	for _, character := range characters {
		if strings.TrimSpace(character.Name) != rankingName {
			continue
		}
		if rankingClass != "" && normalizeClass(character.Class) != rankingClass {
			continue
		}
		if rankingSpec != "" && normalizeSpec(character.Spec) != rankingSpec {
			continue
		}
		return character, true
	}

	for _, character := range characters {
		if strings.TrimSpace(character.Name) == rankingName {
			return character, true
		}
	}

	return types.CharacterOption{}, false
}

func normalizeCastEvents(reportStartTime int64, raw []map[string]interface{}, abilityNames map[int]string) []types.CastEvent {
	events := make([]types.CastEvent, 0, len(raw))
	for _, event := range raw {
		events = append(events, types.CastEvent{
			Timestamp: parseTimestamp(reportStartTime, event["timestamp"]),
			Ability:   parseAbility(event, abilityNames),
			SourceID:  parseActorID(event, "sourceID", "source"),
		})
	}
	return events
}

func normalizeDamageEvents(reportStartTime int64, raw []map[string]interface{}, abilityNames map[int]string) []types.DamageEvent {
	events := make([]types.DamageEvent, 0, len(raw))
	for _, event := range raw {
		events = append(events, types.DamageEvent{
			Timestamp: parseTimestamp(reportStartTime, event["timestamp"]),
			Ability:   parseAbility(event, abilityNames),
			SourceID:  parseActorID(event, "sourceID", "source"),
			TargetID:  parseActorID(event, "targetID", "target"),
			Amount:    parseInt(event["amount"]),
		})
	}
	return events
}

func normalizeHealEvents(reportStartTime int64, raw []map[string]interface{}, abilityNames map[int]string) []types.HealEvent {
	events := make([]types.HealEvent, 0, len(raw))
	for _, event := range raw {
		events = append(events, types.HealEvent{
			Timestamp: parseTimestamp(reportStartTime, event["timestamp"]),
			Ability:   parseAbility(event, abilityNames),
			SourceID:  parseActorID(event, "sourceID", "source"),
			TargetID:  parseActorID(event, "targetID", "target"),
			Amount:    parseInt(event["amount"]),
		})
	}
	return events
}

func normalizeBuffEvents(reportStartTime int64, raw []map[string]interface{}, abilityNames map[int]string) []types.BuffEvent {
	events := make([]types.BuffEvent, 0, len(raw))
	for _, event := range raw {
		eventType := fmt.Sprintf("%v", event["type"])
		normalizedType := ""
		switch {
		case strings.HasPrefix(eventType, "apply"):
			normalizedType = "apply"
		case strings.HasPrefix(eventType, "refresh"):
			normalizedType = "refresh"
		case strings.HasPrefix(eventType, "remove"):
			normalizedType = "remove"
		default:
			continue
		}

		ability := parseAbility(event, abilityNames)
		ability.IsBuff = true
		events = append(events, types.BuffEvent{
			Timestamp: parseTimestamp(reportStartTime, event["timestamp"]),
			Ability:   ability,
			SourceID:  parseActorID(event, "sourceID", "source"),
			TargetID:  parseActorID(event, "targetID", "target"),
			EventType: normalizedType,
		})
	}
	return events
}

func normalizeCooldownEvents(reportStartTime int64, raw []map[string]interface{}, abilityNames map[int]string) []types.CooldownEvent {
	events := make([]types.CooldownEvent, 0)
	seen := map[int]int{}

	for _, event := range raw {
		ability := parseAbility(event, abilityNames)
		if ability.ID == 0 {
			continue
		}
		seen[ability.ID]++
	}

	for _, event := range raw {
		ability := parseAbility(event, abilityNames)
		if ability.ID == 0 || seen[ability.ID] > 5 {
			continue
		}
		ability.IsMajorCD = true
		events = append(events, types.CooldownEvent{
			Timestamp: parseTimestamp(reportStartTime, event["timestamp"]),
			Ability:   ability,
			SourceID:  parseActorID(event, "sourceID", "source"),
			EventType: "start",
		})
	}

	return events
}

func normalizeResourceEvents(reportStartTime int64, raw []map[string]interface{}, playerID int) []types.ResourceEvent {
	events := make([]types.ResourceEvent, 0, len(raw))

	for _, event := range raw {
		eventsByType := make(map[int]types.ResourceEvent)
		timestamp := parseTimestamp(reportStartTime, event["timestamp"])
		sourceID := parseActorID(event, "sourceID", "source")
		targetID := parseActorID(event, "targetID", "target")
		if playerID != 0 && sourceID != playerID && targetID != playerID {
			continue
		}

		resourceTypeID, hasTopLevelResource := parseResourceTypeIDFromEvent(event)
		if hasTopLevelResource {
			change := parseFloat(event["resourceChange"])
			if change == 0 {
				change = parseFloat(event["amount"])
			}
			amount := parseFloat(event["resourceAmount"])
			if amount == 0 {
				amount = parseFloat(event["current"])
			}
			waste := parseFloat(event["waste"])
			maxAmount := parseFloat(event["maxResourceAmount"])
			if maxAmount == 0 {
				maxAmount = parseFloat(event["resourceChangeMax"])
			}

			eventsByType[resourceTypeID] = types.ResourceEvent{
				Timestamp:      timestamp,
				SourceID:       sourceID,
				TargetID:       targetID,
				ResourceTypeID: resourceTypeID,
				ResourceType:   resourceTypeName(resourceTypeID),
				Amount:         amount,
				Change:         change,
				Waste:          waste,
				MaxAmount:      maxAmount,
			}
		}

		if playerID == 0 || sourceID == playerID {
			mergeResourceDataEvents(eventsByType, timestamp, sourceID, targetID, event["sourceResources"])
			mergeAdditionalResourceEvents(eventsByType, timestamp, sourceID, targetID, event["classResources"])
		}
		if playerID == 0 || targetID == playerID {
			mergeResourceDataEvents(eventsByType, timestamp, sourceID, targetID, event["targetResources"])
		}

		for _, normalized := range eventsByType {
			events = append(events, normalized)
		}
	}

	return events
}

func parseResourceTypeIDFromEvent(event map[string]interface{}) (int, bool) {
	if value, ok := event["resourceChangeType"]; ok {
		return parseInt(value), true
	}
	if value, ok := event["type"]; ok {
		return parseResourceTypeID(value)
	}
	return 0, false
}

func parseResourceTypeID(value interface{}) (int, bool) {
	switch typed := value.(type) {
	case float64, float32, int, int64, json.Number:
		return parseInt(typed), true
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return 0, false
		}
		parsed, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return 0, false
		}
		return int(parsed), true
	default:
		return 0, false
	}
}

func mergeResourceDataEvents(eventsByType map[int]types.ResourceEvent, timestamp time.Time, sourceID, targetID int, value interface{}) {
	resourceMap, ok := value.(map[string]interface{})
	if !ok {
		return
	}

	if event, ok := resourceEventFromResourceData(timestamp, sourceID, targetID, resourceMap); ok {
		mergeResourceEvent(eventsByType, event)
	}
	mergeAdditionalResourceEvents(eventsByType, timestamp, sourceID, targetID, resourceMap["additionalResources"])
}

func mergeAdditionalResourceEvents(eventsByType map[int]types.ResourceEvent, timestamp time.Time, sourceID, targetID int, value interface{}) {
	switch typed := value.(type) {
	case []interface{}:
		for _, entry := range typed {
			mergeAdditionalResourceEvents(eventsByType, timestamp, sourceID, targetID, entry)
		}
	case map[string]interface{}:
		if event, ok := resourceEventFromResourceData(timestamp, sourceID, targetID, typed); ok {
			mergeResourceEvent(eventsByType, event)
			return
		}
		for _, entry := range typed {
			mergeAdditionalResourceEvents(eventsByType, timestamp, sourceID, targetID, entry)
		}
	}
}

func resourceEventFromResourceData(timestamp time.Time, sourceID, targetID int, data map[string]interface{}) (types.ResourceEvent, bool) {
	resourceTypeID := parseInt(data["resourceType"])
	if resourceTypeID == 0 {
		if parsed, ok := parseResourceTypeID(data["type"]); ok {
			resourceTypeID = parsed
		}
	}
	if resourceTypeID == 0 {
		return types.ResourceEvent{}, false
	}

	amount := parseFloat(data["resourceAmount"])
	if amount == 0 {
		amount = parseFloat(data["amount"])
	}
	maxAmount := parseFloat(data["resourceCap"])
	if maxAmount == 0 {
		maxAmount = parseFloat(data["max"])
	}
	if maxAmount == 0 {
		maxAmount = parseFloat(data["resourceMax"])
	}
	if maxAmount == 0 {
		maxAmount = parseFloat(data["maxResourceAmount"])
	}
	cost := parseFloat(data["resourceCost"])
	if cost == 0 {
		cost = parseFloat(data["cost"])
	}

	return types.ResourceEvent{
		Timestamp:      timestamp,
		SourceID:       sourceID,
		TargetID:       targetID,
		ResourceTypeID: resourceTypeID,
		ResourceType:   resourceTypeName(resourceTypeID),
		Amount:         amount,
		Change:         -cost,
		MaxAmount:      maxAmount,
	}, true
}

func mergeResourceEvent(eventsByType map[int]types.ResourceEvent, event types.ResourceEvent) {
	existing, exists := eventsByType[event.ResourceTypeID]
	if !exists {
		eventsByType[event.ResourceTypeID] = event
		return
	}
	if existing.Amount == 0 {
		existing.Amount = event.Amount
	}
	if existing.MaxAmount == 0 {
		existing.MaxAmount = event.MaxAmount
	}
	if existing.Change == 0 {
		existing.Change = event.Change
	}
	if existing.Waste == 0 {
		existing.Waste = event.Waste
	}
	if existing.ResourceType == "" {
		existing.ResourceType = event.ResourceType
	}
	eventsByType[event.ResourceTypeID] = existing
}

func parseTimestamp(reportStartTime int64, value interface{}) time.Time {
	ms := absoluteReportTimestamp(reportStartTime, parseInt64(value))
	if ms <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}

func parseAbility(event map[string]interface{}, abilityNames map[int]string) types.Ability {
	if abilityValue, ok := event["ability"]; ok {
		if abilityMap, ok := abilityValue.(map[string]interface{}); ok {
			return types.Ability{
				ID:   parseInt(abilityMap["gameID"]),
				Name: parseString(abilityMap["name"]),
			}
		}
	}

	abilityID := parseInt(event["abilityGameID"])
	abilityName := parseString(event["abilityName"])
	if abilityName == "" && abilityID != 0 {
		abilityName = abilityNames[abilityID]
	}

	return types.Ability{
		ID:   abilityID,
		Name: abilityName,
	}
}

func parseActorID(event map[string]interface{}, idKey, objectKey string) int {
	if value, ok := event[idKey]; ok {
		return parseInt(value)
	}
	if objectValue, ok := event[objectKey]; ok {
		if actorMap, ok := objectValue.(map[string]interface{}); ok {
			return parseInt(actorMap["id"])
		}
	}
	return 0
}

func parseInt(value interface{}) int {
	return int(parseInt64(value))
}

func parseInt64(value interface{}) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int:
		return int64(typed)
	case int64:
		return typed
	case json.Number:
		parsed, _ := typed.Int64()
		return parsed
	case string:
		parsed, _ := strconv.ParseInt(typed, 10, 64)
		return parsed
	default:
		return 0
	}
}

func parseString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func parseFloat(value interface{}) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		parsed, _ := typed.Float64()
		return parsed
	case string:
		parsed, _ := strconv.ParseFloat(typed, 64)
		return parsed
	default:
		return 0
	}
}

func resourceTypeName(id int) string {
	switch id {
	case 0:
		return "Mana"
	case 1:
		return "Rage"
	case 2:
		return "Focus"
	case 3:
		return "Energy"
	case 4:
		return "Combo Points"
	case 5:
		return "Runes"
	case 6:
		return "Runic Power"
	case 7:
		return "Soul Shards"
	case 8:
		return "Lunar Power"
	case 9:
		return "Holy Power"
	case 10:
		return "Alternate Power"
	case 11:
		return "Maelstrom"
	case 12:
		return "Chi"
	case 13:
		return "Insanity"
	case 14:
		return "Obsolete"
	case 15:
		return "Obsolete 2"
	case 16:
		return "Arcane Charges"
	case 17:
		return "Fury"
	case 18:
		return "Pain"
	case 19:
		return "Essence"
	default:
		return fmt.Sprintf("Resource %d", id)
	}
}

func normalizeClass(raw string) string {
	return strings.TrimSpace(raw)
}

func normalizeSpec(raw string) string {
	return strings.TrimSpace(raw)
}

func classMatchKey(raw string) string {
	return normalizeAlphaNumericKey(raw)
}

func specMatchKey(raw string) string {
	return normalizeAlphaNumericKey(raw)
}

func normalizeAlphaNumericKey(raw string) string {
	var builder strings.Builder
	builder.Grow(len(raw))

	for _, r := range strings.ToLower(strings.TrimSpace(raw)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func rankingClassFilterValue(raw string) string {
	switch classMatchKey(raw) {
	case "deathknight":
		return "DeathKnight"
	case "demonhunter":
		return "DemonHunter"
	default:
		return compactAlphaNumeric(raw)
	}
}

func rankingSpecFilterValue(raw string) string {
	normalized := strings.TrimSpace(raw)
	if normalized == "" {
		return ""
	}

	lower := strings.ToLower(normalized)
	return strings.ToUpper(lower[:1]) + lower[1:]
}

func compactAlphaNumeric(raw string) string {
	var builder strings.Builder
	builder.Grow(len(raw))

	for _, r := range strings.TrimSpace(raw) {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func inferRole(spec string) string {
	switch strings.ToLower(strings.TrimSpace(spec)) {
	case "holy", "discipline", "restoration", "mistweaver", "preservation":
		return "Healer"
	case "protection", "blood", "guardian", "vengeance", "brewmaster":
		return "Tank"
	default:
		return "DPS"
	}
}

func rankingDifficultyID(name string) int {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "lfr":
		return 17
	case "normal":
		return 3
	case "heroic":
		return 4
	case "mythic":
		return 5
	default:
		return 0
	}
}

func isSimilarFightDuration(targetMS, candidateMS int) bool {
	if targetMS <= 0 || candidateMS <= 0 {
		return true
	}

	allowedDifference := int(float64(targetMS) * 0.15)
	if allowedDifference < 15000 {
		allowedDifference = 15000
	}

	difference := targetMS - candidateMS
	if difference < 0 {
		difference = -difference
	}

	return difference <= allowedDifference
}

func normalizedFightFromSelection(selection types.FightSelection) types.NormalizedFight {
	return types.NormalizedFight{
		ID:          selection.ID,
		Name:        selection.Name,
		StartTime:   selection.StartTime,
		EndTime:     selection.EndTime,
		EncounterID: selection.EncounterID,
		Difficulty:  selection.Difficulty,
		BossPercent: selection.BossPercent,
	}
}

func classNameFromID(classID int) string {
	switch classID {
	case 1:
		return "Death Knight"
	case 2:
		return "Druid"
	case 3:
		return "Hunter"
	case 4:
		return "Mage"
	case 5:
		return "Monk"
	case 6:
		return "Paladin"
	case 7:
		return "Priest"
	case 8:
		return "Rogue"
	case 9:
		return "Shaman"
	case 10:
		return "Warlock"
	case 11:
		return "Warrior"
	case 12:
		return "Demon Hunter"
	case 13:
		return "Evoker"
	default:
		return "Unknown"
	}
}

func deriveCharacterClassIDFromRecentReports(character WCLUserCharacter) (int, bool) {
	canonicalMatches := make(map[int]int)
	nameServerMatches := make(map[int]int)

	for _, report := range character.RecentReports.Data {
		for _, rankedCharacter := range report.RankedCharacters {
			if rankedCharacter.ClassID == 0 {
				continue
			}
			if rankedCharacter.CanonicalID != 0 && rankedCharacter.CanonicalID == character.CanonicalID {
				canonicalMatches[rankedCharacter.ClassID]++
				continue
			}
			if strings.EqualFold(strings.TrimSpace(rankedCharacter.Name), strings.TrimSpace(character.Name)) &&
				strings.EqualFold(strings.TrimSpace(rankedCharacter.Server.Name), strings.TrimSpace(character.Server.Name)) {
				nameServerMatches[rankedCharacter.ClassID]++
			}
		}
	}

	if classID, ok := mostFrequentClassID(canonicalMatches); ok {
		return classID, true
	}
	if classID, ok := mostFrequentClassID(nameServerMatches); ok {
		return classID, true
	}

	return 0, false
}

func mostFrequentClassID(counts map[int]int) (int, bool) {
	bestClassID := 0
	bestCount := 0

	for classID, count := range counts {
		if count > bestCount {
			bestClassID = classID
			bestCount = count
		}
	}

	if bestClassID == 0 || bestCount == 0 {
		return 0, false
	}

	return bestClassID, true
}

func findCharacterActorID(name, serverName, serverSlug string, actors []WCLActor) int {
	trimmedName := strings.TrimSpace(name)
	trimmedServerName := strings.TrimSpace(serverName)
	trimmedServerSlug := strings.TrimSpace(serverSlug)

	players := make([]WCLActor, 0, len(actors))
	for _, a := range actors {
		if strings.EqualFold(strings.TrimSpace(a.Type), "Player") {
			players = append(players, a)
		}
	}
	if len(players) == 0 {
		players = actors // fall back to all actors if type field is absent
	}

	for _, actor := range players {
		if !strings.EqualFold(strings.TrimSpace(actor.Name), trimmedName) {
			continue
		}
		actorServer := strings.TrimSpace(actor.Server)
		if trimmedServerName != "" && strings.EqualFold(actorServer, trimmedServerName) {
			return actor.ID
		}
		if trimmedServerSlug != "" && strings.EqualFold(actorServer, trimmedServerSlug) {
			return actor.ID
		}
	}
	for _, actor := range players {
		if strings.EqualFold(strings.TrimSpace(actor.Name), trimmedName) {
			return actor.ID
		}
	}
	return 0
}

func fightHasActor(fight WCLFight, actorID int) bool {
	if actorID == 0 {
		return true
	}
	for _, id := range fight.FriendlyPlayers {
		if id == actorID {
			return true
		}
	}
	return false
}

func isAllowedRaidReportTitle(title string) bool {
	normalized := strings.ToUpper(strings.TrimSpace(title))
	if normalized == "" {
		return false
	}

	return strings.Contains(normalized, "VS / DR / MQD")
}

func absoluteReportTimestamp(reportStartTime, value int64) int64 {
	if reportStartTime > 0 && value >= 0 && value < reportStartTime {
		return reportStartTime + value
	}
	return value
}

func parseFightCharacters(raw json.RawMessage, actors []WCLActor, serverByID map[int]string) ([]types.CharacterOption, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var payload interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	characters := make([]types.CharacterOption, 0)
	seen := make(map[int]struct{})
	collectFightCharacters(payload, "", actors, serverByID, seen, &characters)

	return characters, nil
}

func collectFightCharacters(node interface{}, role string, actors []WCLActor, serverByID map[int]string, seen map[int]struct{}, characters *[]types.CharacterOption) {
	switch typed := node.(type) {
	case map[string]interface{}:
		nextRole := role
		for key, value := range typed {
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "tanks":
				collectFightCharacters(value, "Tank", actors, serverByID, seen, characters)
				continue
			case "healers":
				collectFightCharacters(value, "Healer", actors, serverByID, seen, characters)
				continue
			case "dps":
				collectFightCharacters(value, "DPS", actors, serverByID, seen, characters)
				continue
			}
		}

		if character, ok := buildCharacterOption(typed, nextRole, actors, serverByID); ok {
			if _, exists := seen[character.ID]; !exists {
				seen[character.ID] = struct{}{}
				*characters = append(*characters, character)
			}
		}

		for _, value := range typed {
			collectFightCharacters(value, nextRole, actors, serverByID, seen, characters)
		}
	case []interface{}:
		for _, value := range typed {
			collectFightCharacters(value, role, actors, serverByID, seen, characters)
		}
	}
}

func buildCharacterOption(payload map[string]interface{}, role string, actors []WCLActor, serverByID map[int]string) (types.CharacterOption, bool) {
	id := parseInt(payload["id"])
	name := parseString(payload["name"])
	className := normalizeClass(parseString(payload["type"]))

	if id == 0 || strings.TrimSpace(name) == "" || strings.TrimSpace(className) == "" {
		return types.CharacterOption{}, false
	}
	if strings.EqualFold(className, "Pet") {
		return types.CharacterOption{}, false
	}

	spec := extractSpecName(payload)
	if role == "" {
		role = inferRole(spec)
	}

	actorID := resolveActorID(id, name, className, actors)
	if actorID == 0 {
		actorID = id
	}
	talentLookupIDs := []int{actorID}
	if id != actorID {
		talentLookupIDs = append(talentLookupIDs, id)
	}
	if actor, ok := findActorByID(actorID, actors); ok {
		gameID := int(actor.GameID)
		if gameID > 0 {
			talentLookupIDs = append(talentLookupIDs, gameID)
		}
	}

	return types.CharacterOption{
		ID:              actorID,
		Name:            name,
		Class:           className,
		Spec:            spec,
		Role:            role,
		ServerName:      strings.TrimSpace(serverByID[actorID]),
		TalentLookupIDs: uniquePositiveInts(talentLookupIDs),
	}, true
}

func findActorByID(actorID int, actors []WCLActor) (WCLActor, bool) {
	for _, actor := range actors {
		if actor.ID == actorID {
			return actor, true
		}
	}
	return WCLActor{}, false
}

func resolveActorID(playerDetailID int, name, className string, actors []WCLActor) int {
	for _, actor := range actors {
		if actor.ID == playerDetailID {
			return actor.ID
		}
	}

	trimmedName := strings.TrimSpace(name)
	normalizedClass := normalizeClass(className)

	for _, actor := range actors {
		if strings.TrimSpace(actor.Name) != trimmedName {
			continue
		}
		if normalizedClass != "" && normalizeClass(actor.Type) != normalizedClass {
			continue
		}
		return actor.ID
	}

	for _, actor := range actors {
		if strings.TrimSpace(actor.Name) == trimmedName {
			return actor.ID
		}
	}

	return 0
}

func uniquePositiveInts(values []int) []int {
	unique := make([]int, 0, len(values))
	seen := make(map[int]struct{}, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}

func extractSpecName(payload map[string]interface{}) string {
	if value, ok := payload["spec"]; ok {
		if spec := normalizeSpec(parseString(value)); spec != "" {
			return spec
		}
	}

	specsValue, ok := payload["specs"]
	if !ok {
		return ""
	}

	switch typed := specsValue.(type) {
	case []interface{}:
		for _, entry := range typed {
			switch specEntry := entry.(type) {
			case string:
				if spec := normalizeSpec(specEntry); spec != "" {
					return spec
				}
			case map[string]interface{}:
				if spec := normalizeSpec(parseString(specEntry["spec"])); spec != "" {
					return spec
				}
				if spec := normalizeSpec(parseString(specEntry["name"])); spec != "" {
					return spec
				}
			}
		}
	}

	return ""
}

func extractKilledBossNamesFromSummaries(fights []types.CharacterFightSummary) []string {
	seen := make(map[string]struct{}, len(fights))
	bossNames := make([]string, 0, len(fights))
	for _, fight := range fights {
		name := strings.TrimSpace(fight.Name)
		if !fight.Kill || name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		bossNames = append(bossNames, name)
	}
	return bossNames
}

func extractKilledBossNames(fights []WCLFight) []string {
	if len(fights) == 0 {
		return nil
	}

	bossNames := make([]string, 0, len(fights))
	seen := make(map[string]struct{}, len(fights))
	for _, fight := range fights {
		name := strings.TrimSpace(fight.Name)
		if !isRelevantKilledBossFight(fight) || name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		bossNames = append(bossNames, name)
	}

	return bossNames
}

func hasSuccessfulBossKill(fights []WCLFight) bool {
	for _, fight := range fights {
		if isRelevantKilledBossFight(fight) && strings.TrimSpace(fight.Name) != "" {
			return true
		}
	}

	return false
}

func isRelevantKilledBossFight(fight WCLFight) bool {
	return fight.Kill &&
		fight.EncounterID != 0 &&
		isRaidDifficultyID(fight.Difficulty)
}

func isRaidDifficultyID(difficulty int) bool {
	switch difficulty {
	case 1, 2, 3, 4, 5, 17:
		return true
	default:
		return false
	}
}

func parseCharacterReportsCursor(cursor string) (int, int) {
	page := 1
	offset := 0
	trimmed := strings.TrimSpace(cursor)
	if trimmed == "" {
		return page, offset
	}

	parts := strings.Split(trimmed, ":")
	if len(parts) >= 1 {
		if parsedPage, err := strconv.Atoi(parts[0]); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	if len(parts) >= 2 {
		if parsedOffset, err := strconv.Atoi(parts[1]); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return page, offset
}

func formatCharacterReportsCursor(page, offset int) string {
	return fmt.Sprintf("%d:%d", page, offset)
}
