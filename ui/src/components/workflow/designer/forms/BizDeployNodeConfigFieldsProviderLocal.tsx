import { getI18n, useTranslation } from "react-i18next";
import { IconBulb, IconChevronDown } from "@tabler/icons-react";
import { Button, Divider, Form, Input, Popover, Select, Space } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import CodeInput from "@/components/CodeInput";
import PresetScriptTemplatesPopselect from "@/components/preset/PresetScriptTemplatesPopselect";
import Show from "@/components/Show";
import Tips from "@/components/Tips";
import { CERTIFICATE_FORMATS } from "@/domain/certificate";

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

const BizDeployNodeConfigFieldsProviderLocal = () => {
  const { i18n, t } = useTranslation();

  const { parentNamePath } = useFormNestedFieldsContext();
  const formSchema = z.object({
    [parentNamePath]: getSchema({ i18n }),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const formInst = Form.useFormInstance();
  const initialValues = getInitialValues();

  const fieldFormat = Form.useWatch([parentNamePath, "format"], formInst);
  const fieldCertPath = Form.useWatch([parentNamePath, "certPath"], formInst);

  const handleFormatSelect = (value: string) => {
    if (fieldFormat === value) return;

    switch (value) {
      case FORMAT_PEM:
        {
          if (/(.pfx|.jks)$/.test(fieldCertPath)) {
            formInst.setFieldValue([parentNamePath, "certPath"], fieldCertPath.replace(/(.pfx|.jks)$/, ".crt"));
          }
        }
        break;

      case FORMAT_PFX:
        {
          if (/(.crt|.jks)$/.test(fieldCertPath)) {
            formInst.setFieldValue([parentNamePath, "certPath"], fieldCertPath.replace(/(.crt|.jks)$/, ".pfx"));
          }
        }
        break;

      case FORMAT_JKS:
        {
          if (/(.crt|.pfx)$/.test(fieldCertPath)) {
            formInst.setFieldValue([parentNamePath, "certPath"], fieldCertPath.replace(/(.crt|.pfx)$/, ".jks"));
          }
        }
        break;
    }
  };

  const handlePresetPreScriptClick = (key: string) => {
    switch (key) {
      case "sh_backup_files":
      case "ps_backup_files":
        {
          const presetScriptParams = {
            certPath: formInst.getFieldValue([parentNamePath, "certPath"]),
            keyPath: formInst.getFieldValue([parentNamePath, "keyPath"]),
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
            certPath: formInst.getFieldValue([parentNamePath, "certPath"]),
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
        name={[parentNamePath, "format"]}
        initialValue={initialValues.format}
        label={t("workflow_node.deploy.form.local_format.label")}
        rules={[formRule]}
      >
        <Select
          options={[FORMAT_PEM, FORMAT_PFX, FORMAT_JKS].map((s) => ({
            key: s,
            label: t(`workflow_node.deploy.form.local_format.option.${s.toLowerCase()}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.local_format.placeholder")}
          onSelect={handleFormatSelect}
        />
      </Form.Item>

      <Show when={fieldFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "keyPath"]}
          initialValue={initialValues.keyPath}
          label={t("workflow_node.deploy.form.local_key_path.label")}
          extra={t("workflow_node.deploy.form.local_key_path.help")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.local_key_path.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item
        name={[parentNamePath, "certPath"]}
        initialValue={initialValues.certPath}
        label={t(`workflow_node.deploy.form.local_${fieldFormat === FORMAT_PEM ? "fullchaincert" : "cert"}_path.label`)}
        extra={t("workflow_node.deploy.form.local_cert_path.help")}
        rules={[formRule]}
      >
        <Input placeholder={t(`workflow_node.deploy.form.local_${fieldFormat === FORMAT_PEM ? "fullchaincert" : "cert"}_path.placeholder`)} />
      </Form.Item>

      <Show when={fieldFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "certPathForServerOnly"]}
          initialValue={initialValues.certPathForServerOnly}
          label={t("workflow_node.deploy.form.local_servercert_path.label")}
          extra={t("workflow_node.deploy.form.local_servercert_path.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.local_servercert_path.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "certPathForIntermediaOnly"]}
          initialValue={initialValues.certPathForIntermediaOnly}
          label={t("workflow_node.deploy.form.local_intermediacert_path.label")}
          extra={t("workflow_node.deploy.form.local_intermediacert_path.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.local_intermediacert_path.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_PFX}>
        <Form.Item
          name={[parentNamePath, "pfxPassword"]}
          initialValue={initialValues.pfxPassword}
          label={t("workflow_node.deploy.form.local_pfx_password.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_pfx_password.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.local_pfx_password.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_JKS}>
        <Form.Item
          name={[parentNamePath, "jksAlias"]}
          initialValue={initialValues.jksAlias}
          label={t("workflow_node.deploy.form.local_jks_alias.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_jks_alias.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.local_jks_alias.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "jksKeypass"]}
          initialValue={initialValues.jksKeypass}
          label={t("workflow_node.deploy.form.local_jks_keypass.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_jks_keypass.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.local_jks_keypass.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "jksStorepass"]}
          initialValue={initialValues.jksStorepass}
          label={t("workflow_node.deploy.form.local_jks_storepass.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_jks_storepass.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.local_jks_storepass.placeholder")} />
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
            key: s,
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
          <CodeInput
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
          <CodeInput
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
    format: FORMAT_PEM,
    certPath: "/etc/ssl/certimate/cert.crt",
    keyPath: "/etc/ssl/certimate/cert.key",
    shellEnv: SHELLENV_SH,
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      format: z.literal([FORMAT_PEM, FORMAT_PFX, FORMAT_JKS], t("workflow_node.deploy.form.local_format.placeholder")),
      keyPath: z
        .string()
        .max(256, t("common.errmsg.string_max", { max: 256 }))
        .nullish(),
      certPath: z
        .string()
        .min(1, t("workflow_node.deploy.form.local_cert_path.placeholder"))
        .max(256, t("common.errmsg.string_max", { max: 256 })),
      certPathForServerOnly: z
        .string()
        .max(256, t("common.errmsg.string_max", { max: 256 }))
        .nullish(),
      certPathForIntermediaOnly: z
        .string()
        .max(256, t("common.errmsg.string_max", { max: 256 }))
        .nullish(),
      pfxPassword: z.string().nullish(),
      jksAlias: z.string().nullish(),
      jksKeypass: z.string().nullish(),
      jksStorepass: z.string().nullish(),
      shellEnv: z.literal([SHELLENV_SH, SHELLENV_CMD, SHELLENV_POWERSHELL], t("workflow_node.deploy.form.local_shell_env.placeholder")),
      preCommand: z
        .string()
        .max(20480, t("common.errmsg.string_max", { max: 20480 }))
        .nullish(),
      postCommand: z
        .string()
        .max(20480, t("common.errmsg.string_max", { max: 20480 }))
        .nullish(),
    })
    .superRefine((values, ctx) => {
      switch (values.format) {
        case FORMAT_PEM:
          {
            if (!values.keyPath?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.local_key_path.placeholder"),
                path: ["keyPath"],
              });
            }
          }
          break;

        case FORMAT_PFX:
          {
            if (!values.pfxPassword?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.local_pfx_password.placeholder"),
                path: ["pfxPassword"],
              });
            }
          }
          break;

        case FORMAT_JKS:
          {
            if (!values.jksAlias?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.local_jks_alias.placeholder"),
                path: ["jksAlias"],
              });
            }

            if (!values.jksKeypass?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.local_jks_keypass.placeholder"),
                path: ["jksKeypass"],
              });
            }

            if (!values.jksStorepass?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.local_jks_storepass.placeholder"),
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
