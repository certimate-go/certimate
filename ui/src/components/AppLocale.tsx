import { memo } from "react";
import { useTranslation } from "react-i18next";
import { IconLanguage, type IconProps } from "@tabler/icons-react";
import { Dropdown, type DropdownProps, Typography } from "antd";

import { IconLanguageEnZh, IconLanguageZhEn } from "@/components/icons";
import { localeNames, localeResources } from "@/i18n";
import { mergeCls } from "@/utils/css";

export const useAppLocaleMenuItems = () => {
  const { i18n } = useTranslation();

  const items = Object.keys(i18n.store.data).map((key) => {
    return {
      key: key as string,
      label: i18n.store.data[key].name as string,
      onClick: () => {
        if (key !== i18n.language) {
          i18n.changeLanguage(key);
          window.location.reload();
        }
      },
    };
  });

  return items;
};

export type AppLocaleDropdownProps = {
  children?: React.ReactNode;
  trigger?: DropdownProps["trigger"];
};

const AppLocaleDropdown = (props: AppLocaleDropdownProps) => {
  const { children, trigger = ["click"] } = props;

  const items = useAppLocaleMenuItems();

  return (
    <Dropdown menu={{ items }} trigger={trigger}>
      {children}
    </Dropdown>
  );
};

export type AppLocaleIconProps = IconProps;

const AppLocaleIcon = (props: AppLocaleIconProps) => {
  const { i18n } = useTranslation();

  return i18n.language === localeNames.EN ? (
    <IconLanguageEnZh {...props} />
  ) : i18n.language === localeNames.ZH ? (
    <IconLanguageZhEn {...props} />
  ) : (
    <IconLanguage {...props} />
  );
};

export type AppLocaleLinkButtonProps = {
  className?: string;
  style?: React.CSSProperties;
  showIcon?: boolean;
};

const AppLocaleLinkButton = (props: AppLocaleLinkButtonProps) => {
  const { className, style, showIcon = true } = props;

  const { t } = useTranslation();
  const { i18n } = useTranslation();

  return (
    <AppLocaleDropdown trigger={["click", "hover"]}>
      <Typography.Text className={mergeCls("cursor-pointer", className)} style={style} type="secondary">
        <div className="flex items-center justify-center space-x-1">
          {showIcon ? <AppLocaleIcon size="1em" /> : <></>}
          <span>{String(localeResources[i18n.language]?.name ?? t("common.menu.locale"))}</span>
        </div>
      </Typography.Text>
    </AppLocaleDropdown>
  );
};

export default {
  Dropdown: memo(AppLocaleDropdown),
  Icon: memo(AppLocaleIcon),
  LinkButton: memo(AppLocaleLinkButton),
};
