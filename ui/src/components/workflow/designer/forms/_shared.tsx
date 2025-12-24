import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { type FlowNodeEntity, useClientContext, useRefresh } from "@flowgram.ai/fixed-layout-editor";
import { IconChevronDown, IconEye, IconEyeOff, IconX } from "@tabler/icons-react";
import { useControllableValue } from "ahooks";
import { Anchor, type AnchorProps, App, Button, Drawer, Dropdown, Flex, type FormInstance, Space, Tooltip, Typography } from "antd";
import { isEqual } from "radash";

import Show from "@/components/Show";
import { unwrapErrMsg } from "@/utils/error";

import { type NodeRegistry } from "../nodes/typings";

export interface NodeConfigDrawerProps {
  children: React.ReactNode;
  afterClose?: () => void;
  anchor?: Pick<Required<AnchorProps>, "items"> | false;
  footer?: boolean;
  form: FormInstance;
  loading?: boolean;
  node: FlowNodeEntity;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export const NodeConfigDrawer = ({ children, afterClose, anchor, footer = true, form: formInst, loading, node, ...props }: NodeConfigDrawerProps) => {
  const { t } = useTranslation();

  const ctx = useClientContext();
  const { playground } = ctx;

  const refresh = useRefresh();

  const { message, modal, notification } = App.useApp();

  const [open, setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });

  const containerRef = useRef<HTMLDivElement>(null);

  const [formPending, setFormPending] = useState(false);

  const submitForm = async () => {
    let formValues: Record<string, unknown>;

    setFormPending(true);
    try {
      formValues = await formInst.validateFields();
    } catch (err) {
      message.warning(t("common.errmsg.form_invalid"));

      setFormPending(false);
      throw err;
    }

    try {
      node.form!.setValueIn("config", formValues);
      node.form!.validate();
    } catch (err) {
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

      throw err;
    } finally {
      setFormPending(false);
    }
  };

  const nodeRegistry = node?.getNodeRegistry<NodeRegistry>();
  const NodeIcon = nodeRegistry?.meta?.icon;
  const renderNodeIcon = () =>
    NodeIcon == null ? null : (
      <div
        className="flex size-6 items-center justify-center rounded-lg bg-white text-primary shadow-md dark:bg-stone-200"
        style={{
          color: nodeRegistry?.meta?.iconColor,
          backgroundColor: nodeRegistry?.meta?.iconBgColor,
        }}
      >
        <NodeIcon size="1em" color={nodeRegistry?.meta?.iconColor} stroke="1.25" />
      </div>
    );

  const [isNodeDisabled, setIsNodeDisabled] = useState(() => {
    if (node) {
      return node.form?.getValueIn<boolean>("disabled");
    }
    return false;
  });
  useEffect(() => {
    const d1 = playground.config.onDataChange(() => refresh());
    const d2 = node?.onDataChange(() => setIsNodeDisabled(node.form?.getValueIn<boolean>("disabled")));

    return () => {
      d1.dispose();
      d2.dispose();
    };
  });

  const handleOkClick = async () => {
    if (node == null) {
      setOpen(false);
      return;
    }

    await submitForm();
    setOpen(false);
  };

  const handleOkAndContinueClick = async () => {
    if (node == null) {
      setOpen(false);
      return;
    }

    await submitForm();
    message.success(t("common.text.saved"));
  };

  const handleCancelClick = () => {
    if (formPending) return;

    setOpen(false);
  };

  const handleClose = () => {
    if (formPending) return;

    const picker = (obj: Record<string, unknown>) => {
      return Object.entries(obj).reduce(
        (acc, [key, value]) => {
          const isEmpty =
            value == null ||
            (typeof value === "string" && value === "") ||
            (Array.isArray(value) && value.length === 0) ||
            (typeof value === "object" && !Array.isArray(value) && Object.keys(value).length === 0);

          if (!isEmpty) {
            acc[key] = value;
          }

          return acc;
        },
        {} as Record<string, unknown>
      );
    };
    const oldValues = picker(node?.toJSON()?.data?.config ?? {});
    const newValues = picker(formInst.getFieldsValue(true));
    const changed = !isEqual(oldValues, {}) && !isEqual(oldValues, newValues);

    const { promise, resolve } = Promise.withResolvers();
    if (changed) {
      modal.confirm({
        title: t("common.text.operation_confirm"),
        content: t("workflow.detail.design.unsaved_changes.confirm"),
        onOk: () => resolve(void 0),
      });
    } else {
      resolve(void 0);
    }

    promise.then(() => setOpen(false));
  };

