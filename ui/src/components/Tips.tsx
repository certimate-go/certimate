import { IconBulb } from "@tabler/icons-react";
import { Alert, Flex, Typography, theme } from "antd";

export interface TipsProps {
  className?: string;
  style?: React.CSSProperties;
  message: React.ReactNode;
}

const Tips = ({ className, style, message }: TipsProps) => {
  const { token: themeToken } = theme.useToken();

  return (
    <Alert
      className={className}
      style={style}
      title={
        <Flex gap="small">
          <div style={{ marginTop: "1px" }}>
            <IconBulb size={18} color={themeToken.colorInfo} />
          </div>
          <div style={{ flex: 1 }}>
            <Typography.Text>{message}</Typography.Text>
          </div>
        </Flex>
      }
      type="info"
    />
  );
};

export default Tips;
