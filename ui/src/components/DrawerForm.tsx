import { useTranslation } from "react-i18next";
import { IconX } from "@tabler/icons-react";
import { useControllableValue } from "ahooks";
import { Button, Drawer, type DrawerProps, Flex, Form, type FormProps, type ModalProps } from "antd";

import { useAntdForm, useTriggerElement } from "@/hooks";

export interface DrawerFormProps<T extends NonNullable<unknown> = any> extends Omit<FormProps<T>, "title" | "onFinish"> {
  className?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
  cancelButtonProps?: ModalProps["cancelButtonProps"];
  cancelText?: ModalProps["cancelText"];
  defaultOpen?: boolean;
  drawerProps?: Omit<DrawerProps, "defaultOpen" | "forceRender" | "open" | "title" | "width" | "onOpenChange">;
  okButtonProps?: ModalProps["okButtonProps"];
  okText?: ModalProps["okText"];
  open?: boolean;
  title?: React.ReactNode;
  trigger?: React.ReactNode;
  onFinish?: (values: T) => unknown | Promise<unknown>;
  onOpenChange?: (open: boolean) => void;
}

const DrawerForm = <T extends NonNullable<unknown> = any>({
  className,
  style,
  children,
  cancelText,
  cancelButtonProps,
  form,
  drawerProps,
  okText,
  okButtonProps,
  title,
  trigger,
  onFinish,
  ...props
}: DrawerFormProps<T>) => {
  const { t } = useTranslation();

  const [open, setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });

  const triggerEl = useTriggerElement(trigger, {
    onClick: () => {
      setOpen(true);
    },
  });

  const {
    form: formInst,
    formPending,
    formProps,
    submit: submitForm,
  } = useAntdForm({
    form,
    onSubmit: (values) => {
      return onFinish?.(values);
    },
  });

  const mergedFormProps: FormProps = {
    clearOnDestroy: drawerProps?.destroyOnHidden ? true : void 0,
    ...formProps,
    ...props,
  };

  const mergedDrawerProps: DrawerProps = {
    ...drawerProps,
    closeIcon: false,
    onClose: async (e) => {
      if (formPending) return;

      // 关闭 Drawer 时 Promise.reject 阻止关闭
      await drawerProps?.onClose?.(e);
      setOpen(false);

      if (!mergedFormProps.preserve) {
        formInst.resetFields();
      }
    },
  };

  const handleOkClick = async () => {
    // 提交表单返回 Promise.reject 时不关闭 Drawer
    await submitForm();

    setOpen(false);
  };

  const handleCancelClick = () => {
    if (formPending) return;

    setOpen(false);
  };

  return (
    <>
      {triggerEl}

      <Drawer
        {...mergedDrawerProps}
        footer={
          <Flex className="px-2" justify="end" gap="small">
            <Button {...cancelButtonProps} onClick={handleCancelClick}>
              {cancelText ?? t("common.button.cancel")}
            </Button>
            <Button {...okButtonProps} type="primary" loading={formPending} onClick={handleOkClick}>
              {okText ?? t("common.button.ok")}
            </Button>
          </Flex>
        }
        forceRender
        open={open}
        title={
          <Flex align="center" justify="space-between" gap="small">
            <div className="flex-1 truncate">{title}</div>
            {mergedDrawerProps.closeIcon !== false && (
              <Button
                className="ant-drawer-close"
                style={{ marginInline: 0 }}
                icon={mergedDrawerProps.closeIcon ?? <IconX size="1.25em" />}
                size="small"
                type="text"
                onClick={handleCancelClick}
              />
            )}
          </Flex>
        }
      >
        <Form className={className} style={style} {...mergedFormProps} form={formInst}>
          {children}
        </Form>
      </Drawer>
    </>
  );
};

export default DrawerForm;
