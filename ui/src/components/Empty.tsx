import { useTranslation } from "react-i18next";
import { Typography, theme } from "antd";

import Show from "./Show";

export interface EmptyProps {
  className?: string;
  style?: React.CSSProperties;
  title?: React.ReactNode;
  description?: React.ReactNode;
  extra?: React.ReactNode;
  icon?: React.ReactNode;
}

const Empty = (props: EmptyProps) => {
  const { t } = useTranslation();

  const { className, style, title = t("common.text.nodata"), description, extra, icon } = props;

  const { token: themeToken } = theme.useToken();

  const isPrimitive = (node: React.ReactNode): node is string | number | boolean | null => {
    return typeof node === "string" || typeof node === "number" || typeof node === "boolean" || node == null;
  };

  return (
    <div className={className} style={style}>
      <div className="relative w-full overflow-hidden">
        <div className="relative top-0 left-0 z-1 flex size-full py-4">
          <div className="relative mx-auto w-full max-w-lg">
            <div className="flex flex-col gap-2 text-center">
              <Show when={!!icon}>
                <div className="mx-auto">
                  <div className="flex size-12 items-center justify-center overflow-hidden rounded-md border text-gray-600 shadow-sm dark:text-gray-200">
                    {icon}
                  </div>
                </div>
              </Show>
              <div className="my-2">
                <Show when={!!title}>
                  <div className="pb-2 text-xl" style={{ color: themeToken.colorTextLabel }}>
                    {title}
                  </div>
                </Show>
                <Show when={!!description}>
                  {isPrimitive(description) ? (
                    <Typography.Text type="secondary">{description}</Typography.Text>
                  ) : (
                    <div style={{ color: themeToken.colorTextSecondary }}>{description}</div>
                  )}
                </Show>
              </div>
              <Show when={!!extra}>
                <div>{extra}</div>
              </Show>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Empty;
