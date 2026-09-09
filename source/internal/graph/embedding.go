package graph

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdd "github.com/networkteam/sdd/application"
)

// openAIProvider is the only embedding transport the viewer speaks: the
// `POST {endpoint}/v1/embeddings` contract OpenAI and its compatible
// self-hosted services answer. A repository naming another provider searches
// over terms instead.
const openAIProvider = "openai"

const (
	defaultOpenAIEndpoint   = "https://api.openai.com"
	defaultOpenAIBatchSize  = 100
	defaultEmbeddingTimeout = 2 * time.Minute
)

var (
	errEmbeddingUnconfigured = errors.New("the repository configures no embedding provider this viewer speaks")
	errEmbeddingRefused      = errors.New("the embedding endpoint refused the request")
	errEmbeddingShort        = errors.New("the embedding endpoint answered with the wrong number of vectors")
)

// embeddingConfig carries the embedding settings a repository declares. sdd
// keeps its own configuration types unexported, so the shape is declared here
// and holds for the version named. Measured against sdd 0.16.2.
type embeddingConfig struct {
	Provider         string            `yaml:"provider"`
	Model            string            `yaml:"model"`
	Endpoint         string            `yaml:"endpoint"`
	APIKeys          map[string]string `yaml:"api_keys"`
	Timeout          string            `yaml:"timeout"`
	Dimensions       int               `yaml:"dimensions"`
	BatchSize        int               `yaml:"batch_size"`
	QueryTemplate    string            `yaml:"query_template"`
	DocumentTemplate string            `yaml:"document_template"`
}

// fingerprint reproduces the string sdd derives for this embedder. It decides
// which stored index a search reads, so a viewer computing it differently
// builds a second index beside the one sdd's own tool filled, and re-embeds
// every entry to do it.
func (c embeddingConfig) fingerprint() string {
	base := openAIProvider + "/" + c.Model
	if c.Dimensions > 0 {
		base = fmt.Sprintf("%s/%d", base, c.Dimensions)
	}
	if c.DocumentTemplate == "" {
		return base
	}
	sum := sha256.Sum256([]byte(c.DocumentTemplate))
	return base + "/d:" + hex.EncodeToString(sum[:])[:6]
}

func (c embeddingConfig) batchSize() int {
	if c.BatchSize > 0 {
		return c.BatchSize
	}
	return defaultOpenAIBatchSize
}

func (c embeddingConfig) timeout() time.Duration {
	if c.Timeout == "" {
		return defaultEmbeddingTimeout
	}
	parsed, err := time.ParseDuration(c.Timeout)
	if err != nil || parsed <= 0 {
		return defaultEmbeddingTimeout
	}
	return parsed
}

func (c embeddingConfig) template(purpose sdd.EmbeddingPurpose) string {
	if purpose == sdd.EmbeddingQuery {
		return c.QueryTemplate
	}
	return c.DocumentTemplate
}

// mergeEmbeddingConfig overlays what a local configuration declares onto the
// shared one, field by field, which is how a machine's own endpoint and key sit
// beside a committed model choice.
func mergeEmbeddingConfig(base, overlay embeddingConfig) embeddingConfig {
	merged := base
	if overlay.Provider != "" {
		merged.Provider = overlay.Provider
	}
	if overlay.Model != "" {
		merged.Model = overlay.Model
	}
	if overlay.Endpoint != "" {
		merged.Endpoint = overlay.Endpoint
	}
	if len(overlay.APIKeys) > 0 {
		merged.APIKeys = overlay.APIKeys
	}
	if overlay.Timeout != "" {
		merged.Timeout = overlay.Timeout
	}
	if overlay.Dimensions > 0 {
		merged.Dimensions = overlay.Dimensions
	}
	if overlay.BatchSize > 0 {
		merged.BatchSize = overlay.BatchSize
	}
	if overlay.QueryTemplate != "" {
		merged.QueryTemplate = overlay.QueryTemplate
	}
	if overlay.DocumentTemplate != "" {
		merged.DocumentTemplate = overlay.DocumentTemplate
	}
	return merged
}

