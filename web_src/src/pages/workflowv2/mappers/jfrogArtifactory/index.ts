import { ComponentBaseMapper, TriggerRenderer, EventStateRegistry } from "../types";
import { jfrogArtifactoryBaseMapper } from "./base";
import { DEFAULT_STATE_REGISTRY } from "../stateRegistry";

export const componentMappers: Record<string, ComponentBaseMapper> = {
  getArtifactInfo: jfrogArtifactoryBaseMapper,
  uploadArtifact: jfrogArtifactoryBaseMapper,
  deleteArtifact: jfrogArtifactoryBaseMapper,
};

export const triggerRenderers: Record<string, TriggerRenderer> = {};

export const eventStateRegistry: Record<string, EventStateRegistry> = {
  getArtifactInfo: DEFAULT_STATE_REGISTRY,
  uploadArtifact: DEFAULT_STATE_REGISTRY,
  deleteArtifact: DEFAULT_STATE_REGISTRY,
};
