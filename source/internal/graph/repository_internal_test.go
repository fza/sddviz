package graph

import "testing"

func TestReadRepositoryConfigLayersTheMachineFileBeneath(t *testing.T) {
	cases := []struct {
		name           string
		repository     string
		userConfigPath string
		model          string
		endpoint       string
		key            string
	}{
		{
			name:           "a repository configuring no embedding takes the machine's",
			repository:     "testdata/repository",
			userConfigPath: "testdata/userconfig/config.yaml",
			model:          "machine-model",
			endpoint:       "http://127.0.0.1:1234",
			key:            "machine-key",
		},
		{
			name:           "a local file overrides the machine's field by field",
			repository:     "testdata/layered",
			userConfigPath: "testdata/userconfig/config.yaml",
			model:          "repository-model",
			endpoint:       "http://127.0.0.1:1234",
			key:            "machine-key",
		},
		{
			name:           "a shared file cannot send the machine's key elsewhere",
			repository:     "testdata/hostile",
			userConfigPath: "testdata/userconfig/config.yaml",
			model:          "shared-model",
			endpoint:       "http://127.0.0.1:1234",
			key:            "machine-key",
		},
		{
			name:       "a shared file cannot name a key or endpoint on its own",
			repository: "testdata/hostile",
			model:      "shared-model",
		},
		{
			name:           "a machine file that is absent is no failure",
			repository:     "testdata/repository",
			userConfigPath: "testdata/userconfig/absent.yaml",
		},
		{
			name:       "no machine file at all reads the repository alone",
			repository: "testdata/repository",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			config, err := readRepositoryConfig(testCase.repository, testCase.userConfigPath)
			if err != nil {
				t.Fatalf("readRepositoryConfig: %v", err)
			}
			if config.Embedding.Model != testCase.model {
				t.Errorf("the embedding model reads %q, expected the layer closest to the repository", config.Embedding.Model)
			}
			if config.Embedding.Endpoint != testCase.endpoint {
				t.Errorf("the endpoint reads %q, expected one a machine-owned layer supplies", config.Embedding.Endpoint)
			}
			if config.Embedding.APIKeys[openAIProvider] != testCase.key {
				t.Error("the key differs from the one a machine-owned layer supplies")
			}
		})
	}
}

func TestUserConfigPathFollowsTheConfigEnvironment(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/elsewhere")

	if UserConfigPath() != "/tmp/elsewhere/sdd/config.yaml" {
		t.Errorf("the machine file resolves to %q, expected the one sdd reads", UserConfigPath())
	}
}
