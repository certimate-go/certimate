import { memo, useMemo, useRef } from "react";
import { useTranslation } from "react-i18next";
import {
  CloseCircleOutlined as CloseCircleOutlinedIcon,
  EllipsisOutlined as EllipsisOutlinedIcon,
  FormOutlined as FormOutlinedIcon,
  MoreOutlined as MoreOutlinedIcon,
  SnippetsOutlined as SnippetsOutlinedIcon,
} from "@ant-design/icons";
import { useControllableValue } from "ahooks";
import { Button, Card, Drawer, Dropdown, Input, type InputRef, type MenuProps, Modal, Popover, Space } from "antd";
import { produce } from "immer";
import { isEqual } from "radash";

import { type WorkflowNode, WorkflowNodeType } from "@/domain/workflow";
import { useZustandShallowSelector } from "@/hooks";
import { useWorkflowStore } from "@/stores/workflow";

import AddNode from "./AddNode";

export type SharedNodeProps = {
  node: WorkflowNode;
  disabled?: boolean;
};

// #region Title
type SharedNodeTitleProps = SharedNodeProps & {
  className?: string;
  style?: React.CSSProperties;
};

const SharedNodeTitle = ({ className, style, node, disabled }: SharedNodeTitleProps) => {
  const { updateNode } = useWorkflowStore(useZustandShallowSelector(["updateNode"]));

  const handleBlur = (e: React.FocusEvent<HTMLDivElement>) => {
    const oldName = node.name;
    const newName = e.target.innerText.replaceAll("\r", "").replaceAll("\n", "").trim().substring(0, 64) || oldName;
    if (oldName === newName) {
      return;
    }

    updateNode(
      produce(node, (draft) => {
        draft.name = newName;
      })
    );
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    if (e.code === "Enter") {
      e.preventDefault();
      e.currentTarget.blur();
    }
  };

  return (
    <div className="w-full cursor-text overflow-hidden text-center">
      <div className={className} style={style} contentEditable={!disabled} suppressContentEditableWarning onBlur={handleBlur} onKeyDown={handleKeyDown}>
        {node.name}
      </div>
    </div>
  );
};
// #endregion

// #region Menu
type SharedNodeMenuProps = SharedNodeProps & {
  branchId?: string;
  branchIndex?: number;
  menus?: Array<"rename" | "duplicate" | "remove">;
  trigger: React.ReactNode;
  afterUpdate?: () => void;
  afterDelete?: () => void;
};

const isNodeBranchLike = (node: WorkflowNode) => {
  return (
    node.type === WorkflowNodeType.Branch ||
    node.type === WorkflowNodeType.Condition ||
    node.type === WorkflowNodeType.ExecuteResultBranch ||
    node.type === WorkflowNodeType.ExecuteSuccess ||
    node.type === WorkflowNodeType.ExecuteFailure
  );
};

const isNodeUnduplicatable = (node: WorkflowNode) => {
  return (
    node.type === WorkflowNodeType.Start ||
    node.type === WorkflowNodeType.End ||
    node.type === WorkflowNodeType.Branch ||
    node.type === WorkflowNodeType.ExecuteResultBranch ||
    node.type === WorkflowNodeType.ExecuteSuccess ||
    node.type === WorkflowNodeType.ExecuteFailure
  );
};

const isNodeUnremovable = (node: WorkflowNode) => {
  return node.type === WorkflowNodeType.Start || node.type === WorkflowNodeType.End;
};

