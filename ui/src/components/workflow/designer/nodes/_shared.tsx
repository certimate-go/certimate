import { startTransition, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Field,
  type FlowNodeEntity,
  FlowNodeRenderData,
  type NodeRenderReturnType,
  useClientContext,
  useWatchFormState,
  useWatchFormValueIn,
} from "@flowgram.ai/fixed-layout-editor";
import { IconCopy, IconDotsVertical, IconGripVertical, IconLabel, IconX } from "@tabler/icons-react";
import { Button, type ButtonProps, Card, Dropdown, Input, type InputRef, Popover, theme } from "antd";

import { mergeCls } from "@/utils/css";

import { type NodeJSON, type NodeRegistry } from "./typings";
import { duplicateNodeJSON } from "../_util";
import { useNodeRenderContext } from "../NodeRenderContext";

const useInternalRenamingInput = ({ nodeRender }: { nodeRender: NodeRenderReturnType }) => {
  const inputRef = useRef<InputRef>(null);
  const [inputVisible, setInputVisible] = useState(false);
  const [inputValue, setInputValue] = useState("");

  const showInput = () => {
    setInputVisible(true);
    setInputValue(nodeRender.data?.name);
    setTimeout(() => {
      inputRef.current?.focus({ cursor: "end" });
    }, 0);
  };

  const hideInput = () => {
    setInputVisible(false);
    setInputValue(nodeRender.data?.name);
  };

  const handleInputBlur = async (e: React.FocusEvent<HTMLInputElement>) => {
    const value = e.target.value.trim();
    if (!value || value === (nodeRender.data?.name || "")) {
      setInputVisible(false);
      return;
    }

    setInputVisible(false);

    nodeRender.updateData({ ...nodeRender.data, name: value });
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value);
  };

  const handleInputMouseDown = (e: React.MouseEvent<HTMLInputElement>) => {
    e.stopPropagation();
  };

  const handleInputMouseUp = (e: React.MouseEvent<HTMLInputElement>) => {
    e.stopPropagation();
  };

  const handleInputPressEnter = (e: React.KeyboardEvent<HTMLInputElement>) => {
    e.currentTarget.blur();
  };

  return {
    inputRef: inputRef,
    inputProps: {
      value: inputValue,
      onBlur: handleInputBlur,
      onChange: handleInputChange,
      onPressEnter: handleInputPressEnter,
      onMouseDown: handleInputMouseDown,
      onMouseUp: handleInputMouseUp,
    },
    visible: inputVisible,
    value: inputValue,

    show: showInput,
    hide: hideInput,
  };
};

const InternalNodeCard = ({
  className,
  style,
  children,
  nodeRender,
}: {
  className?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
  nodeRender: NodeRenderReturnType;
}) => {
  const nodeRenderData = nodeRender.node.getData(FlowNodeRenderData)!;
  const nodeRegistry = nodeRender.node.getNodeRegistry<NodeRegistry>();

  const isActivated = useMemo(() => nodeRenderData.activated || nodeRenderData.lineActivated, [nodeRenderData.activated, nodeRenderData.lineActivated]);
  const [isHovering, setIsHovering] = useState(false);
  const [isNodeInvalid, setIsNodeInvalid] = useState(false);
  const isNodeDisabled = useWatchFormValueIn(nodeRender.node, "disabled");

  const formState = useWatchFormState(nodeRender.node);
  useEffect(() => setIsNodeInvalid(!!formState?.invalid), [formState?.invalid]);

  return (
    <Card
      className={mergeCls(
        "relative rounded-xl shadow-sm",
        { "border-primary": isActivated },
        { "border-dashed": isNodeDisabled },
        nodeRegistry.meta?.clickable ? "cursor-pointer" : "cursor-default",
        className
      )}
      style={style}
      styles={{ body: { padding: 0 } }}
      hoverable
      onMouseEnter={() => startTransition(() => setIsHovering(true))}
      onMouseLeave={() => startTransition(() => setIsHovering(false))}
    >
      <div
        className="relative z-1 transition-opacity"
        style={{
          opacity: isHovering ? 1 : isNodeDisabled ? 0.3 : void 0,
        }}
      >
        {children}
      </div>
      <div
        className="absolute z-0 rounded-xl border-solid border-transparent transition-all duration-500"
        style={{
          top: "-1px",
          left: "-1px",
          right: "-1px",
          bottom: "-1px",
          borderWidth: "2px",
          borderColor: isHovering ? "var(--color-primary)" : isNodeInvalid ? "var(--color-error)" : void 0,
          borderStyle: isNodeDisabled ? "dashed" : "solid",
        }}
      />
    </Card>
  );
};