  const handleDisableNodeClick = () => {
    node.form!.setValueIn("disabled", !isNodeDisabled);
  };

  return (
    <Drawer
      styles={{
        header: {
          paddingBottom: anchor ? 0 : void 0,
        },
      }}
      afterOpenChange={(open) => !open && afterClose?.()}
      autoFocus
      closeIcon={false}
      destroyOnHidden
      footer={
        footer ? (
          <Flex className="px-2" justify="end" gap="small">
            <Button onClick={handleCancelClick}>{t("common.button.cancel")}</Button>
            <Space.Compact>
              <Button loading={formPending} type="primary" onClick={handleOkClick}>
                {t("common.button.save")}
              </Button>
              <Dropdown
                menu={{
                  items: [
                    {
                      key: "save_and_continue",
                      label: t("common.button.save_and_continue"),
                      onClick: handleOkAndContinueClick,
                    },
                  ],
                }}
                placement="bottomRight"
                trigger={["click"]}
              >
                <Button disabled={formPending} icon={<IconChevronDown size="1.25em" />} type="primary" />
              </Dropdown>
            </Space.Compact>
          </Flex>
        ) : (
          <></>
        )
      }
      forceRender={false}
      loading={loading}
      maskClosable={!formPending}
      open={open}
      size="large"
      title={
        <>
          <Flex align="center" justify="space-between" gap="small">
            <div>{renderNodeIcon()}</div>
            <div className="flex-1 truncate">{node?.toJSON()?.data?.name}</div>
            <Show when={!!node && !nodeRegistry?.meta?.isStart && !nodeRegistry?.meta?.isNodeEnd}>
              <Tooltip
                title={isNodeDisabled ? t("workflow.detail.design.drawer.disabled.on.tooltip") : t("workflow.detail.design.drawer.disabled.off.tooltip")}
              >
                <Button
                  className="ant-drawer-close"
                  style={{ marginInline: 0 }}
                  disabled={playground.config.readonlyOrDisabled}
                  icon={isNodeDisabled ? <IconEyeOff size="1.25em" /> : <IconEye size="1.25em" />}
                  size="small"
                  type="text"
                  onClick={handleDisableNodeClick}
                />
              </Tooltip>
            </Show>
            <Button className="ant-drawer-close" style={{ marginInline: 0 }} icon={<IconX size="1.25em" />} size="small" type="text" onClick={handleClose} />
          </Flex>

          <div className="mt-3 truncate text-sm font-normal">
            <Typography.Text className="text-xs" type="secondary">
              <span>{t("workflow.detail.design.drawer.node_id.label")}</span>
              <span>{node?.id}</span>
            </Typography.Text>
          </div>

          <Show when={!!anchor}>
            <div className="-mx-0.5 mt-3 text-sm font-normal">
              <Anchor
                affix={false}
                getContainer={() => containerRef.current!}
                direction="horizontal"
                items={(anchor as AnchorProps).items}
                onClick={(e, link) => {
                  // https://github.com/ant-design/ant-design/issues/10577
                  // https://github.com/ant-design/ant-design/issues/15326
                  e.preventDefault();

                  // 锚点元素需同时包含 `id` 和 `data-anchor` 两个属性
                  const el = document.querySelector(`[data-anchor="${link.href.replace(/^#/g, "")}"]`);
                  el?.scrollIntoView({ block: "start", behavior: "smooth" });
                }}
              />
            </div>
          </Show>
        </>
      }
      onClose={handleClose}
    >
      <div ref={containerRef} style={{ height: "100%", overflowX: "hidden", overflowY: "auto" }}>
        {children}
      </div>
    </Drawer>
  );
};