const SharedNodeMenu = ({ menus, trigger, node, disabled, branchId, branchIndex, afterUpdate, afterDelete }: SharedNodeMenuProps) => {
  const { t } = useTranslation();

  const { duplicateNode, updateNode, removeNode, duplicateBranch, removeBranch } = useWorkflowStore(
    useZustandShallowSelector(["duplicateNode", "updateNode", "removeNode", "duplicateBranch", "removeBranch"])
  );

  const [modalApi, ModelContextHolder] = Modal.useModal();

  const nameInputRef = useRef<InputRef>(null);
  const nameRef = useRef<string>();

  const handleRenameConfirm = async () => {
    const oldName = node.name;
    const newName = nameRef.current?.replaceAll("\r", "")?.replaceAll("\n", "").trim()?.substring(0, 64) || oldName;
    if (oldName === newName) {
      return;
    }

    await updateNode(
      produce(node, (draft) => {
        draft.name = newName;
      })
    );

    afterUpdate?.();
  };

  const handleDuplicateClick = async () => {
    if (isNodeBranchLike(node)) {
      await duplicateBranch(branchId!, branchIndex!);
    } else {
      await duplicateNode(node);
    }
  };

  const handleRemoveClick = async () => {
    if (isNodeBranchLike(node)) {
      await removeBranch(branchId!, branchIndex!);
    } else {
      await removeNode(node);
    }

    afterDelete?.();
  };

  const menuItems = useMemo(() => {
    let temp = [
      {
        key: "rename",
        disabled: disabled,
        label: isNodeBranchLike(node) ? t("workflow_node.action.rename_branch") : t("workflow_node.action.rename_node"),
        icon: <FormOutlinedIcon />,
        onClick: () => {
          nameRef.current = node.name;

          const dialog = modalApi.confirm({
            title: isNodeBranchLike(node) ? t("workflow_node.action.rename_branch") : t("workflow_node.action.rename_node"),
            content: (
              <div className="pb-2 pt-4">
                <Input
                  ref={nameInputRef}
                  autoFocus
                  defaultValue={node.name}
                  onChange={(e) => (nameRef.current = e.target.value)}
                  onPressEnter={async () => {
                    await handleRenameConfirm();
                    dialog.destroy();
                  }}
                />
              </div>
            ),
            icon: null,
            okText: t("common.button.save"),
            onOk: handleRenameConfirm,
          });
          setTimeout(() => nameInputRef.current?.focus(), 1);
        },
      },
      {
        key: "duplicate",
        disabled: disabled || isNodeUnduplicatable(node),
        label: isNodeBranchLike(node) ? t("workflow_node.action.duplicate_branch") : t("workflow_node.action.duplicate_node"),
        icon: <SnippetsOutlinedIcon />,
        onClick: handleDuplicateClick,
      },
      {
        type: "divider",
      },
      {
        key: "remove",
        disabled: disabled || isNodeUnremovable(node),
        label: isNodeBranchLike(node) ? t("workflow_node.action.remove_branch") : t("workflow_node.action.remove_node"),
        icon: <CloseCircleOutlinedIcon />,
        danger: true,
        onClick: handleRemoveClick,
      },
    ] satisfies MenuProps["items"];

    if (menus) {
      temp = temp.filter((item) => item.type === "divider" || menus.includes(item.key as "rename" | "remove"));
      temp = temp.filter((item, index, array) => {
        if (item.type !== "divider") return true;
        return index === 0 || array[index - 1].type !== "divider";
      });
      if (temp[0]?.type === "divider") {
        temp.shift();
      }
      if (temp[temp.length - 1]?.type === "divider") {
        temp.pop();
      }
    }

    return temp;
  }, [disabled, node]);

  return (
    <>
      {ModelContextHolder}

      <Dropdown
        menu={{
          items: menuItems,
        }}
        trigger={["click"]}
      >
        {trigger}
      </Dropdown>
    </>
  );
};
// #endregion

// #region Block
type SharedNodeBlockProps = SharedNodeProps & {
  children: React.ReactNode;
  onClick?: (e: React.MouseEvent) => void;
};

