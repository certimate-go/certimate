import { useRef } from "react";
import { useTranslation } from "react-i18next";
import { useMount, useSize } from "ahooks";

import { type SelectProps } from "antd";

import { useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";
import { matchSearchString } from "@/utils/search";

type Provider = { type: string; name: string };

const isProviderAccessSatisfied = <T extends Provider>(provider: T, accesses: { provider: string }[]) => {
  if ("accessOptional" in provider && provider.accessOptional) return true;
  return accesses.some((access) => {
    if ("builtin" in provider && provider.builtin) return true;
    if ("provider" in provider) return access.provider === provider.provider;
    return access.provider === provider.type;
  });
};

export interface SharedSelectProps<T extends Provider> extends Omit<SelectProps, "labelRender" | "options" | "optionLabelProp" | "optionRender"> {
  className?: string;
  style?: React.CSSProperties;
  onFilter?: (value: string, option: T) => boolean;
}

export const useSelectDataSource = <T extends Provider>({
  dataSource,
  filters,
}: {
  dataSource: T[];
  filters?: Array<(value: string, option: T) => boolean>;
  deps?: React.DependencyList;
}) => {
  const { accesses, fetchAccesses } = useAccessesStore(useZustandShallowSelector(["accesses", "fetchAccesses"]));
  useMount(() => {
    fetchAccesses(false);
  });

  const dataSourceFiltered = dataSource.filter((provider) => {
    if (filters) {
      for (const filter of filters) {
        if (!filter) continue;
        if (!filter(provider.type, provider)) return false;
      }
    }

    return true;
  });
  const dataSourceAvailable = dataSourceFiltered.filter((provider) => isProviderAccessSatisfied(provider, accesses));
  const dataSourceUnavailable = dataSourceFiltered.filter((item) => !dataSourceAvailable.includes(item));

  return {
    raw: dataSource,
    filtered: dataSourceFiltered,
    available: dataSourceAvailable,
    unavailable: dataSourceUnavailable,
  };
};

export interface SharedPickerProps<T extends Provider> {
  className?: string;
  style?: React.CSSProperties;
  gap?: number | "small" | "medium" | "large";
  placeholder?: string;
  showSearch?: boolean;
  onFilter?: (value: string, option: T) => boolean;
  onSelect?: (value: string) => void;
}

export const usePickerWrapperCols = (width: number) => {
  const wrapperElRef = useRef<HTMLDivElement>(null);
  const wrapperSize = useSize(wrapperElRef);

  const columns = (() => {
    const wWidth = wrapperSize?.width ?? document.body.clientWidth - 256;
    const wCols = Math.floor(wWidth / width);
    return Math.min(9, Math.max(1, wCols));
  })();

  return {
    wrapperElRef,
    cols: columns,
  };
};

export const usePickerDataSource = <T extends Provider>({
  dataSource,
  filters,
  keyword,
}: {
  dataSource: T[];
  filters?: Array<(value: string, option: T) => boolean>;
  keyword?: string;
}) => {
  const { t } = useTranslation();

  const { accesses, fetchAccesses } = useAccessesStore(useZustandShallowSelector(["accesses", "fetchAccesses"]));
  useMount(() => {
    fetchAccesses(false);
  });

  const dataSourceFiltered = dataSource
    .filter((provider) => {
      if (filters) {
        for (const filter of filters) {
          if (!filter) continue;
          if (!filter(provider.type, provider)) return false;
        }
      }

      return true;
    })
    .filter((provider) => {
      if (keyword) {
        return matchSearchString(keyword, provider.type) || matchSearchString(keyword, t(provider.name));
      }

      return true;
    });
  const dataSourceAvailable = dataSourceFiltered.filter((provider) => isProviderAccessSatisfied(provider, accesses));
  const dataSourceUnavailable = dataSourceFiltered.filter((item) => !dataSourceAvailable.includes(item));

  return {
    raw: dataSource,
    filtered: dataSourceFiltered,
    available: dataSourceAvailable,
    unavailable: dataSourceUnavailable,
  };
};
