import {
  type FixedLayoutPluginContext,
  type FlowNodeEntity,
  type FlowNodeJSON,
  type FlowNodeMeta,
  type FlowNodeRegistry,
  type FormMeta,
  type FormRenderProps,
} from "@flowgram.ai/fixed-layout-editor";

import { WORKFLOW_NODE_TYPES, type WorkflowNode } from "@/domain/workflow";

export enum NodeType {
  Start = "start",
  End = "end",
  Delay = "delay",
  Condition = "condition",
  BranchBlock = "branchBlock",
  TryCatch = "tryCatch",
  TryBlock = "tryBlock",
  CatchBlock = "catchBlock",
  BizApply = "bizApply",
  BizUpload = "bizUpload",
  BizMonitor = "bizMonitor",
  BizDeploy = "bizDeploy",
  BizNotify = "bizNotify",
}

/* TYPE GUARD, PLEASE DO NOT REMOVE THESE! */
console.assert(NodeType.Start === WORKFLOW_NODE_TYPES.START);
console.assert(NodeType.End === WORKFLOW_NODE_TYPES.END);
console.assert(NodeType.Delay === WORKFLOW_NODE_TYPES.DELAY);
console.assert(NodeType.Condition === WORKFLOW_NODE_TYPES.CONDITION);
console.assert(NodeType.BranchBlock === WORKFLOW_NODE_TYPES.BRANCHBLOCK);
console.assert(NodeType.TryCatch === WORKFLOW_NODE_TYPES.TRYCATCH);
console.assert(NodeType.TryBlock === WORKFLOW_NODE_TYPES.TRYBLOCK);
console.assert(NodeType.CatchBlock === WORKFLOW_NODE_TYPES.CATCHBLOCK);
console.assert(NodeType.BizApply === WORKFLOW_NODE_TYPES.BIZ_APPLY);
console.assert(NodeType.BizUpload === WORKFLOW_NODE_TYPES.BIZ_UPLOAD);
console.assert(NodeType.BizMonitor === WORKFLOW_NODE_TYPES.BIZ_MONITOR);
console.assert(NodeType.BizDeploy === WORKFLOW_NODE_TYPES.BIZ_DEPLOY);
console.assert(NodeType.BizNotify === WORKFLOW_NODE_TYPES.BIZ_NOTIFY);

export enum NodeKindType {
  Basis = "basis",
  Business = "business",
  Logic = "logic",
}

export interface NodeJSON extends FlowNodeJSON {
  data: WorkflowNode["data"] & {
    [key: string]: any;
  };
}

export interface DocumentJSON {
  nodes: NodeJSON[];
}

export interface NodeMeta extends FlowNodeMeta {
  /**
   * 自定义样式。
   */
  style?: React.CSSProperties;
  /**
   * 标题文本。
   */
  labelText?: React.ReactNode;
  /**
   * 图标组件。
   */
  icon?: React.ExoticComponent<any> | React.ComponentType<any>;
  /**
   * 图标前景色。
   */
  iconColor?: string;
  /**
   * 图标背景色。
   */
  iconBgColor?: string;
  /**
   * 是否可点击。通常配合抽屉表单使用。
   */
  clickable?: boolean;
}

export interface NodeRegistry<V extends NodeJSON["data"] = NodeJSON["data"]> extends FlowNodeRegistry<NodeMeta> {
  /**
   * 节点类型分类。
   */
  kind?: NodeKindType;

  formMeta?: Omit<FormMeta<V>, "render"> & {
    render: (props: FormRenderProps<V>) => React.ReactElement;
  };

  /**
   * 判断是否可以添加一个该类型的节点。
   * 如果不存在该方法，默认等同于返回值为 true。
   * @param {FixedLayoutPluginContext} ctx
   * @param {FlowNodeEntity} from
   * @returns {Boolean}
   */
  canAdd?: (ctx: FixedLayoutPluginContext, from: FlowNodeEntity) => boolean;
  /**
   * 判断是否可以删除一个该类型的节点。
   * 如果不存在该方法，默认等同于返回值为 true。
   * @param {FixedLayoutPluginContext} ctx
   * @param {FlowNodeEntity} from
   * @returns {Boolean}
   */
  canDelete?: (ctx: FixedLayoutPluginContext, from: FlowNodeEntity) => boolean;
  /**
   * 返回一个新的表示该类型的节点结构。
   * @param {FixedLayoutPluginContext} ctx
   * @param {FlowNodeEntity} from
   * @returns {FlowNodeJSON}
   */
  onAdd?: (ctx: FixedLayoutPluginContext, from: FlowNodeEntity) => FlowNodeJSON;
}