const InternalNodeMenuButton = ({
  className,
  style,
  onMenuSelect,
  ...props
}: ButtonProps & {
  onMenuSelect?: (key: "rename" | "duplicate" | "remove") => void;
}) => {
  const { t } = useTranslation();

  const ctx = useClientContext();
  const { operation, playground } = ctx;

  const { node, ...nodeRender } = useNodeRenderContext();
  const nodeRegistry = node.getNodeRegistry<NodeRegistry>();

  const getLatestDuplicateDisabledState = () => {
    if (nodeRegistry.meta?.copyDisable != null && nodeRegistry.meta.copyDisable) {
      return true;
    }
    return false;
  };
  const getLatestRemoveDisabledState = () => {
    if (nodeRegistry.meta?.deleteDisable != null && nodeRegistry.meta.deleteDisable) {
      return true;
    }
    if (nodeRegistry.canDelete != null) {
      return !nodeRegistry.canDelete(ctx, node);
    }
    return false;
  };
  const [menuDuplicateDisabled, setMenuDuplicateDisabled] = useState(() => getLatestDuplicateDisabledState());
  const [menuRemoveDisabled, setMenuRemoveDisabled] = useState(() => getLatestRemoveDisabledState());
  useEffect(() => {
    // 这里不能使用 useMemo() 来决定 menuRemoveDisabled，因为依赖项没有发生改变（对象引用始终是同一个）
    // 因此需要使用事件钩子来监听，并更新 menuRemoveDisabled 的状态
    const callback = () => {
      setMenuDuplicateDisabled(getLatestDuplicateDisabledState());
      setMenuRemoveDisabled(getLatestRemoveDisabledState());
    };
    const d1 = node.onEntityChange(callback);
    const d2 = node.parent?.onEntityChange?.(callback);

    return () => {
      d1?.dispose();
      d2?.dispose();
    };
  }, []);

  const handleClickRename = () => {
    onMenuSelect?.("rename");
  };

  const handleClickDuplicate = () => {
    if (menuDuplicateDisabled) {
      return;
    }

    const parent = node.originParent ?? node.parent;
    if (parent != null) {
      const nodeJSON = duplicateNodeJSON(node.toJSON() as NodeJSON);

      let block: FlowNodeEntity;
      if (nodeRender.isBlockOrderIcon) {
        block = operation.addBlock(parent, nodeJSON);
      } else {
        block = operation.addFromNode(node, nodeJSON);
      }

      setTimeout(() => {
        playground.scrollToView({
          bounds: block.bounds,
          scrollToCenter: true,
        });
      }, 1);

      onMenuSelect?.("duplicate");
    }
  };

  const handleClickRemove = () => {
    if (menuRemoveDisabled) {
      return;
    }

    nodeRender.deleteNode();

    onMenuSelect?.("remove");
  };

  return playground.config.readonlyOrDisabled ? null : (
    <Dropdown
      styles={{
        root: {
          zIndex: 10 /* 确保要比 Minimap 组件层级要高，防止被遮挡而点击不到 */,
        },
      }}
      arrow={false}
      destroyOnHidden
      menu={{
        items: [
          {
            key: "rename",
            label: nodeRender.isBlockOrderIcon ? t("workflow.detail.design.editor.rename_branch") : t("workflow.detail.design.editor.rename_node"),
            icon: <IconLabel size="1em" />,
            onClick: handleClickRename,
          },
          {
            key: "duplicate",
            label: nodeRender.isBlockOrderIcon ? t("workflow.detail.design.editor.duplicate_branch") : t("workflow.detail.design.editor.duplicate_node"),
            icon: <IconCopy size="1em" />,
            disabled: menuDuplicateDisabled,
            onClick: handleClickDuplicate,
          },
          {
            type: "divider",
          },
          {
            key: "remove",
            label: nodeRender.isBlockOrderIcon ? t("workflow.detail.design.editor.remove_branch") : t("workflow.detail.design.editor.remove_node"),
            icon: <IconX size="1em" />,
            danger: true,
            disabled: menuRemoveDisabled,
            onClick: handleClickRemove,
          },
        ],
        onClick: (e) => {
          e.domEvent.stopPropagation();
        },
      }}
      trigger={["click"]}
    >
      <Button
        className={className}
        style={style}
        icon={<IconDotsVertical color="grey" size="1.25em" />}
        type="text"
        {...props}
        onClick={(e) => {
          e.stopPropagation();
          props.onClick?.(e);
        }}
      />
    </Dropdown>
  );
};

export interface BaseNodeProps {
  className?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
  description?: React.ReactNode;
}

