package client

type NotaryClient struct {
	trustDir   string
	serverURL  string
	rootCAFile string
}

type TUFCustom struct {
	TargetFormat string `json:"targetFormat"`
	Uri          string `json:"uri"`
}

type OSTreeCustom struct {
	TUFCustom

	Url string `json:"ostree"`
}

type DockerComposeCustom struct {
	TUFCustom

	TgzUrl       string            `json:"tgz"`
	ComposeFiles []string          `json:"compose-files,omitempty"`
	ComposeEnv   map[string]string `json:"compose-env,omitempty"`
}
