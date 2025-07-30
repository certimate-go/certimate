import { useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import { App, Button, Flex } from "antd";

import AccessForm, { type AccessFormInstance, type AccessFormProps } from "@/components/access/AccessForm";
import AccessProviderPicker from "@/components/provider/AccessProviderPicker";
import Show from "@/components/Show";
import { type AccessModel } from "@/domain/access";
import { ACCESS_USAGES, type AccessProvider } from "@/domain/provider";
import { useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";
import { getErrMsg } from "@/utils/error";

const AccessNew = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const { t } = useTranslation();

  const { notification } = App.useApp();

  const { createAccess } = useAccessesStore(useZustandShallowSelector(["createAccess"]));

  const formRef = useRef<AccessFormInstance>(null);
  const [formModel, setFormModel] = useState<MaybeModelRecord<AccessModel>>({});
  const [formPending, setFormPending] = useState(false);

  const providerUsage = useMemo(() => {
    return searchParams.get("usage") as AccessFormProps["usage"];
  }, [searchParams]);
  const providerFilter = useMemo(() => {
    switch (providerUsage) {
      case "dns":
        return (_: string, record: AccessProvider) => record.usages.includes(ACCESS_USAGES.DNS);
      case "hosting":
        return (_: string, record: AccessProvider) => record.usages.includes(ACCESS_USAGES.HOSTING);
      case "dns-hosting":
        return (_: string, record: AccessProvider) => record.usages.includes(ACCESS_USAGES.DNS) || record.usages.includes(ACCESS_USAGES.HOSTING);
      case "ca":
        return (_: string, record: AccessProvider) => record.usages.includes(ACCESS_USAGES.CA);
      case "notification":
        return (_: string, record: AccessProvider) => record.usages.includes(ACCESS_USAGES.NOTIFICATION);
      default:
        console.warn(`[certimate] unsupported provider usage: '${providerUsage}'`);
    }

    return () => false;
  }, [providerUsage]);

  const handleProviderPick = (value: string) => {
    setFormModel((prev) => {
      return { ...prev, provider: value, config: undefined };
    });
  };

  const handleSubmitClick = async () => {
    setFormPending(true);
    try {
      await formRef.current!.validateFields();
    } catch (err) {
      setFormPending(false);
      throw err;
    }

    try {
      const values: AccessModel = formRef.current!.getFieldsValue();
      await createAccess(values);

      navigate(`/accesses?usage=${providerUsage}`, { replace: true });
    } catch (err) {
      notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

      throw err;
    } finally {
      setFormPending(false);
    }
  };

  const handleCancelClick = () => {
    setFormModel((prev) => {
      return { ...prev, provider: undefined, config: undefined, reserve: undefined };
    });
  };

  return (
    <div className="px-6 py-4">
      <div className="container">
        <h1>{t("access.new.title")}</h1>
        <p className="text-base text-gray-500">{t("access.new.subtitle")}</p>
      </div>

      <div className="container">
        <Show
          when={!!formModel?.provider}
          fallback={
            <AccessProviderPicker
              autoFocus
              gap="large"
              placeholder={t("access.form.provider.search.placeholder")}
              showOptionTags={providerUsage == null || (providerUsage === "dns-hosting" ? { [ACCESS_USAGES.DNS]: true, [ACCESS_USAGES.HOSTING]: true } : false)}
              onFilter={providerFilter}
              onSelect={handleProviderPick}
            />
          }
        >
          <AccessForm
            className="md:max-w-160"
            ref={formRef}
            initialValues={formModel}
            disabled={formPending}
            mode="create"
            usage={providerUsage}
            onValuesChange={(values) => setFormModel(values)}
          />
          <Flex gap="small">
            <Button type="primary" onClick={handleSubmitClick}>
              {t("common.button.submit")}
            </Button>
            <Button onClick={handleCancelClick}>{t("common.button.cancel")}</Button>
          </Flex>
        </Show>
      </div>
    </div>
  );
};

export default AccessNew;
