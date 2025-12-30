import { startTransition, useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { IconChevronDown, IconX } from "@tabler/icons-react";
import { useControllableValue, useGetState } from "ahooks";
import { App, Button, Drawer, Dropdown, Flex, Form, Space } from "antd";

import { testPushNotification } from "@/api/notifications";
import AccessProviderPicker from "@/components/provider/AccessProviderPicker";
import Show from "@/components/Show";
import { type AccessModel } from "@/domain/access";
import { ACCESS_USAGES } from "@/domain/provider";
import { useTriggerElement, useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";
import { unwrapErrMsg } from "@/utils/error";

import AccessForm, { type AccessFormModes, type AccessFormProps, type AccessFormUsages } from "./AccessForm";

export interface AccessEditDrawerProps {
  afterClose?: () => void;
  afterSubmit?: (record: AccessModel) => void;
  data?: AccessFormProps["initialValues"];
  loading?: boolean;
  mode: AccessFormModes;
  open?: boolean;
  trigger?: React.ReactNode;
  usage?: AccessFormUsages;
  onOpenChange?: (open: boolean) => void;
}

const AccessEditDrawer = ({ afterSubmit, mode, data, loading, trigger, usage, ...props }: AccessEditDrawerProps) => {
  const { t } = useTranslation();

  const { message, notification } = App.useApp();

  const { createAccess, updateAccess } = useAccessesStore(useZustandShallowSelector(["createAccess", "updateAccess"]));

  const [open, setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });

  const afterClose = () => {
    setFormPending(false);
    setFormChanged(false);
    setIsTesting(false);
    props.afterClose?.();
  };

  const triggerEl = useTriggerElement(trigger, { onClick: () => setOpen(true) });

  const providerFilter = AccessForm.useProviderFilterByUsage(usage);

  const [formInst] = Form.useForm();
  const [formPending, setFormPending] = useState(false);
  const [formChanged, setFormChanged] = useState(false);

  const submitForm = async () => {
    let formValues: AccessModel;

    setFormPending(true);
    try {
      formValues = await formInst.validateFields();
      formValues.reserve = usage === "ca" ? "ca" : usage === "notification" ? "notif" : void 0;
    } catch (err) {
      message.warning(t("common.errmsg.form_invalid"));

      setFormPending(false);
      throw err;
    }

    try {
      switch (mode) {
        case "create":
          {
            if (data?.id) {
              throw "Invalid props: `data`";
            }

            formValues = await createAccess(formValues);
          }
          break;

        case "modify":
          {
            if (!data?.id) {
              throw "Invalid props: `data`";
            }

            formValues = await updateAccess({ ...data, ...formValues });
          }
          break;

        default:
          throw "Invalid props: `mode`";
      }

      afterSubmit?.(formValues);
    } catch (err) {
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

      throw err;
    } finally {
      setFormPending(false);
    }
  };

  const fieldProvider = Form.useWatch<string>("provider", { form: formInst, preserve: true });

  const [isTesting, setIsTesting] = useState(false);

  const handleProviderPick = (value: string) => {
    formInst.setFieldValue("provider", value);
  };

  const handleFormChange = () => {
    setFormChanged(true);
  };

  const handleOkClick = async () => {
    await submitForm();
    setOpen(false);
  };

  const handleOkAndContinueClick = async () => {
    await submitForm();
    message.success(t("common.text.saved"));
  };

  const handleCancelClick = () => {
    if (formPending) return;

    setOpen(false);
  };

  const handleTestPushClick = async () => {
    setIsTesting(true);

    try {
      await formInst.validateFields();
    } catch {
      setIsTesting(false);
      return;
    }

    try {
      await testPushNotification({ provider: fieldProvider, accessId: data!.id });
      message.success(t("common.text.operation_succeeded"));
    } catch (err) {
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
    } finally {
      setIsTesting(false);
    }
  };

  return (
    <>
      {triggerEl}

      <Drawer
        afterOpenChange={(open) => !open && afterClose?.()}
        autoFocus
        closeIcon={false}
        destroyOnHidden
        footer={
          fieldProvider ? (
            <Flex className="px-2" justify="space-between">
              {usage === "notification" ? (
                <Button className="max-sm:invisible" disabled={mode !== "modify" || formChanged} loading={isTesting} onClick={handleTestPushClick}>
                  {t("access.action.test_notify.button")}
                </Button>
              ) : (
                <span>{/* TODO: 测试连接 */}</span>
              )}
              <Flex justify="end" gap="small">
                <Button disabled={isTesting} onClick={handleCancelClick}>
                  {t("common.button.cancel")}
                </Button>
                <Space.Compact>
                  <Button disabled={isTesting} loading={formPending} type="primary" onClick={handleOkClick}>
                    {mode === "modify" ? t("common.button.save") : t("common.button.submit")}
                  </Button>
                  <Show when={mode === "modify"}>
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
                      <Button disabled={formPending || isTesting} icon={<IconChevronDown size="1.25em" />} type="primary" />
                    </Dropdown>
                  </Show>
                </Space.Compact>
              </Flex>
            </Flex>
          ) : (
            false
          )
        }
        loading={loading}
        maskClosable={!formPending}
        open={open}
        size="large"
        title={
          <Flex align="center" justify="space-between" gap="small">
            <div className="flex-1 truncate">
              {mode === "modify" && !!data?.id ? t("access.action.modify.modal.title") + ` #${data.id}` : t(`access.action.${mode}.modal.title`)}
            </div>
            <Button
              className="ant-drawer-close"
              style={{ marginInline: 0 }}
              icon={<IconX size="1.25em" />}
              size="small"
              type="text"
              onClick={handleCancelClick}
            />
          </Flex>
        }
        onClose={handleCancelClick}
      >
        <Show when={!fieldProvider && !data?.provider}>
          <AccessProviderPicker
            autoFocus
            gap="large"
            placeholder={t("access.form.provider.search.placeholder")}
            showOptionTags={
              usage == null ||
              (usage === "dns-hosting" ? { ["builtin"]: true, [ACCESS_USAGES.DNS]: true, [ACCESS_USAGES.HOSTING]: true } : { ["builtin"]: true })
            }
            showSearch
            onFilter={providerFilter}
            onSelect={handleProviderPick}
          />
        </Show>

        <div style={{ display: fieldProvider || data?.provider ? "block" : "none" }}>
          <AccessForm form={formInst} disabled={formPending} initialValues={data} mode={mode} usage={usage} onFormValuesChange={handleFormChange} />
        </div>
      </Drawer>
    </>
  );
};

const useDrawer = () => {
  type DataType = AccessEditDrawerProps["data"];
  const [data, setData, getData] = useGetState<DataType>();
  const [loading, setLoading] = useState<boolean>();
  const [open, setOpen] = useState(false);

  const onOpenChange = useCallback((open: boolean) => {
    setOpen(open);
  }, []);

  return {
    drawerProps: {
      afterClose: () => {
        startTransition(() => {
          if (!open) {
            setData(void 0);
            setLoading(void 0);
          }
        });
      },
      data,
      loading,
      open,
      onOpenChange,
    },

    open: ({ data, loading }: { data: NonNullable<DataType>; loading?: boolean }) => {
      setData(data);
      setLoading(loading);
      setOpen(true);

      return {
        safeUpdate: ({ data, loading }: { data?: NonNullable<DataType>; loading?: boolean }) => {
          if (data != null) {
            if (data.id !== getData()?.id) return; // 确保数据不脏读

            setData(data);
          }

          if (loading != null) {
            setLoading(loading);
          }
        },
      };
    },
    close: () => {
      setOpen(false);
    },
  };
};

const _default = Object.assign(AccessEditDrawer, {
  useDrawer,
});

export default _default;
