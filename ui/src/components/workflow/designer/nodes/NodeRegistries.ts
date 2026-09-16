import { type FlowNodeType } from "@flowgram.ai/fixed-layout-editor";

import { BizApplyNodeRegistry } from "./BizApplyNodeRegistry";
import { BizDeployNodeRegistry } from "./BizDeployNodeRegistry";
import { BizMonitorNodeRegistry } from "./BizMonitorNodeRegistry";
import { BizNotifyNodeRegistry } from "./BizNotifyNodeRegistry";
import { BizUploadNodeRegistry } from "./BizUploadNodeRegistry";
import { BranchBlockNodeRegistry, ConditionNodeRegistry } from "./ConditionNode";
import { DelayNodeRegistry } from "./DelayNode";
import { EndNodeRegistry } from "./EndNode";
import { StartNodeRegistry } from "./StartNode";
import { CatchBlockNodeRegistry, TryCatchNodeRegistry } from "./TryCatchNode";

const nodeRegistries = [
  StartNodeRegistry,
  EndNodeRegistry,
  DelayNodeRegistry,
  BizApplyNodeRegistry,
  BizUploadNodeRegistry,
  BizMonitorNodeRegistry,
  BizDeployNodeRegistry,
  BizNotifyNodeRegistry,
  ConditionNodeRegistry,
  BranchBlockNodeRegistry,
  TryCatchNodeRegistry,
  CatchBlockNodeRegistry,
];

export const getAllNodeRegistries = () => {
  return [...nodeRegistries];
};

export const isNodeTypeRegistered = (type: FlowNodeType) => {
  return nodeRegistries.some((registry) => registry.type === type);
};
