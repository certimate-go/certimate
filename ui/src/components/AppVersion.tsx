import { Badge, Typography } from "antd";

import { APP_DOWNLOAD_URL, APP_VERSION } from "@/domain/app";
import { useVersionChecker } from "@/hooks";

export interface AppVersionLinkButtonProps {
  className?: string;
  style?: React.CSSProperties;
}

const AppVersionLinkButton = ({ className, style }: AppVersionLinkButtonProps) => {
  return (
    <AppVersionBadge>
      <Typography.Link className={className} style={style} type="secondary" href={APP_DOWNLOAD_URL} target="_blank">
        {APP_VERSION}
      </Typography.Link>
    </AppVersionBadge>
  );
};

export interface AppVersionBadgeProps {
  className?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
}

const AppVersionBadge = ({ className, style, children }: AppVersionBadgeProps) => {
  const { hasUpdate } = useVersionChecker();

  return (
    <Badge
      className={className}
      style={style}
      styles={{
        indicator: { transform: "scale(0.75) translate(50%, -85%)" },
      }}
      count={hasUpdate ? "NEW" : void 0}
    >
      {children}
    </Badge>
  );
};

export default {
  LinkButton: AppVersionLinkButton,
  Badge: AppVersionBadge,
};
