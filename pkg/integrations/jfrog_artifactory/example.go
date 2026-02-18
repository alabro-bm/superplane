package jfrog_artifactory

import (
	_ "embed"
	"sync"

	"github.com/superplanehq/superplane/pkg/utils"
)

//go:embed example_output_get_artifact_info.json
var exampleOutputGetArtifactInfoBytes []byte

//go:embed example_output_upload_artifact.json
var exampleOutputUploadArtifactBytes []byte

//go:embed example_output_delete_artifact.json
var exampleOutputDeleteArtifactBytes []byte

var exampleOutputGetArtifactInfoOnce sync.Once
var exampleOutputGetArtifactInfo map[string]any

var exampleOutputUploadArtifactOnce sync.Once
var exampleOutputUploadArtifact map[string]any

var exampleOutputDeleteArtifactOnce sync.Once
var exampleOutputDeleteArtifact map[string]any

func (g *GetArtifactInfo) ExampleOutput() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleOutputGetArtifactInfoOnce, exampleOutputGetArtifactInfoBytes, &exampleOutputGetArtifactInfo)
}

func (u *UploadArtifact) ExampleOutput() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleOutputUploadArtifactOnce, exampleOutputUploadArtifactBytes, &exampleOutputUploadArtifact)
}

func (d *DeleteArtifact) ExampleOutput() map[string]any {
	return utils.UnmarshalEmbeddedJSON(&exampleOutputDeleteArtifactOnce, exampleOutputDeleteArtifactBytes, &exampleOutputDeleteArtifact)
}
