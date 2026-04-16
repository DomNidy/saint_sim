package simc

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	wowheadItemURLFormat = "https://www.wowhead.com/item=%d"
	defaultIconURL       = "https://wow.zamimg.com/images/wow/icons/large/inv_misc_questionmark.jpg"
)

var slotLabels = map[string]string{
	"head":      "Head",
	"neck":      "Neck",
	"shoulder":  "Shoulder",
	"back":      "Back",
	"chest":     "Chest",
	"wrist":     "Wrist",
	"hands":     "Hands",
	"waist":     "Waist",
	"legs":      "Legs",
	"feet":      "Feet",
	"finger":    "Finger",
	"trinket":   "Trinket",
	"main_hand": "Main Hand",
	"off_hand":  "Off Hand",
}

var orderedSlots = []string{
	"head", "neck", "shoulder", "back", "chest", "wrist", "hands", "waist",
	"legs", "feet", "finger", "trinket", "main_hand", "off_hand",
}

type GearPreviewGroup struct {
	Slot  string
	Label string
	Items []GearPreviewItem
}

type GearPreviewItem struct {
	Fingerprint string
	Slot        string
	Name        string
	DisplayName string
	ItemID      int
	ItemLevel   *int
	IconURL     *string
	WowheadURL  string
	WowheadData string
	Source      string
	RawLine     string
}

type WowheadItemMetadata struct {
	Name      string
	ItemLevel *int
	IconURL   string
	URL       string
}

type ItemMetadataProvider interface {
	Metadata(ctx context.Context, itemID int) (WowheadItemMetadata, error)
}

type parsedEquipmentItem struct {
	Slot          string
	RawLine       string
	ItemName      string
	DisplayName   string
	ItemID        int
	BonusIDs      []int
	EnchantID     *int
	GemIDs        []int
	ExplicitItemL *int
	CommentItemLv *int
	Source        string
	Fingerprint   string
}

func BuildGearPreview(
	ctx context.Context,
	simcAddonExport string,
	provider ItemMetadataProvider,
) ([]GearPreviewGroup, error) {
	if strings.TrimSpace(simcAddonExport) == "" {
		return nil, fmt.Errorf("simc_addon_export is required")
	}

	parsed := Parse(simcAddonExport)
	if len(parsed.equipmentItems) == 0 {
		return nil, fmt.Errorf("no equipment lines found in addon export")
	}

	items := make([]parsedEquipmentItem, 0, len(parsed.equipmentItems))
	itemIDs := make(map[int]struct{})

	for idx := range parsed.equipmentItems {
		lineItem, ok := parsePreviewEquipmentItem(parsed.equipmentItems[idx])
		if !ok {
			continue
		}

		items = append(items, lineItem)
		itemIDs[lineItem.ItemID] = struct{}{}
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no valid equipment lines found in addon export")
	}

	metadataByID := make(map[int]WowheadItemMetadata, len(itemIDs))
	for itemID := range itemIDs {
		metadata, err := provider.Metadata(ctx, itemID)
		if err == nil {
			metadataByID[itemID] = metadata
		}
	}

	groupsBySlot := map[string][]GearPreviewItem{}
	for idx := range items {
		item := items[idx]
		metadata, hasMetadata := metadataByID[item.ItemID]

		displayName := item.DisplayName
		if displayName == "" {
			displayName = item.ItemName
		}
		if displayName == "" && hasMetadata {
			displayName = metadata.Name
		}
		if displayName == "" {
			displayName = fmt.Sprintf("Item %d", item.ItemID)
		}

		itemName := item.ItemName
		if itemName == "" && hasMetadata {
			itemName = metadata.Name
		}
		if itemName == "" {
			itemName = displayName
		}

		itemLevel := item.CommentItemLv
		if itemLevel == nil {
			itemLevel = item.ExplicitItemL
		}
		if itemLevel == nil && hasMetadata {
			itemLevel = metadata.ItemLevel
		}

		iconURL := metadata.IconURL
		if iconURL == "" {
			iconURL = defaultIconURL
		}

		wowheadURL := metadata.URL
		if wowheadURL == "" {
			wowheadURL = fmt.Sprintf(wowheadItemURLFormat, item.ItemID)
		}

		iconURLCopy := iconURL
		groupsBySlot[item.Slot] = append(groupsBySlot[item.Slot], GearPreviewItem{
			Fingerprint: item.Fingerprint,
			Slot:        item.Slot,
			Name:        itemName,
			DisplayName: displayName,
			ItemID:      item.ItemID,
			ItemLevel:   itemLevel,
			IconURL:     &iconURLCopy,
			WowheadURL:  wowheadURL,
			WowheadData: buildWowheadData(item),
			Source:      item.Source,
			RawLine:     item.RawLine,
		})
	}

	orderedGroups := orderGearPreviewGroups(groupsBySlot)

	return orderedGroups, nil
}