const SharedNodeBlock = ({ children, node, disabled, onClick }: SharedNodeBlockProps) => {
  const handleNodeClick = (e: React.MouseEvent) => {
    onClick?.(e);
  };

  return (
    <>
      <Popover
        classNames={{ root: "shadow-md" }}
        styles={{ body: { padding: 0 } }}
        arrow={false}
        content={<SharedNodeMenu node={node} disabled={disabled} trigger={<Button color="primary" icon={<MoreOutlinedIcon />} variant="text" />} />}
        placement="rightTop"
      >
        <div className="relative w-[256px] overflow-hidden">
          <Card className="shadow-md" styles={{ body: { padding: 0 } }} hoverable>
            <div className="bg-primary flex h-[48px] flex-col items-center justify-center truncate px-4 py-2 text-white">
              <SharedNodeTitle
                className="focus:bg-background focus:text-foreground overflow-hidden outline-hidden focus:rounded-xs"
                node={node}
                disabled={disabled}
              />
            </div>

            <div className="flex cursor-pointer flex-col justify-center px-4 py-2" onClick={handleNodeClick}>
              <div className="overflow-hidden text-sm">{children}</div>
            </div>
          </Card>
        </div>
      </Popover>

      <AddNode node={node} disabled={disabled} />
    </>
  );
};
// #endregion

// #region EditDrawer
type SharedNodeEditDrawerProps = SharedNodeProps & {
  children: React.ReactNode;
  footer?: boolean;
  loading?: boolean;
  open?: boolean;
  pending?: boolean;
  onOpenChange?: (open: boolean) => void;
  onConfirm: () => void | Promise<unknown>;
  getConfigNewValues: () => NonNullable<unknown>; // 用于获取节点配置的新值，以便在抽屉关闭前进行对比，决定是否提示保存
};

const SharedNodeConfigDrawer = ({
  children,
  node,
  disabled,
  footer = true,
  loading,
  pending,
  onConfirm,
  getConfigNewValues,
  ...props
}: SharedNodeEditDrawerProps) => {
  const { t } = useTranslation();

  const [modalApi, ModelContextHolder] = Modal.useModal();

  const [open, setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });

  const handleConfirmClick = async () => {
    await onConfirm();
    setOpen(false);
  };

  const handleCancelClick = () => {
    if (pending) return;

    setOpen(false);
  };

  const handleClose = () => {
    if (pending) return;

    const oldValues = JSON.parse(JSON.stringify(node.config ?? {}));
    const newValues = JSON.parse(JSON.stringify(getConfigNewValues()));
    const changed = !isEqual(oldValues, {}) && !isEqual(oldValues, newValues);

    const { promise, resolve, reject } = Promise.withResolvers();
    if (changed) {
      modalApi.confirm({
        title: t("common.text.operation_confirm"),
        content: t("workflow_node.unsaved_changes.confirm"),
        onOk: () => resolve(void 0),
        onCancel: () => reject(),
      });
    } else {
      resolve(void 0);
    }

    promise.then(() => setOpen(false));
  };

  return (
    <>
      {ModelContextHolder}

      <Drawer
        afterOpenChange={setOpen}
        closable={!pending}
        destroyOnHidden
        extra={
          <SharedNodeMenu
            menus={["rename", "remove"]}
            node={node}
            disabled={disabled}
            trigger={<Button icon={<EllipsisOutlinedIcon />} type="text" />}
            afterDelete={() => {
              setOpen(false);
            }}
          />
        }
        footer={
          !!footer && (
            <Space className="w-full justify-end">
              <Button onClick={handleCancelClick}>{t("common.button.cancel")}</Button>
              <Button disabled={disabled} loading={pending} type="primary" onClick={handleConfirmClick}>
                {t("common.button.save")}
              </Button>
            </Space>
          )
        }
        loading={loading}
        maskClosable={!pending}
        open={open}
        title={<div className="max-w-[480px] truncate">{node.name}</div>}
        width={720}
        onClose={handleClose}
      >
        {children}
      </Drawer>
    </>
  );
};
// #endregion

export default {
  Title: memo(SharedNodeTitle),
  Menu: memo(SharedNodeMenu),
  Block: memo(SharedNodeBlock),
  ConfigDrawer: memo(SharedNodeConfigDrawer),
};
