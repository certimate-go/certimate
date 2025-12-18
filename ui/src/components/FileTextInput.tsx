import { type ChangeEvent, useContext, useRef } from "react";
import { useTranslation } from "react-i18next";
import { IconFileImport } from "@tabler/icons-react";
import { Button, type ButtonProps, Input } from "antd";
import DisabledContext from "antd/es/config-provider/DisabledContext";
import { type TextAreaProps } from "antd/es/input/TextArea";

import { readFileAsText } from "@/utils/file";

export interface FileTextInputProps extends Omit<TextAreaProps, "onChange"> {
  accept?: string;
  uploadButtonProps?: Omit<ButtonProps, "disabled" | "onClick">;
  uploadText?: string;
  onChange?: (value: string) => void;
}

const FileTextInput = ({ className, style, accept, disabled, readOnly, uploadText, uploadButtonProps, onChange, ...props }: FileTextInputProps) => {
  const { t } = useTranslation();

  const injectedDisabled = useContext(DisabledContext);
  const mergedDisabled = disabled ?? injectedDisabled;

  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleButtonClick = () => {
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  const handleFileChange = async (e: ChangeEvent<HTMLInputElement>) => {
    const { files } = e.target as HTMLInputElement;
    if (files?.length) {
      const value = await readFileAsText(files[0]);
      onChange?.(value);
    }
  };

  return (
    <div className={className} style={style}>
      <div className="flex flex-col items-center gap-2">
        <Input.TextArea {...props} disabled={mergedDisabled} readOnly={readOnly} onChange={(e) => onChange?.(e.target.value)} />
        {!readOnly && (
          <>
            <Button {...uploadButtonProps} block disabled={mergedDisabled} icon={<IconFileImport size="1.25em" />} onClick={handleButtonClick}>
              {uploadText ?? t("common.text.import_from_file")}
            </Button>
            <input ref={fileInputRef} type="file" style={{ display: "none" }} accept={accept} onChange={handleFileChange} />
          </>
        )}
      </div>
    </div>
  );
};

export default FileTextInput;
