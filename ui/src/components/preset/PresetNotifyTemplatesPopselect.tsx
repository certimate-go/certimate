import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { IconMoodEmpty } from "@tabler/icons-react";
import { useMount } from "ahooks";
import { Dropdown, type DropdownProps } from "antd";

import { useZustandShallowSelector } from "@/hooks";
import { useNotifyTemplatesStore } from "@/stores/settings";

type PresetTemplate = {
  subject: string;
  message: string;
};

export interface PresetNotifyTemplatesPopselectProps extends Omit<DropdownProps, "menu"> {
  options?: NonNullable<DropdownProps["menu"]>["items"];
  onSelect?: (value: string, template?: PresetTemplate | undefined) => void;
}

const PresetNotifyTemplatesPopselect = ({ className, options, onSelect, ...props }: PresetNotifyTemplatesPopselectProps) => {
  const { t } = useTranslation();

  const { templates, fetchTemplates } = useNotifyTemplatesStore(useZustandShallowSelector(["templates", "fetchTemplates"]));
  useMount(() => {
    fetchTemplates(false);
  });

  const menuItems = useMemo(() => {
    type MenuItem = NonNullable<NonNullable<DropdownProps["menu"]>["items"]>[number];
    const temp: MenuItem[] = [];

    if (!options?.length && !templates?.length) {
      temp.push({
        key: "nodata",
        label: t("common.text.nodata"),
        icon: (
          <span className="anticon scale-125">
            <IconMoodEmpty size="1em" />
          </span>
        ),
        disabled: true,
      });
      return temp;
    }

    if (options?.length) {
      temp.push(
        ...options.map((option) => {
          return {
            ...option!,
            onClick: (e: any) => {
              if ("onClick" in option!) {
                option.onClick?.(e);
              }

              onSelect?.(String(option!.key!));
            },
          };
        })
      );
    }

    if (templates?.length) {
      temp.push({
        key: "custom",
        label: t("preset.dropdown.option_group.custom"),
        children: templates.map((template) => ({
          key: `custom/${template.name}`,
          label: template.name,
          onClick: () => {
            onSelect?.(template.name, template);
          },
        })),
      });
    }

    return temp;
  }, [options, templates, onSelect]);

  return <Dropdown className={className} menu={{ items: menuItems }} {...props} />;
};

export default PresetNotifyTemplatesPopselect;
