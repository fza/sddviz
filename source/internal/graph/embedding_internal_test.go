package graph

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networkteam/sdd/pkg/llm/embed"
	"github.com/networkteam/sdd/pkg/sddtest"
)

// The fingerprint decides which stored index a search reads. This holds it to
// the string sdd's own tool derives, because a viewer computing it differently
// silently builds a second index and re-embeds every entry to fill it.
func TestFingerprintMatchesTheOneSddDerives(t *testing.T) {
	cases := []struct {
		name     string
		config   embeddingConfig
		expected string
	}{
		{
			name: "a model with a document template",
			config: embeddingConfig{
				Model:            "lmstudio-community/text-embedding-embeddinggemma-300m-qat",
				DocumentTemplate: "title: none | text: {text}",
			},
			expected: "openai/lmstudio-community/text-embedding-embeddinggemma-300m-qat/d:455438",
		},
		{
			name:     "a model with no template at all",
			config:   embeddingConfig{Model: "text-embedding-3-small"},
			expected: "openai/text-embedding-3-small",
		},
		{
			name:     "a model whose dimensions are pinned",
			config:   embeddingConfig{Model: "text-embedding-3-small", Dimensions: 512},
			expected: "openai/text-embedding-3-small/512",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.config.fingerprint() != testCase.expected {
				t.Errorf("the fingerprint reads %q, expected the one sdd derives for the same settings", testCase.config.fingerprint())
			}
		})
	}
}

func TestMergeEmbeddingConfigLetsTheLocalFileWin(t *testing.T) {
	base := embeddingConfig{Provider: "openai", Model: "shared", Timeout: "30s"}
	overlay := embeddingConfig{Model: "local", Endpoint: "http://127.0.0.1:1234"}

	merged := mergeEmbeddingConfig(base, overlay)

	if merged.Model != "local" {
		t.Errorf("the merged model reads %q, expected the local file's", merged.Model)
	}
	if merged.Endpoint != "http://127.0.0.1:1234" {
		t.Errorf("the merged endpoint reads %q", merged.Endpoint)
	}
	if merged.Provider != "openai" || merged.Timeout != "30s" {
		t.Error("a field the local file leaves out lost the shared file's value")
	}
}

func TestNewEmbedderRefusesWhatItCannotSpeak(t *testing.T) {
	cases := []struct {
		name   string
		config embeddingConfig
	}{
		{name: "no provider at all", config: embeddingConfig{}},
		{name: "a provider the viewer does not speak", config: embeddingConfig{Provider: "ollama", Model: "nomic"}},
		{name: "no model", config: embeddingConfig{Provider: "openai"}},
		{name: "no key", config: embeddingConfig{Provider: "openai", Model: "a-model"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := newEmbedder(testCase.config)
			if err == nil {
				t.Fatal("the viewer built an embedder from settings it cannot use")
			}
		})
	}
}

func TestIndexRepoKeyFollowsTheDeclaredIdentity(t *testing.T) {
	declared := indexRepoKey("gitlab.com/format-d/devops/format-d-fdbox", "/anywhere")
	if declared != "gitlab.com/format-d/devops/format-d-fdbox" {
		t.Errorf("a repository declaring an identity is keyed by %q", declared)
	}

	local := indexRepoKey("", "/anywhere")
	if len(local) != len("local/")+12 || local[:6] != "local/" {
		t.Errorf("a repository declaring no identity is keyed by %q", local)
	}
	if indexRepoKey("", "/elsewhere") == local {
		t.Error("two repositories at different paths share one key")
	}
}

func TestIndexCacheRootFollowsTheCacheEnvironment(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "/tmp/somewhere")

	root, err := indexCacheRoot()
	if err != nil {
		t.Fatalf("indexCacheRoot: %v", err)
	}
	if root != "/tmp/somewhere/sdd" {
		t.Errorf("the cache root reads %q, expected the one sdd's own tool uses", root)
	}
}