func orderGearPreviewGroups(groupsBySlot map[string][]GearPreviewItem) []GearPreviewGroup {
	ordered := make([]GearPreviewGroup, 0, len(groupsBySlot))
	seen := map[string]struct{}{}

	for _, slot := range orderedSlots {
		items, ok := groupsBySlot[slot]
		if !ok {
			continue
		}

		ordered = append(ordered, GearPreviewGroup{
			Slot:  slot,
			Label: slotLabel(slot),
			Items: items,
		})
		seen[slot] = struct{}{}
	}

	unknownSlots := make([]string, 0)
	for slot := range groupsBySlot {
		if _, ok := seen[slot]; ok {
			continue
		}
		unknownSlots = append(unknownSlots, slot)
	}
	sort.Strings(unknownSlots)

	for _, slot := range unknownSlots {
		ordered = append(ordered, GearPreviewGroup{
			Slot:  slot,
			Label: slotLabel(slot),
			Items: groupsBySlot[slot],
		})
	}

	return ordered
}

func slotLabel(slot string) string {
	if label, ok := slotLabels[slot]; ok {
		return label
	}

	parts := strings.Split(slot, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}

	return strings.Join(parts, " ")
}

func parsePreviewEquipmentItem(item equipmentItem) (parsedEquipmentItem, bool) {
	slot, attributes, ok := parseEquipmentAssignment(item.equipment)
	if !ok {
		return parsedEquipmentItem{}, false
	}

	itemID, ok := parseIntAttribute(attributes, "id")
	if !ok {
		return parsedEquipmentItem{}, false
	}

	displayName, commentItemLv := parseItemCommentMetadata(item.name)
	source := "equipped"
	if item.bagItem {
		source = "bag"
	}

	parsed := parsedEquipmentItem{
		Slot:        slot,
		RawLine:     item.equipment,
		ItemName:    displayName,
		DisplayName: displayName,
		ItemID:      itemID,
		BonusIDs:    parseIntListAttribute(attributes, "bonus_id", "/"),
		EnchantID:   parseOptionalIntAttribute(attributes, "enchant_id"),
		GemIDs:      parseGemIDs(attributes),
		ExplicitItemL: firstNonNil(
			parseOptionalIntAttribute(attributes, "ilevel"),
			parseOptionalIntAttribute(attributes, "ilvl"),
		),
		CommentItemLv: commentItemLv,
		Source:        source,
		Fingerprint:   fingerprintForItem(item.equipment, source),
	}

	return parsed, true
}

func parseEquipmentAssignment(line string) (string, map[string]string, bool) {
	slot, rest, ok := strings.Cut(line, "=")
	if !ok || slot == "" {
		return "", nil, false
	}

	attributes := map[string]string{}
	for _, segment := range strings.Split(rest, ",") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}

		key, value, keyOK := strings.Cut(segment, "=")
		if !keyOK {
			continue
		}

		attributes[key] = value
	}

	return slot, attributes, true
}

func parseItemCommentMetadata(comment string) (string, *int) {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return "", nil
	}

	openIdx := strings.LastIndex(comment, "(")
	closeIdx := strings.LastIndex(comment, ")")
	if openIdx <= 0 || closeIdx != len(comment)-1 || openIdx >= closeIdx {
		return comment, nil
	}

	levelText := strings.TrimSpace(comment[openIdx+1 : closeIdx])
	level, err := strconv.Atoi(levelText)
	if err != nil {
		return comment, nil
	}

	name := strings.TrimSpace(comment[:openIdx])
	if name == "" {
		name = comment
	}

	return name, &level
}

func parseIntAttribute(attributes map[string]string, key string) (int, bool) {
	value, ok := attributes[key]
	if !ok || value == "" {
		return 0, false
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}

	return parsed, true
}

func parseOptionalIntAttribute(attributes map[string]string, key string) *int {
	value, ok := parseIntAttribute(attributes, key)
	if !ok {
		return nil
	}

	return &value
}

func parseIntListAttribute(attributes map[string]string, key, separator string) []int {
	value, ok := attributes[key]
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, separator)
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		parsed, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		result = append(result, parsed)
	}

	return result
}

