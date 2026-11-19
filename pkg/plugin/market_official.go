package plugin

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	OfficialIndexDomain     = "certimate:plugin-market:index:v1\x00"
	OfficialIndexPayloadMax = 2 << 20
	OfficialIndexMaxBytes   = 3 << 20
	OfficialZIPMaxBytes     = 20_000_000
	officialMaxIndexAge     = 7 * 24 * time.Hour
	officialMaxClockSkew    = 5 * time.Minute
)

var (
	ErrOfficialIndexExpired = errors.New("market: official index signature has expired")
	officialKeyIDPattern    = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,63}$`)
	officialVersionPattern  = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
	officialCommitPattern   = regexp.MustCompile(`^[0-9a-f]{40}$`)
	officialDigestPattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	officialPlatforms       = []string{"darwin/amd64", "darwin/arm64", "linux/amd64", "linux/arm", "linux/arm64", "windows/386", "windows/amd64", "windows/arm64"}
)

type OfficialMarketTrust struct {
	Epoch uint32
	Keys  map[string]ed25519.PublicKey
}

type OfficialIndexState struct {
	Epoch      uint32
	Generation uint64
	Digest     [sha256.Size]byte
}

type OfficialMarketIndex struct {
	State     OfficialIndexState
	ExpiresAt time.Time
	Plugins   []*MarketManifest
}

type OfficialMarketRelease struct {
	Repository    string
	ReleaseCommit string
	Icon          *OfficialMarketIcon
	Artifacts     map[string]OfficialMarketArtifact
}

type OfficialMarketIcon struct {
	SHA256 string `json:"sha256"`
	Format string `json:"format"`
}

type OfficialMarketArtifact struct {
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type officialMarketSummary struct {
	ProviderType    string                       `json:"provider_type"`
	Version         string                       `json:"version"`
	Capability      string                       `json:"capability"`
	Category        string                       `json:"category"`
	DisplayNameKey  string                       `json:"display_name_key"`
	Description     string                       `json:"description"`
	ProtocolVersion uint32                       `json:"protocol_version"`
	MinCoreVersion  string                       `json:"min_core_version"`
	MaxCoreVersion  string                       `json:"max_core_version"`
	Repository      string                       `json:"repository"`
	ReleaseCommit   string                       `json:"release_commit"`
	Platforms       []string                     `json:"platforms"`
	I18n            map[string]MarketTranslation `json:"i18n"`
	Detail          struct {
		SHA256 string `json:"sha256"`
	} `json:"detail"`
	Icon *OfficialMarketIcon `json:"icon"`
}

func VerifyOfficialMarketIndex(raw []byte, trust OfficialMarketTrust, floor OfficialIndexState, now time.Time) (*OfficialMarketIndex, error) {
	if len(raw) == 0 || len(raw) > OfficialIndexMaxBytes {
		return nil, errors.New("market: invalid signed index size")
	}
	var envelope struct {
		KeyID     string `json:"key_id"`
		Payload   []byte `json:"payload"`
		Signature []byte `json:"signature"`
	}
	if err := decodeOfficialJSON(raw, &envelope); err != nil {
		return nil, fmt.Errorf("market: decode signed index envelope: %w", err)
	}
	key := trust.Keys[envelope.KeyID]
	if !officialKeyIDPattern.MatchString(envelope.KeyID) || len(key) != ed25519.PublicKeySize {
		return nil, errors.New("market: official index key is not trusted")
	}
	if len(envelope.Payload) == 0 || len(envelope.Payload) > OfficialIndexPayloadMax || len(envelope.Signature) != ed25519.SignatureSize ||
		!ed25519.Verify(key, append([]byte(OfficialIndexDomain), envelope.Payload...), envelope.Signature) {
		return nil, errors.New("market: invalid official index signature")
	}

	var payload struct {
		FormatVersion uint32                   `json:"format_version"`
		Epoch         uint32                   `json:"epoch"`
		Generation    uint64                   `json:"generation"`
		IssuedAt      time.Time                `json:"issued_at"`
		ExpiresAt     time.Time                `json:"expires_at"`
		SourceCommit  string                   `json:"source_commit"`
		Plugins       []*officialMarketSummary `json:"plugins"`
	}
	if err := decodeOfficialJSON(envelope.Payload, &payload); err != nil {
		return nil, fmt.Errorf("market: decode official index payload: %w", err)
	}
	if payload.FormatVersion != 1 || payload.Epoch == 0 || payload.Epoch != trust.Epoch ||
		payload.Generation == 0 || !officialCommitPattern.MatchString(payload.SourceCommit) || payload.Plugins == nil {
		return nil, errors.New("market: invalid official index format, trust epoch or source commit")
	}
	if payload.IssuedAt.Year() < 2000 || payload.IssuedAt.After(now.Add(officialMaxClockSkew)) ||
		!payload.ExpiresAt.After(payload.IssuedAt) || payload.ExpiresAt.Sub(payload.IssuedAt) > officialMaxIndexAge {
		return nil, errors.New("market: invalid official index validity window")
	}
	state := OfficialIndexState{Epoch: payload.Epoch, Generation: payload.Generation, Digest: sha256.Sum256(envelope.Payload)}
	if floor.Epoch > state.Epoch || (floor.Epoch == state.Epoch &&
		(state.Generation < floor.Generation || (state.Generation == floor.Generation && state.Digest != floor.Digest))) {
		return nil, errors.New("market: official index generation rollback or conflict")
	}
	result := &OfficialMarketIndex{State: state, ExpiresAt: payload.ExpiresAt, Plugins: make([]*MarketManifest, 0, len(payload.Plugins))}
	seen := make(map[string]bool, len(payload.Plugins))
	for _, summary := range payload.Plugins {
		if summary == nil || seen[summary.ProviderType] {
			return nil, errors.New("market: duplicate or missing official index entry")
		}
		seen[summary.ProviderType] = true
		manifest, err := summary.marketManifest()
		if err != nil {
			return nil, err
		}
		result.Plugins = append(result.Plugins, manifest)
	}
	if !now.Before(result.ExpiresAt) {
		return result, ErrOfficialIndexExpired
	}
	return result, nil
}

func (summary officialMarketSummary) marketManifest() (*MarketManifest, error) {
	platforms := slices.Clone(summary.Platforms)
	slices.Sort(platforms)
	if ValidateProviderType(summary.ProviderType) != nil || !validOfficialVersion(summary.Version) ||
		!IsValidCapability(summary.Capability) || summary.ProtocolVersion == 0 ||
		!validOfficialRepo(summary.Repository) || !officialCommitPattern.MatchString(summary.ReleaseCommit) ||
		!officialDigestPattern.MatchString(summary.Detail.SHA256) || !slices.Equal(platforms, officialPlatforms) ||
		!validOfficialIcon(summary.Icon) || !validOfficialBounds(summary.MinCoreVersion, summary.MaxCoreVersion) {
		return nil, fmt.Errorf("market: invalid official summary for %q", summary.ProviderType)
	}
	return &MarketManifest{
		Manifest: Manifest{
			ProviderType: summary.ProviderType, Version: summary.Version, Capability: summary.Capability, Category: summary.Category,
			DisplayNameKey: summary.DisplayNameKey, Description: summary.Description, ProtocolVersion: summary.ProtocolVersion,
			MinCoreVersion: summary.MinCoreVersion, MaxCoreVersion: summary.MaxCoreVersion,
		},
		Detail:   &MarketDetailRef{SHA256: summary.Detail.SHA256, Repo: summary.Repository, Platforms: platforms},
		I18n:     summary.I18n,
		Official: &OfficialMarketRelease{Repository: summary.Repository, ReleaseCommit: summary.ReleaseCommit, Icon: summary.Icon},
	}, nil
}

func ParseOfficialMarketDetail(data []byte, summary *MarketManifest) (*MarketManifest, error) {
	if summary == nil || summary.Detail == nil || summary.Official == nil || len(data) == 0 || len(data) > 64<<10 ||
		fmt.Sprintf("%x", sha256.Sum256(data)) != summary.Detail.SHA256 {
		return nil, errors.New("market: invalid official detail checksum or size")
	}
	var detail struct {
		ProviderType  string                   `json:"provider_type"`
		Repository    string                   `json:"repository"`
		Version       string                   `json:"version"`
		Tag           string                   `json:"tag"`
		SourceCommit  string                   `json:"source_commit"`
		BuildRunID    string                   `json:"build_run_id"`
		Artifacts     []OfficialMarketArtifact `json:"artifacts"`
		ReleaseCommit string                   `json:"release_commit"`
		Manifest      Manifest                 `json:"manifest"`
		Icon          *OfficialMarketIcon      `json:"icon"`
	}
	if err := decodeOfficialJSON(data, &detail); err != nil {
		return nil, fmt.Errorf("market: decode official detail: %w", err)
	}
	manifest := detail.Manifest
	official := summary.Official
	sameIcon := (detail.Icon == nil && official.Icon == nil) ||
		(detail.Icon != nil && official.Icon != nil && *detail.Icon == *official.Icon)
	runID, runErr := strconv.ParseUint(detail.BuildRunID, 10, 64)
	if detail.ProviderType != summary.ProviderType || manifest.ProviderType != summary.ProviderType ||
		detail.Version != summary.Version || manifest.Version != summary.Version || detail.Tag != "v"+summary.Version ||
		detail.Repository != official.Repository || detail.ReleaseCommit != official.ReleaseCommit ||
		!officialCommitPattern.MatchString(detail.SourceCommit) || detail.SourceCommit == detail.ReleaseCommit ||
		runErr != nil || runID == 0 || strconv.FormatUint(runID, 10) != detail.BuildRunID ||
		manifest.ProtocolVersion != summary.ProtocolVersion || manifest.Capability != summary.Capability ||
		manifest.Category != summary.Category || manifest.DisplayNameKey != summary.DisplayNameKey || manifest.Description != summary.Description ||
		manifest.MinCoreVersion != summary.MinCoreVersion || manifest.MaxCoreVersion != summary.MaxCoreVersion || !sameIcon ||
		manifest.OS != "" || manifest.Arch != "" || manifest.SHA256 != "" || manifest.Validate() != nil ||
		ValidateProviderType(manifest.AccessProviderType) != nil || ValidateBinaryName(manifest.Binary) != nil {
		return nil, fmt.Errorf("market: official detail does not match index entry for %q", summary.ProviderType)
	}
	artifacts := make(map[string]OfficialMarketArtifact, len(detail.Artifacts))
	for _, artifact := range detail.Artifacts {
		platform := AssetKey(artifact.OS, artifact.Arch)
		if !slices.Contains(summary.Detail.Platforms, platform) || artifacts[platform].Path != "" ||
			artifact.Path != "build/"+artifact.OS+"_"+artifact.Arch+".zip" ||
			artifact.Size <= 0 || artifact.Size > OfficialZIPMaxBytes || !officialDigestPattern.MatchString(artifact.SHA256) {
			return nil, fmt.Errorf("market: invalid or duplicate official artifact for %q", platform)
		}
		artifacts[platform] = artifact
	}
	if len(artifacts) != len(summary.Detail.Platforms) {
		return nil, errors.New("market: official artifact platform matrix does not match index")
	}
	release := *official
	release.Artifacts = artifacts
	return &MarketManifest{Manifest: manifest, Detail: summary.Detail, I18n: summary.I18n, Official: &release}, nil
}

func validOfficialVersion(value string) bool {
	if len(value) > 64 || !officialVersionPattern.MatchString(value) {
		return false
	}
	for part := range strings.SplitSeq(value, ".") {
		if _, err := strconv.ParseInt(part, 10, 64); err != nil {
			return false
		}
	}
	return true
}

func validOfficialBounds(minimum, maximum string) bool {
	return (minimum == "" || validOfficialVersion(minimum)) && (maximum == "" || validOfficialVersion(maximum)) &&
		(minimum == "" || maximum == "" || !semverLess(maximum, minimum))
}

func validOfficialRepo(repository string) bool {
	name, ok := strings.CutPrefix(repository, "certimate-go/plugin-provider-")
	return ok && ValidateProviderType(name) == nil
}

func validOfficialIcon(icon *OfficialMarketIcon) bool {
	return icon == nil || (officialDigestPattern.MatchString(icon.SHA256) && (icon.Format == "png" || icon.Format == "jpeg"))
}

func decodeOfficialJSON(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return errors.New("market: unexpected data after JSON value")
	}
	return nil
}