// embedder answers sdd's EmbeddingExecutor port over an OpenAI-compatible
// service.
type embedder struct {
	config embeddingConfig
	client *http.Client
}

func newEmbedder(config embeddingConfig) (*embedder, error) {
	if config.Provider != openAIProvider || config.Model == "" {
		return nil, errEmbeddingUnconfigured
	}
	if config.APIKeys[openAIProvider] == "" {
		return nil, fmt.Errorf("%s: %w", "embedding.api_keys.openai names no key", errEmbeddingUnconfigured)
	}
	if config.Endpoint == "" {
		config.Endpoint = defaultOpenAIEndpoint
	}
	return &embedder{config: config, client: &http.Client{Timeout: config.timeout()}}, nil
}

func (e *embedder) Spec(context.Context) (sdd.EmbeddingSpec, error) {
	return sdd.EmbeddingSpec{Fingerprint: e.config.fingerprint()}, nil
}

func (e *embedder) Embed(ctx context.Context, inputs []sdd.EmbeddingInput) ([]sdd.EmbeddingVector, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	vectors := make([]sdd.EmbeddingVector, 0, len(inputs))
	size := e.config.batchSize()
	for start := 0; start < len(inputs); start += size {
		end := min(start+size, len(inputs))
		batch := inputs[start:end]

		texts := make([]string, len(batch))
		for index, input := range batch {
			texts[index] = applyTemplate(e.config.template(input.Purpose), input.Text)
		}

		values, err := e.embedBatch(ctx, texts)
		if err != nil {
			return nil, fmt.Errorf("embedding inputs %d to %d: %w", start, end, err)
		}
		for index, input := range batch {
			vectors = append(vectors, sdd.EmbeddingVector{ID: input.ID, Values: values[index]})
		}
	}
	return vectors, nil
}

type embedRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	Dimensions     int      `json:"dimensions,omitempty"`
	EncodingFormat string   `json:"encoding_format"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (e *embedder) embedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	body, err := json.Marshal(embedRequest{
		Input:          texts,
		Model:          e.config.Model,
		Dimensions:     e.config.Dimensions,
		EncodingFormat: "float",
	})
	if err != nil {
		return nil, fmt.Errorf("preparing the embedding request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, e.config.Endpoint+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("preparing the embedding request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+e.config.APIKeys[openAIProvider])

	response, err := e.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("reaching the embedding endpoint: %w", err)
	}
	defer func() {
		err = errors.Join(err, response.Body.Close())
	}()

	var decoded embedResponse
	err = json.NewDecoder(response.Body).Decode(&decoded)
	if err != nil {
		return nil, fmt.Errorf("reading the embedding answer (status %d): %w", response.StatusCode, err)
	}
	if response.StatusCode != http.StatusOK {
		if decoded.Error != nil {
			return nil, fmt.Errorf("%w with status %d: %s", errEmbeddingRefused, response.StatusCode, decoded.Error.Message)
		}
		return nil, fmt.Errorf("%w with status %d", errEmbeddingRefused, response.StatusCode)
	}
	if len(decoded.Data) != len(texts) {
		return nil, fmt.Errorf("%w: %d for %d inputs", errEmbeddingShort, len(decoded.Data), len(texts))
	}

	values := make([][]float32, len(decoded.Data))
	for index, item := range decoded.Data {
		values[index] = item.Embedding
	}
	return values, nil
}

func applyTemplate(template, text string) string {
	if template == "" {
		return text
	}
	return strings.ReplaceAll(template, "{text}", text)
}

// indexCacheRoot is where sdd's own tool keeps every stored index, so a viewer
// reading it finds what that tool already built rather than filling a second
// copy.
func indexCacheRoot() (string, error) {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving the cache directory: %w", err)
		}
		base = filepath.Join(home, ".cache")
	}
	return filepath.Join(base, "sdd"), nil
}

// indexRepoKey names the stored index of one repository. A repository
// declaring an identity is keyed by it, and one declaring none by a digest of
// where it sits.
func indexRepoKey(repoID, repositoryPath string) string {
	if repoID != "" {
		return repoID
	}
	sum := sha256.Sum256([]byte(repositoryPath))
	return "local/" + hex.EncodeToString(sum[:])[:12]
}
