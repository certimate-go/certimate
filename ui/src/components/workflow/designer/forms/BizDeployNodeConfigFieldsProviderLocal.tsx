import { getI18n, useTranslation } from "react-i18next";
import { IconBulb, IconChevronDown, IconDice6 } from "@tabler/icons-react";
import { Button, Divider, Form, type FormInstance, Input, Popover, Select, Space, Tooltip } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import CodeTextInput from "@/components/CodeTextInput";
import PresetScriptTemplatesPopselect from "@/components/preset/PresetScriptTemplatesPopselect";
import Show from "@/components/Show";
import Tips from "@/components/Tips";
import { CERTIFICATE_FORMATS } from "@/domain/certificate";
import { randomString } from "@/utils/random";

import { useFormNestedFieldsContext } from "./_context";

const FORMAT_PEM = CERTIFICATE_FORMATS.PEM;
const FORMAT_PFX = CERTIFICATE_FORMATS.PFX;
const FORMAT_JKS = CERTIFICATE_FORMATS.JKS;

const SHELLENV_SH = "sh" as const;
const SHELLENV_CMD = "cmd" as const;
const SHELLENV_POWERSHELL = "powershell" as const;

export const initPresetScript = (
  key: "sh_backup_files" | "ps_backup_files" | "sh_reload_nginx" | "ps_binding_iis" | "ps_binding_netsh" | "ps_binding_rdp",
  params?: {
    certPath?: string;
    certPathForServerOnly?: string;
    certPathForIntermediaOnly?: string;
    keyPath?: string;
    pfxPassword?: string;
    jksAlias?: string;
    jksKeypass?: string;
    jksStorepass?: string;
  }
) => {
  switch (key) {
    case "sh_backup_files":
      return `# 请将以下路径替换为实际值
cp "${params?.certPath || "<your-cert-path>"}" "${params?.certPath || "<your-cert-path>"}.bak" 2>/dev/null || :
cp "${params?.keyPath || "<your-key-path>"}" "${params?.keyPath || "<your-key-path>"}.bak" 2>/dev/null || :
      `.trim();

    case "ps_backup_files":
      return `# 请将以下路径替换为实际值
if (Test-Path -Path "${params?.certPath || "<your-cert-path>"}" -PathType Leaf) {
  Copy-Item -Path "${params?.certPath || "<your-cert-path>"}" -Destination "${params?.certPath || "<your-cert-path>"}.bak" -Force
}
if (Test-Path -Path "${params?.keyPath || "<your-key-path>"}" -PathType Leaf) {
  Copy-Item -Path "${params?.keyPath || "<your-key-path>"}" -Destination "${params?.keyPath || "<your-key-path>"}.bak" -Force
}
      `.trim();

    case "sh_reload_nginx":
      return `# *** 需要 root 权限 ***

sudo service nginx reload
      `.trim();

    case "ps_binding_iis":
      return `# *** 需要管理员权限 ***

# 请将以下变量替换为实际值
$pfxPath = "\${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_PATH}" # PFX 文件路径（与表单中保持一致）
$pfxPassword = "\${CERTIMATE_DEPLOYER_CMDVAR_PFX_PASSWORD}" # PFX 密码（与表单中保持一致）
$siteName = "<your-site-name>" # IIS 网站名称
$domain = "<your-domain-name>" # 域名
$ipaddr = "<your-binding-ip>"  # 绑定 IP，“*”表示所有 IP 绑定
$port = "<your-binding-port>"  # 绑定端口

# 导入证书到本地计算机的个人存储区
$cert = Import-PfxCertificate -FilePath "$pfxPath" -CertStoreLocation Cert:\\LocalMachine\\My -Password (ConvertTo-SecureString -String "$pfxPassword" -AsPlainText -Force) -Exportable
# 获取 Thumbprint
$thumbprint = $cert.Thumbprint
# 导入 WebAdministration 模块
Import-Module WebAdministration
# 检查是否已存在 HTTPS 绑定
$existingBinding = Get-WebBinding -Name "$siteName" -Protocol "https" -Port $port -HostHeader "$domain" -ErrorAction SilentlyContinue
if (!$existingBinding) {
  # 添加新的 HTTPS 绑定
  New-WebBinding -Name "$siteName" -Protocol "https" -Port $port -IPAddress "$ipaddr" -HostHeader "$domain"
}
# 获取绑定对象
$binding = Get-WebBinding -Name "$siteName" -Protocol "https" -Port $port -IPAddress "$ipaddr" -HostHeader "$domain"
# 绑定 SSL 证书
$binding.AddSslCertificate($thumbprint, "My")
# 删除目录下的证书文件
Remove-Item -Path "$pfxPath" -Force
      `.trim();

    case "ps_binding_netsh":
      return `# *** 需要管理员权限 ***

# 请将以下变量替换为实际值
$pfxPath = "\${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_PATH}" # PFX 文件路径（与表单中保持一致）
$pfxPassword = "\${CERTIMATE_DEPLOYER_CMDVAR_PFX_PASSWORD}" # PFX 密码（与表单中保持一致）
$ipaddr = "<your-binding-ip>"  # 绑定 IP，“0.0.0.0”表示所有 IP 绑定，可填入域名
$port = "<your-binding-port>"  # 绑定端口

# 导入证书到本地计算机的个人存储区
$addr = $ipaddr + ":" + $port
$cert = Import-PfxCertificate -FilePath "$pfxPath" -CertStoreLocation Cert:\\LocalMachine\\My -Password (ConvertTo-SecureString -String "$pfxPassword" -AsPlainText -Force) -Exportable
# 获取 Thumbprint
$thumbprint = $cert.Thumbprint
# 检测端口是否绑定证书，如绑定则删除绑定
$isExist = netsh http show sslcert ipport=$addr
if ($isExist -like "*$addr*"){ netsh http delete sslcert ipport=$addr }
# 绑定到端口
netsh http add sslcert ipport=$addr certhash=$thumbprint
# 删除目录下的证书文件
Remove-Item -Path "$pfxPath" -Force
      `.trim();

    case "ps_binding_rdp":
      return `# *** 需要管理员权限 ***

# 请将以下变量替换为实际值
$pfxPath = "\${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_PATH}" # PFX 文件路径（与表单中保持一致）
$pfxPassword = "\${CERTIMATE_DEPLOYER_CMDVAR_PFX_PASSWORD}" # PFX 密码（与表单中保持一致）

# 导入证书到本地计算机的个人存储区
$cert = Import-PfxCertificate -FilePath "$pfxPath" -CertStoreLocation Cert:\\LocalMachine\\My -Password (ConvertTo-SecureString -String "$pfxPassword" -AsPlainText -Force) -Exportable
# 获取 Thumbprint
$thumbprint = $cert.Thumbprint
# 绑定到 RDP
$rdpCertPath = "HKLM:\\SYSTEM\\CurrentControlSet\\Control\\Terminal Server\\WinStations\\RDP-Tcp"
Set-ItemProperty -Path $rdpCertPath -Name "SSLCertificateSHA1Hash" -Value "$thumbprint"
      `.trim();
  }
};

