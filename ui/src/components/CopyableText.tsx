import { CopyToClipboard } from "react-copy-to-clipboard";
import { useTranslation } from "react-i18next";
import { App, Button } from "antd";

export interface CopyableTextProps {
  className?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
  text?: string;
}

const CopyableText = ({ className, style, children, text }: CopyableTextProps) => {
  const { t } = useTranslation();

  const { message } = App.useApp();

  return (
    <CopyToClipboard
      text={text ?? (children as string)}
      onCopy={() => {
        message.success(t("common.text.copied"));
      }}
    >
      <Button className={className} style={style} size="small" type="text">
        {children}
      </Button>
    </CopyToClipboard>
  );
};

export default CopyableText;
