import { useMemo, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useMount, useSize } from "ahooks";

import { type SelectProps } from "antd";

import { useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";

type Provider = { type: string; name: string };

export interface SharedSelectProps<T extends Provider>
  extends Omit<SelectProps, "filterOption" | "filterSort" | "labelRender" | "options" | "optionFilterProp" | "optionLabelProp" | "optionRender"> {
  className?: string;
  style?: React.CSSProperties;
  onFilter?: (value: string, option: T) => boolean;
}

export const useSelectDataSource = <T extends Provider>({
  dataSource,
  filters,
  deps = [],
}: {
  dataSource: T[];
  filters?: Array<(value: string, option: T) => boolean>;
  deps?: React.DependencyList;
}) => {
  const { accesses, fetchAccesses } = useAccessesStore(useZustandShallowSelector(["accesses", "fetchAccesses"]));
  useMount(() => {
    fetchAccesses(false);
  });

  const filteredDataSource = useMemo(() => {
    return dataSource.filter((provider) => {
      if (filters) {
        for (const filter of filters) {
          if (!filter) continue;
          if (!filter(provider.type, provider)) return false;
        }
      }

      return true;
    });
  }, [dataSource, filters, deps]);

  const availableDataSource = useMemo(() => {
    return filteredDataSource.filter((provider) => {
      return accesses.some((access) => {
        if ("builtin" in provider && provider.builtin) return true;
        if ("provider" in provider) return access.provider === provider.provider;
        return access.provider === provider.type;
      });
    });
  }, [accesses, filteredDataSource, deps]);

  const unavailableDataSource = useMemo(() => {
    return filteredDataSource.filter((item) => !availableDataSource.includes(item));
  }, [filteredDataSource, availableDataSource, deps]);

  return {
    raw: dataSource,
    filtered: filteredDataSource,
    available: availableDataSource,
    unavailable: unavailableDataSource,
  };
};

export interface SharedPickerProps<T extends Provider> {
  className?: string;
  style?: React.CSSProperties;
  autoFocus?: boolean;
  gap?: number | "small" | "middle" | "large";
  placeholder?: string;
  showSearch?: boolean;
  onFilter?: (value: string, option: T) => boolean;
  onSelect?: (value: string) => void;
}

export const usePickerWrapperCols = (width: number) => {
  const wrapperElRef = useRef<HTMLDivElement>(null);
  const wrapperSize = useSize(wrapperElRef);

  const columns = useMemo(() => {
    const wWidth = wrapperSize?.width ?? document.body.clientWidth - 256;
    const wCols = Math.floor(wWidth / width);
    return Math.min(9, Math.max(1, wCols));
  }, [wrapperSize?.width, width]);

  return {
    wrapperElRef,
    cols: columns,
  };
};

export const usePickerDataSource = <T extends Provider>({
  dataSource,
  filters,
  keyword,
  deps = [],
}: {
  dataSource: T[];
  filters?: Array<(value: string, option: T) => boolean>;
  keyword?: string;
  deps?: React.DependencyList;
}) => {
  const { t } = useTranslation();

  const { accesses, fetchAccesses } = useAccessesStore(useZustandShallowSelector(["accesses", "fetchAccesses"]));
  useMount(() => {
    fetchAccesses(false);
  });

  const filteredDataSource = useMemo(() => {
    return dataSource
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
          const value = keyword.toLowerCase();
          return provider.type.toLowerCase().includes(value) || t(provider.name).toLowerCase().includes(value);
        }

        return true;
      });
  }, [dataSource, filters, keyword, deps]);

  const availableDataSource = useMemo(() => {
    return filteredDataSource.filter((provider) => {
      return accesses.some((access) => {
        if ("builtin" in provider && provider.builtin) return true;
        if ("provider" in provider) return access.provider === provider.provider;
        return access.provider === provider.type;
      });
    });
  }, [accesses, filteredDataSource, deps]);

  const unavailableDataSource = useMemo(() => {
    return filteredDataSource.filter((item) => !availableDataSource.includes(item));
  }, [filteredDataSource, availableDataSource, deps]);

  return {
    raw: dataSource,
    filtered: filteredDataSource,
    available: availableDataSource,
    unavailable: unavailableDataSource,
  };
};
