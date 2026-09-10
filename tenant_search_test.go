package main

import "testing"

func TestTenantFilterDecision(t *testing.T) {
	hits := []SearchHit{{ID: "onboard", Metadata: map[string]any{"tenant": "acme"}}, {ID: "admin", Metadata: map[string]any{"tenant": "other"}}}
	got := tenantHits(hits, "acme")
	if len(got) != 1 || got[0].ID != "onboard" {
		t.Fatalf("tenant filter = %#v", got)
	}
}

func tenantHits(hits []SearchHit, tenant string) []SearchHit {
	out := make([]SearchHit, 0, len(hits))
	for _, h := range hits {
		if h.Metadata["tenant"] == tenant {
			out = append(out, h)
		}
	}
	return out
}
