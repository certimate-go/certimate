import { useEffect, useMemo } from "react";
import { Form } from "antd";

import { useZustandShallowSelector } from "@/hooks";
import { useSystemEnvironmentStore } from "@/stores/system";

import AccessConfigFieldsProviderSSH, {
  getAccessConfigFieldsProviderSSHInitialValues,
} from "./AccessConfigFieldsProviderSSH";
import { useFormNestedFieldsContext } from "./_context";

const FALLBACK_HOST = "host.docker.internal";

const AccessConfigFieldsProviderDockerHost = ({ disabled }: { disabled?: boolean }) => {
  const formInst = Form.useFormInstance();
  const { parentNamePath } = useFormNestedFieldsContext();

  const { environment, fetchEnvironment, loadedEnvironment } = useSystemEnvironmentStore(
    useZustandShallowSelector(["environment", "fetchEnvironment", "loadedEnvironment"])
  );

  useEffect(() => {
    if (!loadedEnvironment) {
      fetchEnvironment(false);
    }
  }, [fetchEnvironment, loadedEnvironment]);

  const resolvedHost = useMemo(() => {
    if (environment?.dockerHost.reachable && environment.dockerHost.address) {
      return environment.dockerHost.address;
    }
    return FALLBACK_HOST;
  }, [environment]);

  useEffect(() => {
    const defaultHost = getAccessConfigFieldsProviderSSHInitialValues()?.host;
    const currentHost = formInst.getFieldValue([parentNamePath, "host"]);

    if (!currentHost || currentHost === defaultHost || currentHost === FALLBACK_HOST) {
      formInst.setFieldValue([parentNamePath, "host"], resolvedHost);
    }
  }, [formInst, parentNamePath, resolvedHost]);

  return <AccessConfigFieldsProviderSSH disabled={disabled} hideJumpServers hostDisabled />;
};

export default AccessConfigFieldsProviderDockerHost;