func parseGemIDs(attributes map[string]string) []int {
	keys := make([]string, 0)
	for key := range attributes {
		if strings.HasPrefix(key, "gem_id") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	gems := make([]int, 0, len(keys))
	for _, key := range keys {
		parsed, err := strconv.Atoi(attributes[key])
		if err == nil {
			gems = append(gems, parsed)
		}
	}

	return gems
}

func firstNonNil(values ...*int) *int {
	for _, value := range values {
		if value != nil {
			return value
		}
	}

	return nil
}

func fingerprintForItem(rawLine, source string) string {
	normalized := strings.TrimSpace(strings.ToLower(rawLine)) + "|" + source
	hash := sha1.Sum([]byte(normalized))

	return hex.EncodeToString(hash[:])
}

func buildWowheadData(item parsedEquipmentItem) string {
	pairs := []string{fmt.Sprintf("item=%d", item.ItemID)}
	if len(item.BonusIDs) > 0 {
		pairs = append(pairs, "bonus="+joinInts(item.BonusIDs, ":"))
	}
	if item.EnchantID != nil {
		pairs = append(pairs, fmt.Sprintf("ench=%d", *item.EnchantID))
	}
	if len(item.GemIDs) > 0 {
		pairs = append(pairs, "gems="+joinInts(item.GemIDs, ":"))
	}
	if item.ExplicitItemL != nil {
		pairs = append(pairs, fmt.Sprintf("ilvl=%d", *item.ExplicitItemL))
	}

	return strings.Join(pairs, "&")
}

func joinInts(values []int, separator string) string {
	if len(values) == 0 {
		return ""
	}

	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, strconv.Itoa(value))
	}

	return strings.Join(result, separator)
}

type wowheadXMLProvider struct {
	httpClient *http.Client
	ttl        time.Duration
	nowFn      func() time.Time

	mu    sync.RWMutex
	cache map[int]cachedMetadata
}

type cachedMetadata struct {
	metadata  WowheadItemMetadata
	expiresAt time.Time
}

func NewWowheadXMLProvider(httpClient *http.Client, ttl time.Duration) ItemMetadataProvider {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	return &wowheadXMLProvider{
		httpClient: httpClient,
		ttl:        ttl,
		nowFn:      time.Now,
		cache:      map[int]cachedMetadata{},
	}
}

func (provider *wowheadXMLProvider) Metadata(
	ctx context.Context,
	itemID int,
) (WowheadItemMetadata, error) {
	if metadata, ok := provider.readFromCache(itemID); ok {
		return metadata, nil
	}

	metadata, err := provider.fetchMetadata(ctx, itemID)
	if err != nil {
		return WowheadItemMetadata{}, err
	}

	provider.writeToCache(itemID, metadata)

	return metadata, nil
}

func (provider *wowheadXMLProvider) readFromCache(itemID int) (WowheadItemMetadata, bool) {
	provider.mu.RLock()
	cached, ok := provider.cache[itemID]
	provider.mu.RUnlock()
	if !ok {
		return WowheadItemMetadata{}, false
	}

	if provider.nowFn().After(cached.expiresAt) {
		return WowheadItemMetadata{}, false
	}

	return cached.metadata, true
}

func (provider *wowheadXMLProvider) writeToCache(itemID int, metadata WowheadItemMetadata) {
	provider.mu.Lock()
	provider.cache[itemID] = cachedMetadata{
		metadata:  metadata,
		expiresAt: provider.nowFn().Add(provider.ttl),
	}
	provider.mu.Unlock()
}

func (provider *wowheadXMLProvider) fetchMetadata(
	ctx context.Context,
	itemID int,
) (WowheadItemMetadata, error) {
	parsedURL := fmt.Sprintf("https://www.wowhead.com/item=%d&xml", itemID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL, nil)
	if err != nil {
		return WowheadItemMetadata{}, fmt.Errorf("new wowhead request: %w", err)
	}
	req.Header.Set("User-Agent", "saint-sim/1.0")

	resp, err := provider.httpClient.Do(req)
	if err != nil {
		return WowheadItemMetadata{}, fmt.Errorf("fetch wowhead xml: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return WowheadItemMetadata{}, fmt.Errorf("wowhead xml returned status %d", resp.StatusCode)
	}

	type wowheadItemXML struct {
		XMLName xml.Name `xml:"wowhead"`
		Item    struct {
			ID    int    `xml:"id,attr"`
			Name  string `xml:"name"`
			Level int    `xml:"level"`
			Icon  string `xml:"icon"`
			Link  string `xml:"link"`
		} `xml:"item"`
	}

	var payload wowheadItemXML
	if decodeErr := xml.NewDecoder(resp.Body).Decode(&payload); decodeErr != nil {
		return WowheadItemMetadata{}, fmt.Errorf("decode wowhead xml: %w", decodeErr)
	}

	if payload.Item.ID == 0 {
		return WowheadItemMetadata{}, fmt.Errorf("wowhead xml missing item id")
	}

	urlValue := payload.Item.Link
	if urlValue == "" {
		urlValue = fmt.Sprintf(wowheadItemURLFormat, itemID)
	}

	iconURL := ""
	if payload.Item.Icon != "" {
		iconURL = fmt.Sprintf(
			"https://wow.zamimg.com/images/wow/icons/large/%s.jpg",
			url.PathEscape(strings.ToLower(payload.Item.Icon)),
		)
	}

	var itemLevel *int
	if payload.Item.Level > 0 {
		level := payload.Item.Level
		itemLevel = &level
	}

	return WowheadItemMetadata{
		Name:      strings.TrimSpace(payload.Item.Name),
		ItemLevel: itemLevel,
		IconURL:   iconURL,
		URL:       urlValue,
	}, nil
}
