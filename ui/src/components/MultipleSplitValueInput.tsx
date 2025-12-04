import { type ChangeEvent, useEffect } from "react";
import { IconList } from "@tabler/icons-react";
import { useControllableValue } from "ahooks";
import { Button, Form, Input, type InputProps, Space } from "antd";
import { nanoid } from "nanoid/non-secure";

import { useAntdForm } from "@/hooks";
import ModalForm from "./ModalForm";
import MultipleInput from "./MultipleInput";

type SplitOptions = {
  removeEmpty?: boolean;
  trimSpace?: boolean;
};

export interface MultipleSplitValueInputProps extends Omit<InputProps, "count" | "defaultValue" | "showCount" | "value" | "onChange"> {
  defaultValue?: string;
  maxCount?: number;
  minCount?: number;
  modalTitle?: React.ReactNode;
  modalWidth?: number | string;
  placeholderInModal?: string;
  showSortButton?: boolean;
  separator?: string;
  splitOptions?: SplitOptions;
  value?: string[];
  onChange?: (value: string) => void;
}

const DEFAULT_SEPARATOR = ";";

const MultipleSplitValueInput = ({
  className,
  style,
  size,
  separator: delimiter = DEFAULT_SEPARATOR,
  disabled,
  maxCount,
  minCount,
  modalTitle,
  modalWidth = "480px",
  placeholder,
  placeholderInModal,
  showSortButton = true,
  splitOptions = {},
  onClear,
  ...props
}: MultipleSplitValueInputProps) => {
  const [value, setValue] = useControllableValue<string>(props, {
    valuePropName: "value",
    defaultValuePropName: "defaultValue",
    trigger: "onChange",
  });

  const { form: formInst, formProps } = useAntdForm({
    name: "componentMultipleSplitValueInput_" + nanoid(),
    initialValues: { value: value?.split(delimiter) },
    onSubmit: (values) => {
      const temp = (values.value ?? []) as string[];
      if (splitOptions.trimSpace) {
        temp.map((e) => e.trim());
      }
      if (splitOptions.removeEmpty) {
        temp.filter((e) => !!e);
      }

      setValue(temp.join(delimiter));
    },
  });

  useEffect(() => {
    formInst.setFieldValue("value", value?.split(delimiter));
  }, [delimiter, value]);

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    setValue(e.target.value);
  };

  const handleClear = () => {
    setValue("");
    onClear?.();
  };

  return (
    <div className={className} style={style}>
      <Space.Compact className="w-full">
        <Input {...props} disabled={disabled} placeholder={placeholder} size={size} value={value} onChange={handleChange} onClear={handleClear} />
        <ModalForm
          {...formProps}
          layout="vertical"
          form={formInst}
          modalProps={{ destroyOnHidden: true }}
          title={modalTitle}
          trigger={
            <Button className="px-2" disabled={disabled} size={size}>
              <IconList size="1.25em" />
            </Button>
          }
          validateTrigger="onSubmit"
          width={modalWidth}
        >
          <Form.Item name="value" noStyle>
            <MultipleInput minCount={minCount} maxCount={maxCount} placeholder={placeholderInModal ?? placeholder} showSortButton={showSortButton} />
          </Form.Item>
        </ModalForm>
      </Space.Compact>
    </div>
  );
};

export default MultipleSplitValueInput;