export const useSharedFormFieldsAndHandlers = (
  form: FormInstance,
  {
    format = "fileFormat",
    certPath = "filePathForCrt",
    pfxPassword = "pfxPassword",
    jksAlias = "jksAlias",
    jksKeypass = "jksKeypass",
    jksStorepass = "jksStorepass",
  }: {
    format?: string;
    certPath?: string;
    pfxPassword?: string;
    jksAlias?: string;
    jksKeypass?: string;
    jksStorepass?: string;
  }
) => {
  const { parentNamePath } = useFormNestedFieldsContext();

  const fieldFormat = Form.useWatch([parentNamePath, format], form);
  const fieldCertPath = Form.useWatch([parentNamePath, certPath], form);

  const handleChangeFormat = (value: string) => {
    if (fieldFormat === value) return;

    switch (value) {
      case FORMAT_PEM:
        {
          if (/(.pfx|.jks)$/.test(fieldCertPath)) {
            form.setFieldValue([parentNamePath, certPath], fieldCertPath.replace(/(.pfx|.jks)$/, ".crt"));
          }
        }
        break;

      case FORMAT_PFX:
        {
          if (/(.pem|.crt|.jks)$/.test(fieldCertPath)) {
            form.setFieldValue([parentNamePath, certPath], fieldCertPath.replace(/(.pem|.crt|.jks)$/, ".pfx"));
          }
        }
        break;

      case FORMAT_JKS:
        {
          if (/(.pem|.crt|.pfx)$/.test(fieldCertPath)) {
            form.setFieldValue([parentNamePath, certPath], fieldCertPath.replace(/(.pem|.crt|.pfx)$/, ".jks"));
          }
        }
        break;
    }
  };

  const handleRandomPfxPassword = () => {
    const password = randomString();
    form.setFieldValue([parentNamePath, pfxPassword], password);
  };

  const handleRandomJksAlias = () => {
    const suffixlen = 6;
    const alias = "certimate_" + Math.random().toFixed(suffixlen).slice(-suffixlen);
    form.setFieldValue([parentNamePath, jksAlias], alias);
  };

  const handleRandomJksKeypass = () => {
    const password = randomString();
    form.setFieldValue([parentNamePath, jksKeypass], password);
  };

  const handleRandomJksStorepass = () => {
    const password = randomString();
    form.setFieldValue([parentNamePath, jksStorepass], password);
  };

  return {
    fieldFormat,
    fieldCertPath,

    handleChangeFormat,
    handleRandomPfxPassword,
    handleRandomJksAlias,
    handleRandomJksKeypass,
    handleRandomJksStorepass,
  };
};