export const BaseNode = ({ className, style, children, description }: BaseNodeProps) => {
  const { token: themeToken } = theme.useToken();

  const ctx = useClientContext();
  const { playground } = ctx;

  const nodeRender = useNodeRenderContext();
  const nodeRegistry = nodeRender.node.getNodeRegistry<NodeRegistry>();

  const NodeIcon = nodeRegistry.meta?.icon;
  const renderNodeIcon = () => {
    return NodeIcon == null ? null : (
      <div
        className="mr-2 flex size-9 items-center justify-center rounded-lg bg-white text-primary shadow-md dark:bg-stone-200"
        style={{
          color: nodeRegistry.meta?.iconColor,
          backgroundColor: nodeRegistry.meta?.iconBgColor,
        }}
      >
        <NodeIcon size="1.75em" color={nodeRegistry.meta?.iconColor} stroke="1.25" />
      </div>
    );
  };

  const { inputRef, inputProps, visible: inputVisible, show: showInput } = useInternalRenamingInput({ nodeRender });

  return (
    <Popover
      classNames={{ root: "shadow-md" }}
      styles={{ container: { padding: 0 } }}
      arrow={false}
      content={
        inputVisible ? null : (
          <InternalNodeMenuButton
            variant="text"
            onMenuSelect={(key) => {
              switch (key) {
                case "rename":
                  showInput();
                  break;
              }
            }}
          />
        )
      }
      placement="rightTop"
    >
      <div
        className="group/node relative"
        onClick={(e) => {
          if (inputVisible) {
            e.stopPropagation();
          }
        }}
      >
        <InternalNodeCard className={mergeCls("w-[320px]", className)} style={style} nodeRender={nodeRender}>
          {children != null ? (
            children
          ) : (
            <div className={mergeCls("flex items-center gap-1 overflow-hidden p-3", inputVisible ? "invisible" : "visible")}>
              {renderNodeIcon()}
              <div className="flex-1 overflow-hidden">
                <div className="truncate">
                  <Field<string> name="name">{({ field: { value } }) => <>{value || "\u00A0"}</>}</Field>
                </div>
                {description != null && (
                  <div className="truncate text-xs" style={{ color: themeToken.colorTextTertiary }}>
                    {description}
                  </div>
                )}
              </div>
            </div>
          )}
        </InternalNodeCard>

        {!playground.config.readonlyOrDisabled && nodeRegistry.meta?.draggable === true && (
          <div className="absolute top-1/2 -left-4 z-1 hidden -translate-y-1/2 group-hover/node:block">
            <IconGripVertical size="1em" stroke="1" />
          </div>
        )}

        {!playground.config.readonlyOrDisabled && (
          <div className={mergeCls("absolute top-1/2 left-2 right-2 -translate-y-1/2 z-1", inputVisible ? "block" : "hidden")}>
            <Input ref={inputRef} maxLength={100} variant="underlined" {...inputProps} />
          </div>
        )}
      </div>
    </Popover>
  );
};

export interface BranchNodeProps extends BaseNodeProps {}

export const BranchNode = ({ className, style, children, description }: BranchNodeProps) => {
  const ctx = useClientContext();
  const { playground } = ctx;

  const nodeRender = useNodeRenderContext();
  const nodeRegistry = nodeRender.node.getNodeRegistry<NodeRegistry>();

  const { inputRef, inputProps, visible: inputVisible, show: showInput } = useInternalRenamingInput({ nodeRender });

  return (
    <Popover
      classNames={{ root: "shadow-md" }}
      styles={{ container: { padding: 0 } }}
      arrow={false}
      content={
        inputVisible ? null : (
          <InternalNodeMenuButton
            variant="text"
            onMenuSelect={(key) => {
              switch (key) {
                case "rename":
                  showInput();
                  break;
              }
            }}
          />
        )
      }
      placement="rightTop"
    >
      <div
        className="group/node relative"
        onClick={(e) => {
          if (inputVisible) {
            e.stopPropagation();
          }
        }}
      >
        <InternalNodeCard className={mergeCls("w-[240px]", className)} style={style} nodeRender={nodeRender}>
          {children != null ? (
            children
          ) : (
            <div className={mergeCls("overflow-hidden px-3 py-2", inputVisible ? "invisible" : "visible")}>
              <div className="truncate text-center">
                {description ?? <Field<string> name="name">{({ field: { value } }) => <>{value || "\u00A0"}</>}</Field>}
              </div>
            </div>
          )}
        </InternalNodeCard>

        {!playground.config.readonlyOrDisabled && nodeRegistry.meta?.draggable === true && (
          <div className="absolute top-1/2 -left-4 z-1 hidden -translate-y-1/2 group-hover/node:block">
            <IconGripVertical size="1em" stroke="1" />
          </div>
        )}

        {!playground.config.readonlyOrDisabled && (
          <div className={mergeCls("absolute top-1/2 left-2 right-2 -translate-y-1/2 z-1", inputVisible ? "block" : "hidden")}>
            <Input ref={inputRef} maxLength={100} variant="underlined" {...inputProps} />
          </div>
        )}
      </div>
    </Popover>
  );
};