// embeddingEndpoint answers the OpenAI embeddings contract with one short
// vector per input, and records every text it received in seen.
func embeddingEndpoint(t *testing.T, seen *[]string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/embeddings" {
			t.Errorf("the request reached %q", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer a-key" {
			t.Error("the request carried no key")
		}

		var decoded embedRequest
		err := json.NewDecoder(request.Body).Decode(&decoded)
		if err != nil {
			t.Errorf("decoding the request: %v", err)
			return
		}
		*seen = append(*seen, decoded.Input...)

		answer := embedResponse{}
		for range decoded.Input {
			answer.Data = append(answer.Data, struct {
				Embedding []float32 `json:"embedding"`
			}{Embedding: []float32{0.5, 0.25}})
		}
		writer.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(writer).Encode(answer)
		if err != nil {
			t.Errorf("answering: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func TestEmbedderConformsToSddsPort(t *testing.T) {
	sddtest.RunEmbedderTests(t, func(t *testing.T) sddtest.EmbedderFixture {
		var seen []string
		server := embeddingEndpoint(t, &seen)
		built, err := newEmbedder(embeddingConfig{
			Provider:  openAIProvider,
			Model:     "a-model",
			Endpoint:  server.URL,
			APIKeys:   map[string]string{openAIProvider: "a-key"},
			BatchSize: 2,
		})
		if err != nil {
			t.Fatalf("newEmbedder: %v", err)
		}
		return sddtest.EmbedderFixture{Embedder: built, Texts: []string{"alpha", "beta", "gamma"}}
	})
}

func TestEmbedAppliesTheTemplateOfEachPurpose(t *testing.T) {
	var seen []string
	server := embeddingEndpoint(t, &seen)

	built, err := newEmbedder(embeddingConfig{
		Provider:         openAIProvider,
		Model:            "a-model",
		Endpoint:         server.URL,
		APIKeys:          map[string]string{openAIProvider: "a-key"},
		QueryTemplate:    "query: {text}",
		DocumentTemplate: "document: {text}",
		BatchSize:        1,
	})
	if err != nil {
		t.Fatalf("newEmbedder: %v", err)
	}

	documents, err := built.Embed(t.Context(), embed.Request{Purpose: embed.PurposeDocument, Texts: []string{"alpha", "gamma"}})
	if err != nil {
		t.Fatalf("Embed, documents: %v", err)
	}
	query, err := built.Embed(t.Context(), embed.Request{Purpose: embed.PurposeQuery, Texts: []string{"beta"}})
	if err != nil {
		t.Fatalf("Embed, query: %v", err)
	}

	expected := []string{"document: alpha", "document: gamma", "query: beta"}
	if len(seen) != len(expected) {
		t.Fatalf("the endpoint received %d texts, expected one per input", len(seen))
	}
	for index, want := range expected {
		if seen[index] != want {
			t.Errorf("the endpoint received %q, expected the templated text for that purpose", seen[index])
		}
	}

	if len(documents.Vectors) != 2 || len(query.Vectors) != 1 {
		t.Errorf("the calls returned %d and %d vectors, expected one per text across batches", len(documents.Vectors), len(query.Vectors))
	}
	if documents.Identity.Model != "a-model" || documents.Identity.Provider != openAIProvider {
		t.Errorf("the result names %+v rather than what served the call", documents.Identity)
	}
}

func TestEmbedReportsWhatTheEndpointRefused(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusUnauthorized)
		_, err := writer.Write([]byte(`{"error":{"message":"no key"}}`))
		if err != nil {
			t.Fatalf("answering: %v", err)
		}
	}))
	defer server.Close()

	built, err := newEmbedder(embeddingConfig{
		Provider: openAIProvider,
		Model:    "a-model",
		Endpoint: server.URL,
		APIKeys:  map[string]string{openAIProvider: "a-key"},
	})
	if err != nil {
		t.Fatalf("newEmbedder: %v", err)
	}

	_, err = built.Embed(t.Context(), embed.Request{Purpose: embed.PurposeDocument, Texts: []string{"alpha"}})
	if err == nil {
		t.Fatal("a refused embedding request reported success")
	}
}

func TestChooseSearchModeFollowsTheWordCount(t *testing.T) {
	cases := []struct {
		name      string
		available string
		words     []string
		expected  string
	}{
		{
			name:      "one word where meaning is available",
			available: vectorMode,
			words:     []string{"gateway"},
			expected:  textMode,
		},
		{
			name:      "several words where meaning is available",
			available: vectorMode,
			words:     []string{"how", "a", "database", "opens"},
			expected:  vectorMode,
		},
		{
			name:      "several words where meaning is not available",
			available: textMode,
			words:     []string{"how", "a", "database", "opens"},
			expected:  textMode,
		},
		{
			name:      "no words at all",
			available: vectorMode,
			words:     nil,
			expected:  textMode,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			chosen := chooseSearchMode(testCase.available, testCase.words)
			if chosen != testCase.expected {
				t.Errorf("the search would run in mode %q", chosen)
			}
		})
	}
}
