package manifest

type IToolReleaseFile struct {
	Filename        string  `json:"filename"`
	Platform        string  `json:"platform"`
	PlatformVersion *string `json:"platform_version,omitempty"`
	Arch            string  `json:"arch"`
	DownloadURL     string  `json:"download_url"`
}

type IToolRelease struct {
	Version    string             `json:"version"`
	Stable     bool               `json:"stable"`
	ReleaseURL string             `json:"release_url"`
	Files      []IToolReleaseFile `json:"files"`
}

func InternalFindMatch(versionSpec string, stable bool, candidates []IToolRelease, archFilter string) (*IToolRelease, error) {
	
}
