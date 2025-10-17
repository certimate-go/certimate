import { useCallback } from "react";
import { useLocalStorageState } from "ahooks";

const LOCAL_STORAGE_KEY = "certimate-ui-appsettings";
if (!localStorage.getItem(LOCAL_STORAGE_KEY)) {
  localStorage.setItem(LOCAL_STORAGE_KEY, "{}");
}

type AppSettings = {
  /**
   * 每页显示的默认条目数。
   */
  defaultPerPage?: number;
  /**
   * 工作流默认布局。
   */
  defaultWorkflowLayout?: "horizontal" | "vertical";
};

export type UseAppSettingsReturns = {
  appSettings: AppSettings;
  setAppSettings: (value: AppSettings) => void;
  resetAppSettings: () => void;
};

/**
 * 获取并设置当前应用全局配置。
 * @returns {UseAppSettingsReturns}
 */
const useAppSettings = (): UseAppSettingsReturns => {
  const [state, setState] = useLocalStorageState<AppSettings>(LOCAL_STORAGE_KEY, {
    defaultValue: {},
  });
  state.defaultPerPage ??= 15;
  state.defaultWorkflowLayout ??= "vertical";

  const resetState = useCallback(() => {
    localStorage.removeItem(LOCAL_STORAGE_KEY);
    setState({});
  }, [setState]);

  return {
    appSettings: state,
    setAppSettings: setState,
    resetAppSettings: resetState,
  };
};

export default useAppSettings;
