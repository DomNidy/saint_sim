package simc

import (
	"context"
	"errors"
	"testing"
)

type testMetadataProvider struct {
	metadata map[int]WowheadItemMetadata
	errors   map[int]error
	hits     map[int]int
}

func (provider *testMetadataProvider) Metadata(_ context.Context, itemID int) (WowheadItemMetadata, error) {
	if provider.hits == nil {
		provider.hits = map[int]int{}
	}
	provider.hits[itemID]++

	if err, ok := provider.errors[itemID]; ok {
		return WowheadItemMetadata{}, err
	}

	if metadata, ok := provider.metadata[itemID]; ok {
		return metadata, nil
	}

	return WowheadItemMetadata{}, errors.New("not found")
}

func TestBuildGearPreview(t *testing.T) {
	t.Parallel()

	provider := &testMetadataProvider{
		metadata: map[int]WowheadItemMetadata{
			250458: {
				Name:    "Host Commander's Casque",
				IconURL: "https://example.com/icon-250458.jpg",
				URL:     "https://www.wowhead.com/item=250458",
			},
			249671: {
				Name:    "Gnarlroot Spinecleaver",
				IconURL: "https://example.com/icon-249671.jpg",
				URL:     "https://www.wowhead.com/item=249671",
			},
		},
		errors: map[int]error{
			258876: errors.New("failed"),
		},
	}

	const input = `
# Host Commander's Casque (253)
head=,id=250458,bonus_id=6652/12667/13577/13333/12787
# Gnarlroot Spinecleaver (250)
main_hand=,id=249671,enchant_id=3368,bonus_id=12786/6652
### Gear from Bags
# Frayed Guise (201)
# head=,id=258876,bonus_id=13611,drop_level=90,gem_id=23121,gem_id2=23122,ilevel=210
`

	groups, err := BuildGearPreview(t.Context(), input, provider)
	if err != nil {
		t.Fatalf("BuildGearPreview() error = %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("len(groups) = %d, want 2", len(groups))
	}

	if groups[0].Slot != "head" || groups[1].Slot != "main_hand" {
		t.Fatalf("group order = [%s,%s], want [head,main_hand]", groups[0].Slot, groups[1].Slot)
	}

	headItems := groups[0].Items
	if len(headItems) != 2 {
		t.Fatalf("len(headItems) = %d, want 2", len(headItems))
	}

	if headItems[0].Source != "equipped" || headItems[1].Source != "bag" {
		t.Fatalf("sources = [%s,%s], want [equipped,bag]", headItems[0].Source, headItems[1].Source)
	}

	if headItems[0].DisplayName != "Host Commander's Casque" {
		t.Fatalf("display name = %q, want Host Commander's Casque", headItems[0].DisplayName)
	}

	if headItems[0].WowheadData != "item=250458&bonus=6652:12667:13577:13333:12787" {
		t.Fatalf("wowhead data = %q", headItems[0].WowheadData)
	}

	if headItems[1].WowheadData != "item=258876&bonus=13611&gems=23121:23122&ilvl=210" {
		t.Fatalf("wowhead data = %q", headItems[1].WowheadData)
	}

	if headItems[1].ItemLevel == nil || *headItems[1].ItemLevel != 201 {
		t.Fatalf("bag item level = %v, want 201", headItems[1].ItemLevel)
	}

	if headItems[1].IconURL == nil || *headItems[1].IconURL != defaultIconURL {
		t.Fatalf("fallback icon not set, got %v", headItems[1].IconURL)
	}

	if provider.hits[250458] != 1 || provider.hits[249671] != 1 || provider.hits[258876] != 1 {
		t.Fatalf("metadata hit counts = %#v, want each item fetched once", provider.hits)
	}

	if headItems[0].Fingerprint == "" || headItems[1].Fingerprint == "" {
		t.Fatal("fingerprints should be set")
	}

	if headItems[0].Fingerprint == headItems[1].Fingerprint {
		t.Fatal("fingerprints should differ for distinct lines/sources")
	}
}

func TestBuildGearPreviewRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	_, err := BuildGearPreview(t.Context(), "", &testMetadataProvider{})
	if err == nil {
		t.Fatal("BuildGearPreview() error = nil, want error")
	}

	_, err = BuildGearPreview(t.Context(), "# comments only", &testMetadataProvider{})
	if err == nil {
		t.Fatal("BuildGearPreview() error = nil, want error")
	}
}