const BizDeployNodeConfigFieldsProviderLocal = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const {
    fieldFormat: fieldFileFormat,
    handleChangeFormat: handleFileFormatSelect,
    handleRandomPfxPassword: handleRandomPfxPasswordClick,
    handleRandomJksAlias: handleRandomJksAliasClick,
    handleRandomJksKeypass: handleRandomJksKeypassClick,
    handleRandomJksStorepass: handleRandomJksStorepassClick,
  } = useSharedFormFieldsAndHandlers(formInst, {});

  const handlePresetPreScriptClick = (key: string) => {
    switch (key) {
      case "sh_backup_files":
      case "ps_backup_files":
        {
          const presetScriptParams = {
            certPath: formInst.getFieldValue([parentNamePath, "filePathForCrt"]),
            keyPath: formInst.getFieldValue([parentNamePath, "filePathForKey"]),
          };
          formInst.setFieldValue([parentNamePath, "shellEnv"], SHELLENV_SH);
          formInst.setFieldValue([parentNamePath, "preCommand"], initPresetScript(key, presetScriptParams));
        }
        break;
    }
  };

  const handlePresetPostScriptClick = (key: string) => {
    switch (key) {
      case "sh_reload_nginx":
        {
          formInst.setFieldValue([parentNamePath, "shellEnv"], SHELLENV_SH);
          formInst.setFieldValue([parentNamePath, "postCommand"], initPresetScript(key));
        }
        break;

      case "ps_binding_iis":
      case "ps_binding_netsh":
      case "ps_binding_rdp":
        {
          const presetScriptParams = {
            certPath: formInst.getFieldValue([parentNamePath, "filePathForCrt"]),
            pfxPassword: formInst.getFieldValue([parentNamePath, "pfxPassword"]),
          };
          formInst.setFieldValue([parentNamePath, "shellEnv"], SHELLENV_POWERSHELL);
          formInst.setFieldValue([parentNamePath, "postCommand"], initPresetScript(key, presetScriptParams));
        }
        break;
    }
  };

  return (
    <>
      <Form.Item>
        <Tips message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local.guide") }}></span>} />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "fileFormat"]}
        initialValue={initialValues.fileFormat}
        label={t("workflow_node.deploy.form.shared_file_format.label")}
        rules={[formRule]}
      >
        <Select
          options={[FORMAT_PEM, FORMAT_PFX, FORMAT_JKS].map((s) => ({
            label: t(`workflow_node.deploy.form.shared_file_format.option.${s.toLowerCase()}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.shared_file_format.placeholder")}
          onSelect={handleFileFormatSelect}
        />
      </Form.Item>

      <Show when={fieldFileFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "filePathForKey"]}
          initialValue={initialValues.filePathForKey}
          label={t("workflow_node.deploy.form.shared_file_path_for_key.label")}
          extra={t("workflow_node.deploy.form.shared_file_path_for_key.help")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.shared_file_path_for_key.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item
        name={[parentNamePath, "filePathForCrt"]}
        initialValue={initialValues.filePathForCrt}
        label={t(`workflow_node.deploy.form.shared_file_path_for_crt.label`)}
        extra={t("workflow_node.deploy.form.shared_file_path_for_crt.help")}
        rules={[formRule]}
      >
        <Input placeholder={t(`workflow_node.deploy.form.shared_file_path_for_crt.placeholder`)} />
      </Form.Item>

      <Show when={fieldFileFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "filePathForCrtOnlyServer"]}
          initialValue={initialValues.filePathForCrtOnlyServer}
          label={t("workflow_node.deploy.form.shared_file_path_for_servercrt.label")}
          extra={t("workflow_node.deploy.form.shared_file_path_for_servercrt.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.shared_file_path_for_servercrt.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "filePathForCrtOnlyIntermedia"]}
          initialValue={initialValues.filePathForCrtOnlyIntermedia}
          label={t("workflow_node.deploy.form.shared_file_path_for_intermediacrt.label")}
          extra={t("workflow_node.deploy.form.shared_file_path_for_intermediacrt.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.shared_file_path_for_intermediacrt.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFileFormat === FORMAT_PFX}>
        <Form.Item label={t("workflow_node.deploy.form.shared_pfx_password.label")}>
          <Space.Compact className="w-full">
            <Form.Item name={[parentNamePath, "pfxPassword"]} initialValue={initialValues.pfxPassword} rules={[formRule]} noStyle>
              <Input placeholder={t("workflow_node.deploy.form.shared_pfx_password.placeholder")} />
            </Form.Item>
            <Tooltip title={t("common.text.random_roll")}>
              <Button className="px-2" onClick={handleRandomPfxPasswordClick}>
                <IconDice6 size="1.25em" />
              </Button>
            </Tooltip>
          </Space.Compact>
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "pfxEncoder"]}
          initialValue={initialValues.pfxEncoder}
          label={t("workflow_node.deploy.form.shared_pfx_encoder.label")}
          rules={[formRule]}
        >
          <Select
            options={["LegacyRC2", "LegacyDES", "Modern2023", "Modern2026"].map((s) => ({
              label: t(`workflow_node.deploy.form.shared_pfx_encoder.option.${s.toLowerCase()}.label`),
              value: s,
            }))}
            placeholder={t("workflow_node.deploy.form.shared_pfx_encoder.placeholder")}
          />
        </Form.Item>
      </Show>

      <Show when={fieldFileFormat === FORMAT_JKS}>
        <Form.Item
          label={t("workflow_node.deploy.form.shared_jks_alias.label")}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_jks_alias.tooltip") }}></span>}
        >
          <Space.Compact className="w-full">
            <Form.Item name={[parentNamePath, "jksAlias"]} initialValue={initialValues.jksAlias} rules={[formRule]} noStyle>
              <Input placeholder={t("workflow_node.deploy.form.shared_jks_alias.placeholder")} />
            </Form.Item>
            <Tooltip title={t("common.text.random_roll")}>
              <Button className="px-2" onClick={handleRandomJksAliasClick}>
                <IconDice6 size="1.25em" />
              </Button>
            </Tooltip>
          </Space.Compact>
        </Form.Item>

        <Form.Item
          label={t("workflow_node.deploy.form.shared_jks_keypass.label")}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_jks_keypass.tooltip") }}></span>}
        >
          <Space.Compact className="w-full">
            <Form.Item name={[parentNamePath, "jksKeypass"]} initialValue={initialValues.jksKeypass} rules={[formRule]} noStyle>
              <Input placeholder={t("workflow_node.deploy.form.shared_jks_keypass.placeholder")} />
            </Form.Item>
            <Tooltip title={t("common.text.random_roll")}>
              <Button className="px-2" onClick={handleRandomJksKeypassClick}>
                <IconDice6 size="1.25em" />
              </Button>
            </Tooltip>
          </Space.Compact>
        </Form.Item>

        <Form.Item
          label={t("workflow_node.deploy.form.shared_jks_storepass.label")}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_jks_storepass.tooltip") }}></span>}
        >
          <Space.Compact className="w-full">
            <Form.Item name={[parentNamePath, "jksStorepass"]} initialValue={initialValues.jksStorepass} rules={[formRule]} noStyle>
              <Input placeholder={t("workflow_node.deploy.form.shared_jks_storepass.placeholder")} />
            </Form.Item>
            <Tooltip title={t("common.text.random_roll")}>
              <Button className="px-2" onClick={handleRandomJksStorepassClick}>
                <IconDice6 size="1.25em" />
              </Button>
            </Tooltip>
          </Space.Compact>
        </Form.Item>
      </Show>

      <Form.Item
        name={[parentNamePath, "shellEnv"]}
        initialValue={initialValues.shellEnv}
        label={t("workflow_node.deploy.form.local_shell_env.label")}
        rules={[formRule]}
      >
        <Select
          options={[SHELLENV_SH, SHELLENV_CMD, SHELLENV_POWERSHELL].map((s) => ({
            label: t(`workflow_node.deploy.form.local_shell_env.option.${s.toLowerCase()}.label`),
            value: s,
          }))}
        />
      </Form.Item>

      <Form.Item label={t("workflow_node.deploy.form.local_pre_command.label")}>
        <div className="absolute -top-1.5 right-0 -translate-y-full">
          <PresetScriptTemplatesPopselect
            options={["sh_backup_files", "ps_backup_files"].map((key) => ({
              key,
              label: t(`workflow_node.deploy.form.local_preset_scripts.${key}`),
            }))}
            trigger={["click"]}
            onSelect={(key, template) => {
              if (template) {
                formInst.setFieldValue([parentNamePath, "preCommand"], template.command);
              } else {
                handlePresetPreScriptClick(key);
              }
            }}
          >
            <Button size="small" type="link">
              {t("preset.dropdown.script.button")}
              <IconChevronDown size="1.25em" />
            </Button>
          </PresetScriptTemplatesPopselect>
        </div>
        <Form.Item name={[parentNamePath, "preCommand"]} initialValue={initialValues.preCommand} noStyle rules={[formRule]}>
          <CodeTextInput
            height="auto"
            minHeight="64px"
            maxHeight="256px"
            language={["shell", "powershell"]}
            placeholder={t("workflow_node.deploy.form.local_pre_command.placeholder")}
          />
        </Form.Item>
      </Form.Item>

      <Form.Item label={t("workflow_node.deploy.form.local_post_command.label")}>
        <div className="absolute -top-1.5 right-0 -translate-y-full">
          <Space align="center" separator={<Divider orientation="vertical" />} size={0}>
            <Popover content={<div dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_script_command.vartips") }} />} mouseEnterDelay={1}>
              <Button color="default" size="small" variant="link">
                <IconBulb size="1.25em" />
              </Button>
            </Popover>
            <PresetScriptTemplatesPopselect
              options={["sh_reload_nginx", "ps_binding_iis", "ps_binding_netsh", "ps_binding_rdp"].map((key) => ({
                key,
                label: t(`workflow_node.deploy.form.local_preset_scripts.${key}`),
                onClick: () => handlePresetPostScriptClick(key),
              }))}
              trigger={["click"]}
              onSelect={(key, template) => {
                if (template) {
                  formInst.setFieldValue([parentNamePath, "postCommand"], template.command);
                } else {
                  handlePresetPostScriptClick(key);
                }
              }}
            >
              <Button size="small" type="link">
                {t("preset.dropdown.script.button")}
                <IconChevronDown size="1.25em" />
              </Button>
            </PresetScriptTemplatesPopselect>
          </Space>
        </div>
        <Form.Item name={[parentNamePath, "postCommand"]} initialValue={initialValues.postCommand} noStyle rules={[formRule]}>
          <CodeTextInput
            height="auto"
            minHeight="64px"
            maxHeight="256px"
            language={["shell", "powershell"]}
            placeholder={t("workflow_node.deploy.form.local_post_command.placeholder")}
          />
        </Form.Item>
      </Form.Item>
    </>
  );
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    fileFormat: FORMAT_PEM,
    filePathForKey: "/etc/ssl/certimate/cert.key",
    filePathForCrt: "/etc/ssl/certimate/cert.crt",
    shellEnv: SHELLENV_SH,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t: _ } = i18n;

  return z
    .object({
      fileFormat: z.enum([FORMAT_PEM, FORMAT_PFX, FORMAT_JKS]),
      filePathForKey: z.string().max(256).nullish(),
      filePathForCrt: z.string().max(256).nullish(),
      filePathForCrtOnlyServer: z.string().max(256).nullish(),
      filePathForCrtOnlyIntermedia: z.string().max(256).nullish(),
      pfxPassword: z.string().nullish(),
      pfxEncoder: z.string().nullish(),
      jksAlias: z.string().nullish(),
      jksKeypass: z.string().nullish(),
      jksStorepass: z.string().nullish(),
      shellEnv: z.enum([SHELLENV_SH, SHELLENV_CMD, SHELLENV_POWERSHELL]),
      preCommand: z.string().max(20480).nullish(),
      postCommand: z.string().max(20480).nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.fileFormat) {
        case FORMAT_PFX:
          {
            const scPfxPassword = z.string().nonempty();
            const spPfxPassword = scPfxPassword.safeParse(values.pfxPassword);
            if (!spPfxPassword.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spPfxPassword.error).errors.join(),
                path: ["pfxPassword"],
              });
            }
          }
          break;

        case FORMAT_JKS:
          {
            const scJksAlias = z.string().nonempty();
            const spJksAlias = scJksAlias.safeParse(values.jksAlias);
            if (!spJksAlias.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spJksAlias.error).errors.join(),
                path: ["jksAlias"],
              });
            }

            const scJksKeypass = z.string().nonempty();
            const spJksKeypass = scJksKeypass.safeParse(values.jksKeypass);
            if (!spJksKeypass.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spJksKeypass.error).errors.join(),
                path: ["jksKeypass"],
              });
            }

            const scJksStorepass = z.string().nonempty();
            const spJksStorepass = scJksStorepass.safeParse(values.jksStorepass);
            if (!spJksStorepass.success) {
              ctx.addIssue({
                code: "custom",
                message: z.treeifyError(spJksStorepass.error).errors.join(),
                path: ["jksStorepass"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderLocal, {
  getInitialValues,
  getSchema,
});

export default _default;
