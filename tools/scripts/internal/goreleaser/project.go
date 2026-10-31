package goreleaser

const (
	mainPath      = "./cmd/battleserver"
	binaryName    = "battleserver"
	documentFiles = "LICENSE"
)

func newConfig(legacyGo string) Config {
	return Config{
		Version:     2,
		ProjectName: "ageBattleServer",
		Builds: []Build{
			build("battleserver-win7-386", "windows", "386", legacyGo, nil, nil),
			build("battleserver-win7-amd64", "windows", "amd64", legacyGo, nil, nil),
			build("battleserver-win10-386", "windows", "386", "", nil, nil),
			build("battleserver-win10-amd64", "windows", "amd64", "", nil, nil),
			build("battleserver-win10-arm64", "windows", "arm64", "", nil, nil),
			build("battleserver-linux-386", "linux", "386", "", nil, nil),
			build("battleserver-linux-amd64", "linux", "amd64", "", nil, nil),
			build("battleserver-linux-arm64", "linux", "arm64", "", nil, nil),
			build("battleserver-linux-arm", "linux", "arm", "", []string{"5", "6"}, nil),
			build("battleserver-darwin-amd64", "darwin", "amd64", "", nil, nil),
			build("battleserver-darwin-arm64", "darwin", "arm64", "", nil, []string{"v8.0"}),
		},
		Archives: []Archive{
			archive("battleserver-win7", []string{"battleserver-win7-386", "battleserver-win7-amd64"}, "zip", "{{ .ProjectName }}_{{ .RawVersion }}_win7_{{ if eq .Arch \"386\" }}x86-32{{ else if eq .Arch \"amd64\" }}x86-64{{ else }}{{ .Arch }}{{ end }}"),
			archive("battleserver-win10", []string{"battleserver-win10-386", "battleserver-win10-amd64", "battleserver-win10-arm64"}, "zip", "{{ .ProjectName }}_{{ .RawVersion }}_win10_{{ if eq .Arch \"386\" }}x86-32{{ else if eq .Arch \"amd64\" }}x86-64{{ else }}{{ .Arch }}{{ end }}"),
			archive("battleserver-linux-386", []string{"battleserver-linux-386"}, "tar.gz", "{{ .ProjectName }}_{{ .RawVersion }}_linux_x86-32"),
			archive("battleserver-linux-amd64", []string{"battleserver-linux-amd64"}, "tar.gz", "{{ .ProjectName }}_{{ .RawVersion }}_linux_x86-64"),
			archive("battleserver-linux-arm64", []string{"battleserver-linux-arm64"}, "tar.gz", "{{ .ProjectName }}_{{ .RawVersion }}_linux_arm64"),
			archive("battleserver-linux-arm", []string{"battleserver-linux-arm"}, "tar.gz", "{{ .ProjectName }}_{{ .RawVersion }}_linux_arm-{{ .Arm }}"),
			archive("battleserver-darwin-universal", []string{"battleserver-darwin-universal"}, "tar.gz", "{{ .ProjectName }}_{{ .RawVersion }}_mac"),
		},
		UniversalBinaries: []UniversalBinary{{
			ID: "battleserver-darwin-universal", IDs: []string{"battleserver-darwin-amd64", "battleserver-darwin-arm64"}, NameTemplate: binaryName,
		}},
		Checksum:  Checksum{NameTemplate: "{{ .ProjectName }}_{{ .RawVersion }}_checksums.txt"},
		Release:   Release{Draft: true, Prerelease: "auto"},
		Changelog: Changelog{Sort: "asc", Filters: Filters{Exclude: []string{"^docs:", "^test:"}}},
		Signs: []Sign{{Artifacts: "checksum", Cmd: "gpg", Args: []string{"--batch", "--pinentry-mode", "loopback", "--passphrase", "{{ .Env.GPG_PASSPHRASE }}", "--local-user", "{{ .Env.GPG_FINGERPRINT }}", "--output", "${signature}", "--detach-sign", "${artifact}"}}},
	}
}

func build(id, goos, goarch, tool string, goarm, goarm64 []string) Build {
	return Build{ID: id, Main: mainPath, Binary: binaryName, Goos: []string{goos}, Goarch: []string{goarch}, Tool: tool, Goarm: goarm, Goarm64: goarm64}
}

func archive(id string, ids []string, format, name string) Archive {
	return Archive{ID: id, IDs: ids, Formats: []string{format}, NameTemplate: name, Files: []string{documentFiles, "README.md", "DEVELOPMENT.md"}}
}
