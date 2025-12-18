import { useContext, useMemo, useRef } from "react";
import { json } from "@codemirror/lang-json";
import { yaml } from "@codemirror/lang-yaml";
import { StreamLanguage } from "@codemirror/language";
import { powerShell } from "@codemirror/legacy-modes/mode/powershell";
import { shell } from "@codemirror/legacy-modes/mode/shell";
import { basicSetup } from "@uiw/codemirror-extensions-basic-setup";
import { vscodeDark, vscodeLight } from "@uiw/codemirror-theme-vscode";
import CodeMirror, { EditorView, type ReactCodeMirrorProps, type ReactCodeMirrorRef } from "@uiw/react-codemirror";
import { useFocusWithin, useHover } from "ahooks";
import { theme } from "antd";
import DisabledContext from "antd/es/config-provider/DisabledContext";

import { useBrowserTheme } from "@/hooks";
import { mergeCls } from "@/utils/css";

export interface CodeTextInputProps extends Omit<ReactCodeMirrorProps, "extensions" | "lang" | "theme"> {
  disabled?: boolean;
  language?: string | string[];
  lineNumbers?: boolean;
  lineWrapping?: boolean;
  readOnly?: boolean;
}

const CodeTextInput = ({ className, style, disabled, language, lineNumbers = true, lineWrapping = true, readOnly, ...props }: CodeTextInputProps) => {
  const { token: themeToken } = theme.useToken();

  const { theme: browserTheme } = useBrowserTheme();

  const injectedDisabled = useContext(DisabledContext);
  const mergedDisabled = disabled ?? injectedDisabled;

  const cmRef = useRef<ReactCodeMirrorRef>(null);
  const isFocusing = useFocusWithin(cmRef.current?.editor);
  const isHovering = useHover(cmRef.current?.editor);

  const cmTheme = useMemo(() => {
    if (browserTheme === "dark") {
      return vscodeDark;
    }
    return vscodeLight;
  }, [browserTheme]);

  const cmExtensions = useMemo(() => {
    const temp: NonNullable<ReactCodeMirrorProps["extensions"]> = [
      basicSetup({
        foldGutter: false,
        dropCursor: false,
        allowMultipleSelections: false,
        indentOnInput: false,
      }),
    ];

    if (lineWrapping) {
      temp.push(EditorView.lineWrapping);
    }

    const langs = Array.isArray(language) ? language : [language];
    langs.forEach((lang) => {
      switch (lang) {
        case "shell":
          temp.push(StreamLanguage.define(shell));
          break;
        case "json":
          temp.push(json());
          break;
        case "powershell":
          temp.push(StreamLanguage.define(powerShell));
          break;
        case "yaml":
          temp.push(yaml());
          break;
      }
    });

    return temp;
  }, [language, lineWrapping]);

  return (
    <div
      className={mergeCls("ant-input", className)}
      style={{
        ...style,
        paddingBlock: themeToken.Input?.paddingBlock,
        paddingInline: themeToken.Input?.paddingInline,
        fontSize: themeToken.Input?.inputFontSize,
        lineHeight: themeToken.lineHeight,
        color: mergedDisabled ? themeToken.colorTextDisabled : themeToken.colorText,
        backgroundColor: mergedDisabled
          ? themeToken.colorBgContainerDisabled
          : isFocusing
            ? (themeToken.Input?.activeBg ?? themeToken.colorBgContainer)
            : isHovering
              ? (themeToken.Input?.hoverBg ?? themeToken.colorBgContainer)
              : void 0,
        borderWidth: `${themeToken.lineWidth}px`,
        borderStyle: themeToken.lineType,
        borderColor: isFocusing
          ? (themeToken.Input?.activeBorderColor ?? themeToken.colorPrimaryActive)
          : isHovering
            ? (themeToken.Input?.hoverBorderColor ?? themeToken.colorPrimaryHover)
            : themeToken.colorBorder,
        borderRadius: `${themeToken.borderRadius}px`,
        boxShadow: isFocusing ? themeToken.Input?.activeShadow : void 0,
        overflow: "hidden",
      }}
    >
      <CodeMirror
        ref={cmRef}
        height="100%"
        style={{ height: "100%" }}
        {...props}
        basicSetup={{
          allowMultipleSelections: false,
          dropCursor: false,
          foldGutter: false,
          lineNumbers: lineNumbers,
          indentOnInput: false,
        }}
        extensions={cmExtensions}
        readOnly={readOnly || mergedDisabled}
        theme={cmTheme}
      />
    </div>
  );
};

export default CodeTextInput;
