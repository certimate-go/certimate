import { useControllableValue } from "ahooks";
import { Form, type FormProps, Modal, type ModalProps } from "antd";

import { useAntdForm, useTriggerElement } from "@/hooks";

export interface ModalFormProps<T extends NonNullable<unknown> = any> extends Omit<FormProps<T>, "title" | "onFinish"> {
  className?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
  cancelButtonProps?: ModalProps["cancelButtonProps"];
  cancelText?: ModalProps["cancelText"];
  defaultOpen?: boolean;
  modalProps?: Omit<
    ModalProps,
    | "cancelButtonProps"
    | "cancelText"
    | "confirmLoading"
    | "defaultOpen"
    | "forceRender"
    | "okButtonProps"
    | "okText"
    | "okType"
    | "open"
    | "title"
    | "width"
    | "onCancel"
    | "onOk"
    | "onOpenChange"
  >;
  okButtonProps?: ModalProps["okButtonProps"];
  okText?: ModalProps["okText"];
  open?: boolean;
  title?: ModalProps["title"];
  trigger?: React.ReactNode;
  width?: ModalProps["width"];
  onFinish?: (values: T) => unknown | Promise<unknown>;
  onOpenChange?: (open: boolean) => void;
}

const ModalForm = <T extends NonNullable<unknown> = any>({
  className,
  style,
  children,
  cancelButtonProps,
  cancelText,
  form,
  modalProps,
  okButtonProps,
  okText,
  title,
  trigger,
  width,
  onFinish,
  ...props
}: ModalFormProps<T>) => {
  const [open, setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });

  const triggerEl = useTriggerElement(trigger, { onClick: () => setOpen(true) });

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
    clearOnDestroy: modalProps?.destroyOnHidden ? true : void 0,
    ...formProps,
    ...props,
  };

  const mergedModalProps: ModalProps = {
    ...modalProps,
    afterClose: () => {
      if (!mergedFormProps.preserve) {
        formInst.resetFields();
      }

      modalProps?.afterClose?.();
    },
  };

  const handleOkClick = async () => {
    // 提交表单返回 Promise.reject 时不关闭 Modal
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

      <Modal
        {...mergedModalProps}
        cancelButtonProps={cancelButtonProps}
        cancelText={cancelText}
        confirmLoading={formPending}
        forceRender
        okButtonProps={okButtonProps}
        okText={okText}
        okType="primary"
        open={open}
        title={title}
        width={width}
        onOk={handleOkClick}
        onCancel={handleCancelClick}
      >
        <div className="py-3">
          <Form className={className} style={style} {...mergedFormProps} form={formInst}>
            {children}
          </Form>
        </div>
      </Modal>
    </>
  );
};

export default ModalForm;
