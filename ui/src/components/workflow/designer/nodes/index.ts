import { BizApplyNodeRegistry } from "./BizApplyNodeRegistry";
import { BizDeployNodeRegistry } from "./BizDeployNodeRegistry";
import { BizMonitorNodeRegistry } from "./BizMonitorNodeRegistry";
import { BizNotifyNodeRegistry } from "./BizNotifyNodeRegistry";
import { BizPurgeNodeRegistry } from "./BizPurgeNodeRegistry";
import { BizUploadNodeRegistry } from "./BizUploadNodeRegistry";
import { BranchBlockNodeRegistry, ConditionNodeRegistry } from "./ConditionNode";
import { DelayNodeRegistry } from "./DelayNode";
import { EndNodeRegistry } from "./EndNode";
import { StartNodeRegistry } from "./StartNode";
import { CatchBlockNodeRegistry, TryCatchNodeRegistry } from "./TryCatchNode";

export const getAllNodeRegistries = () => {
  return [
    StartNodeRegistry,
    EndNodeRegistry,
    DelayNodeRegistry,
    BizApplyNodeRegistry,
    BizUploadNodeRegistry,
    BizMonitorNodeRegistry,
    BizDeployNodeRegistry,
    BizPurgeNodeRegistry,
    BizNotifyNodeRegistry,
    ConditionNodeRegistry,
    BranchBlockNodeRegistry,
    TryCatchNodeRegistry,
    CatchBlockNodeRegistry,
  ];
};

export type * from "./typings";
