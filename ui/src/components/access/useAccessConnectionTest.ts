import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { App, type FormInstance } from "antd";

import { type AccessTestCapabilities, type AccessTestRequest, type AccessTestResult, fetchAccessTestCapabilities, testAccess } from "@/api/accesses";
import { usePluginCatalogStore } from "@/stores/pluginCatalog";
import { unwrapErrMsg } from "@/utils/error";
import { applyTrimmedFormValues } from "@/utils/form";

import { type AccessFormModes } from "./AccessForm";

export interface UseAccessConnectionTestOptions {
  enabled?: boolean;
  formInst: FormInstance;
  mode: AccessFormModes;
  provider?: string;
  recordId?: string;
}

export interface UseAccessConnectionTestReturns {
  result?: AccessTestResult;
  run: () => Promise<void>;
  reset: () => void;
  notifyConnectionConfigChanged: () => void;
  supported: boolean;
  testing: boolean;
}

const useAccessConnectionTest = ({ enabled = true, formInst, mode, provider, recordId }: UseAccessConnectionTestOptions): UseAccessConnectionTestReturns => {
  const { t } = useTranslation();
  const { notification } = App.useApp();

  const catalogVersion = usePluginCatalogStore((state) => state.version);

  const [capabilitiesState, setCapabilitiesState] = useState<{ provider: string; data?: AccessTestCapabilities }>();
  const [testing, setTesting] = useState(false);
  const [result, setResult] = useState<AccessTestResult>();

  const seqRef = useRef(0);
  const revRef = useRef(0);
  const configChangedRef = useRef(false);
  const testingRef = useRef(false);

  useEffect(() => {
    if (!enabled || !provider) return;

    let cancelled = false;
    void fetchAccessTestCapabilities(provider).then(
      (data) => {
        if (!cancelled) setCapabilitiesState({ provider, data });
      },
      () => {
        if (!cancelled) setCapabilitiesState({ provider });
      }
    );
    return () => {
      cancelled = true;
    };
  }, [enabled, provider, catalogVersion]);

  const capabilitiesStateMatched = enabled && capabilitiesState != null && capabilitiesState.provider === provider ? capabilitiesState : undefined;
  const capabilities = capabilitiesStateMatched?.data;

  const notifyConnectionConfigChanged = () => {
    revRef.current += 1;
    configChangedRef.current = true;
    setResult(void 0);
  };

  const reset = () => {
    seqRef.current += 1;
    revRef.current += 1;
    configChangedRef.current = false;
    testingRef.current = false;
    setTesting(false);
    setResult(void 0);
  };

  const run = async () => {
    if (!enabled || !provider || testingRef.current) {
      return;
    }

    const useDraft = mode === "create" || configChangedRef.current || !recordId;
    applyTrimmedFormValues(formInst);
    const formValues = formInst.getFieldsValue(true);
    const request: AccessTestRequest = useDraft
      ? { source: "draft", provider, config: (formValues?.config as Record<string, unknown> | undefined) ?? {} }
      : { source: "stored", accessId: recordId };

    const seq = ++seqRef.current;
    const rev = revRef.current;

    testingRef.current = true;
    setTesting(true);
    setResult(void 0);

    try {
      const data = await testAccess(request);
      if (seq !== seqRef.current || rev !== revRef.current) {
        return;
      }
      setResult(data);
    } catch (err) {
      if (seq !== seqRef.current || rev !== revRef.current) {
        return;
      }
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });
    } finally {
      if (seq === seqRef.current) {
        testingRef.current = false;
        setTesting(false);
      }
    }
  };

  const supported = capabilities?.supported === true;

  return {
    supported,
    result,
    testing,
    notifyConnectionConfigChanged,
    reset,
    run,
  };
};

export default useAccessConnectionTest;
