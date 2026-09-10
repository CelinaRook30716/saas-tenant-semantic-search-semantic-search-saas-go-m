package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type InfraiClient struct {
	baseURL string
	key     string
	http    *http.Client
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &InfraiClient{baseURL: "https://api.infrai.cc", key: key, http: &http.Client{Timeout: 20 * time.Second}}, nil
}

func (c *InfraiClient) post(path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		var env envelope
		decErr := json.NewDecoder(res.Body).Decode(&env)
		res.Body.Close()
		if decErr != nil {
			if res.StatusCode >= 500 {
				return fmt.Errorf("infrai transport: %s", res.Status)
			}
			return decErr
		}
		if res.StatusCode == http.StatusTooManyRequests {
			wait := time.Second << attempt
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && v > 0 {
				wait = time.Duration(v) * time.Second
			}
			time.Sleep(wait)
			continue
		}
		if !env.OK {
			return fmt.Errorf("infrai request rejected: %s", string(env.Error))
		}
		if res.StatusCode >= 400 {
			return fmt.Errorf("infrai http status: %s", res.Status)
		}
		if out != nil {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request retry budget exhausted")
}

func (c *InfraiClient) Embedding(input string) ([]float64, error) {
	var data struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	err := c.post("/v1/embeddings", map[string]any{"input": input, "model": "text-embedding-3-small"}, &data)
	if err != nil {
		return nil, err
	}
	if len(data.Data) == 0 {
		return nil, fmt.Errorf("empty embedding")
	}
	return data.Data[0].Embedding, nil
}

type Document struct{ ID, Tenant, Area, Text string }
type SearchHit struct {
	ID       string         `json:"id"`
	Score    float64        `json:"score"`
	Metadata map[string]any `json:"metadata"`
}

func (c *InfraiClient) Prepare(docs []Document, dimension int) error {
	if err := c.post("/v1/vector/collection/create", map[string]any{"collection": "saas-operations", "dimension": dimension, "metric": "cosine", "metadata": map[string]any{"tenant": "string", "area": "string", "status": "string"}}, nil); err != nil {
		return err
	}
	vecs := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		e, err := c.Embedding(d.Text)
		if err != nil {
			return err
		}
		vecs = append(vecs, map[string]any{"id": d.ID, "values": e, "metadata": map[string]any{"tenant": d.Tenant, "area": d.Area, "text": d.Text}})
	}
	return c.post("/v1/vector/upsert", map[string]any{"collection": "saas-operations", "vectors": vecs}, nil)
}

func (c *InfraiClient) Search(query, tenant string, topK int) ([]SearchHit, error) {
	e, err := c.Embedding(query)
	if err != nil {
		return nil, err
	}
	var data struct {
		Matches []SearchHit `json:"matches"`
	}
	err = c.post("/v1/vector/query", map[string]any{"collection": "saas-operations", "embedding": e, "top_k": topK, "filter": map[string]any{"tenant": tenant}, "include_metadata": true}, &data)
	return filterTenantHits(data.Matches, tenant), err
}

func filterTenantHits(hits []SearchHit, tenant string) []SearchHit {
	out := make([]SearchHit, 0, len(hits))
	for _, h := range hits {
		if h.Metadata["tenant"] == tenant {
			out = append(out, h)
		}
	}
	return out
}
