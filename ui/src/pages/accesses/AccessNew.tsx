import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import { App, Button, Flex, Form } from "antd";

import AccessForm, { type AccessFormUsages } from "@/components/access/AccessForm";
import AccessProviderPicker from "@/components/provider/AccessProviderPicker";
import Show from "@/components/Show";
import { type AccessModel } from "@/domain/access";
import { ACCESS_USAGES } from "@/domain/provider";
import { useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";
import { unwrapErrMsg } from "@/utils/error";

const AccessNew = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const { t } = useTranslation();

  const { notification } = App.useApp();

  const { createAccess } = useAccessesStore(useZustandShallowSelector(["createAccess"]));

  const providerUsage = useMemo(() => searchParams.get("usage") as AccessFormUsages, [searchParams]);
  const providerFilter = AccessForm.useProviderFilterByUsage(providerUsage);

  const [formInst] = Form.useForm();
  const [formPending, setFormPending] = useState(false);

  const fieldProvider = Form.useWatch<string>("provider", { form: formInst, preserve: true });

  const handleProviderPick = (value: string) => {
    formInst.setFieldValue("provider", value);
  };

  const handleSubmitClick = async () => {
    let formValues: AccessModel;

    setFormPending(true);
    try {
      formValues = await formInst.validateFields();
      formValues.reserve = providerUsage === "ca" ? "ca" : providerUsage === "notification" ? "notif" : void 0;
    } catch (err) {
      setFormPending(false);
      throw err;
    }

    try {
      await createAccess(formValues);

      navigate(`/accesses?usage=${providerUsage}`, { replace: true });
    } catch (err) {
      notification.error({ title: t("common.text.request_error"), description: unwrapErrMsg(err) });

      throw err;
    } finally {
      setFormPending(false);
    }
  };

  const handleCancelClick = () => {
    formInst.resetFields();
  };

  return (
    <div className="px-6 py-4">
      <div className="container">
        <h1>{t("access.new.title")}</h1>
        <p className="text-base text-gray-500">{t("access.new.subtitle")}</p>
      </div>

      <div className="container">
        <Show when={!fieldProvider}>
          <AccessProviderPicker
            autoFocus
            gap="large"
            placeholder={t("access.form.provider.search.placeholder")}
            showOptionTags={
              providerUsage == null ||
              (providerUsage === "dns-hosting" ? { ["builtin"]: true, [ACCESS_USAGES.DNS]: true, [ACCESS_USAGES.HOSTING]: true } : { ["builtin"]: true })
            }
            showSearch
            onFilter={providerFilter}
            onSelect={handleProviderPick}
          />
        </Show>

        <div style={{ display: fieldProvider ? "block" : "none" }}>
          <div className="md:max-w-160">
            <AccessForm form={formInst} disabled={formPending} mode="create" usage={providerUsage} />
          </div>
          <Flex gap="small">
            <Button type="primary" onClick={handleSubmitClick}>
              {t("common.button.submit")}
            </Button>
            <Button onClick={handleCancelClick}>{t("common.button.cancel")}</Button>
          </Flex>
        </div>
      </div>
    </div>
  );
};

export default AccessNew;
