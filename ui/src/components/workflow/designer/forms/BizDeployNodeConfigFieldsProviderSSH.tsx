import { getI18n, useTranslation } from "react-i18next";
import { IconBulb, IconChevronDown } from "@tabler/icons-react";
import { Button, Divider, Form, Input, Popover, Select, Space, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import CodeInput from "@/components/CodeInput";
import PresetScriptTemplatesPopselect from "@/components/preset/PresetScriptTemplatesPopselect";
import Show from "@/components/Show";
import { CERTIFICATE_FORMATS } from "@/domain/certificate";

import { useFormNestedFieldsContext } from "./_context";
import { initPresetScript as _initPresetScript } from "./BizDeployNodeConfigFieldsProviderLocal";

const FORMAT_PEM = CERTIFICATE_FORMATS.PEM;
const FORMAT_PFX = CERTIFICATE_FORMATS.PFX;
const FORMAT_JKS = CERTIFICATE_FORMATS.JKS;

const initPresetScript = (
  key: Parameters<typeof _initPresetScript>[0] | "sh_replace_synologydsm_ssl" | "sh_replace_fnos_ssl" | "sh_replace_qnap_ssl",
  params?: Parameters<typeof _initPresetScript>[1]
) => {
  switch (key) {
    case "sh_replace_synologydsm_ssl":
      return `# *** 需要 root 权限 ***
# 注意仅支持替换证书，需本身已开启过一次 HTTPS
# 脚本参考 https://github.com/catchdave/ssl-certs/blob/main/replace_synology_ssl_certs.sh

# 请将以下变量替换为实际值
$tmpFullchainPath = "\${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_PATH}" # 证书文件路径（与表单中保持一致）
$tmpCertPath = "\${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_SERVER_PATH}" # 服务器证书文件路径（与表单中保持一致）
$tmpKeyPath = "\${CERTIMATE_DEPLOYER_CMDVAR_PRIVATEKEY_PATH}" # 私钥文件路径（与表单中保持一致）

DEBUG=1
error_exit() { echo "[ERROR] $1"; exit 1; }
warn() { echo "[WARN] $1"; }
info() { echo "[INFO] $1"; }
debug() { [[ "\${DEBUG}" ]] && echo "[DEBUG] $1"; }

certs_src_dir="/usr/syno/etc/certificate/system/default"
target_cert_dirs=(
  "/usr/syno/etc/certificate/system/FQDN"
  "/usr/local/etc/certificate/ScsiTarget/pkg-scsi-plugin-server/"
  "/usr/local/etc/certificate/SynologyDrive/SynologyDrive/"
  "/usr/local/etc/certificate/WebDAVServer/webdav/"
  "/usr/local/etc/certificate/ActiveBackup/ActiveBackup/"
  "/usr/syno/etc/certificate/smbftpd/ftpd/")

# 获取证书目录
default_dir_name=$(</usr/syno/etc/certificate/_archive/DEFAULT)
if [[ -n "$default_dir_name" ]]; then
  target_cert_dirs+=("/usr/syno/etc/certificate/_archive/\${default_dir_name}")
  debug "Default cert directory found: '/usr/syno/etc/certificate/_archive/\${default_dir_name}'"
else
  warn "No default directory found. Probably unusual? Check: 'cat /usr/syno/etc/certificate/_archive/DEFAULT'"
fi

# 获取反向代理证书目录
for proxy in /usr/syno/etc/certificate/ReverseProxy/*/; do
  debug "Found proxy dir: \${proxy}"
  target_cert_dirs+=("\${proxy}")
done

[[ "\${DEBUG}" ]] && set -x

# 复制文件
cp -rf "$tmpFullchainPath" "\${certs_src_dir}/fullchain.pem" || error_exit "Halting because of error moving fullchain file"
cp -rf "$tmpCertPath" "\${certs_src_dir}/cert.pem" || error_exit "Halting because of error moving cert file"
cp -rf "$tmpKeyPath" "\${certs_src_dir}/privkey.pem" || error_exit "Halting because of error moving privkey file"
chown root:root "\${certs_src_dir}/"{privkey,fullchain,cert}.pem || error_exit "Halting because of error chowning files"
info "Certs moved from /tmp & chowned."

# 替换证书
for target_dir in "\${target_cert_dirs[@]}"; do
  if [[ ! -d "$target_dir" ]]; then
    debug "Target cert directory '$target_dir' not found, skipping..."
    continue
  fi
  info "Copying certificates to '$target_dir'"
  if ! (cp "\${certs_src_dir}/"{privkey,fullchain,cert}.pem "$target_dir/" && \
    chown root:root "$target_dir/"{privkey,fullchain,cert}.pem); then
      warn "Error copying or chowning certs to \${target_dir}"
  fi
done

# 重启服务
info "Rebooting all the things..."
/usr/syno/bin/synosystemctl restart nmbd
/usr/syno/bin/synosystemctl restart avahi
/usr/syno/bin/synosystemctl restart ldap-server
/usr/syno/bin/synopkg is_onoff ScsiTarget 1>/dev/null && /usr/syno/bin/synopkg restart ScsiTarget
/usr/syno/bin/synopkg is_onoff SynologyDrive 1>/dev/null && /usr/syno/bin/synopkg restart SynologyDrive
/usr/syno/bin/synopkg is_onoff WebDAVServer 1>/dev/null && /usr/syno/bin/synopkg restart WebDAVServer
/usr/syno/bin/synopkg is_onoff ActiveBackup 1>/dev/null && /usr/syno/bin/synopkg restart ActiveBackup
if ! /usr/syno/bin/synow3tool --gen-all && sudo /usr/syno/bin/synosystemctl restart nginx; then
  warn "nginx failed to restart"
fi

info "Completed"
      `.trim();

    case "sh_replace_fnos_ssl":
      return `# *** 需要 root 权限 ***
# 注意仅支持替换证书，需本身已开启过一次 HTTPS
# 脚本参考 https://github.com/lfgyx/fnos_certificate_update/blob/main/src/update_cert.sh

# 请将以下变量替换为实际值
# 飞牛证书实际存放路径请在 \`/usr/trim/etc/network_cert_all.conf\` 中查看，注意不要修改文件名
$tmpFullchainPath = "\${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_PATH}" # 证书文件路径（与表单中保持一致）
$tmpCertPath = "\${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_SERVER_PATH}" # 服务器证书文件路径（与表单中保持一致）
$tmpKeyPath = "\${CERTIMATE_DEPLOYER_CMDVAR_PRIVATEKEY_PATH}" # 私钥文件路径（与表单中保持一致）
$fnFullchainPath = "/usr/trim/var/trim_connect/ssls/example.com/1234567890/fullchain.crt" # 飞牛证书文件路径
$fnCertPath = "/usr/trim/var/trim_connect/ssls/example.com/1234567890/example.com.crt" # 飞牛服务器证书文件路径
$fnKeyPath = "/usr/trim/var/trim_connect/ssls/example.com/1234567890/example.com.key" # 飞牛私钥文件路径
$domain = "<your-domain-name>" # 域名

# 复制文件
cp -rf "$tmpFullchainPath" "$fnFullchainPath"
cp -rf "$tmpCertPath" "$fnCertPath"
cp -rf "$tmpKeyPath" "$fnKeyPath"
chmod 755 "$fnFullchainPath"
chmod 755 "$fnCertPath"
chmod 755 "$fnKeyPath"

# 更新数据库
NEW_EFFECT_DATE=$(openssl x509 -startdate -noout -in "$fnCertPath" | sed "s/^.*=\\(.*\\)$/\\1/")
NEW_EFFECT_TIMESTAMP=$(date -d "$NEW_EFFECT_DATE" +%s%3N)
NEW_EXPIRY_DATE=$(openssl x509 -enddate -noout -in "$fnCertPath" | sed "s/^.*=\\(.*\\)$/\\1/")
NEW_EXPIRY_TIMESTAMP=$(date -d "$NEW_EXPIRY_DATE" +%s%3N)
psql -U postgres -d trim_connect -c "UPDATE cert SET valid_from=$NEW_EFFECT_TIMESTAMP, valid_to=$NEW_EXPIRY_TIMESTAMP WHERE domain='$domain'"

# 重启服务
systemctl restart webdav.service
systemctl restart smbftpd.service
systemctl restart trim_nginx.service
      `.trim();

    case "sh_replace_qnap_ssl":
      return `# *** 需要 root 权限 ***
# 注意仅支持替换证书，需本身已开启过一次 HTTPS

# 请将以下变量替换为实际值
$tmpFullchainPath = "\${CERTIMATE_DEPLOYER_CMDVAR_CERTIFICATE_PATH}"}" # 证书文件路径（与表单中保持一致）
$tmpKeyPath = "\${CERTIMATE_DEPLOYER_CMDVAR_PRIVATEKEY_PATH}" # 私钥文件路径（与表单中保持一致）

# 复制文件
cp -rf "$tmpFullchainPath" /etc/stunnel/backup.cert
cp -rf "$tmpKeyPath" /etc/stunnel/backup.key
cat /etc/stunnel/backup.key > /etc/stunnel/stunnel.pem
cat /etc/stunnel/backup.cert >> /etc/stunnel/stunnel.pem
chmod 600 /etc/stunnel/backup.cert
chmod 600 /etc/stunnel/backup.key
chmod 600 /etc/stunnel/stunnel.pem

# 重启服务
/etc/init.d/stunnel.sh restart
/etc/init.d/reverse_proxy.sh reload
      `.trim();
  }

  return _initPresetScript(key as Parameters<typeof _initPresetScript>[0], params);
};

const BizDeployNodeConfigFieldsProviderSSH = () => {
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
          formInst.setFieldValue([parentNamePath, "preCommand"], initPresetScript(key, presetScriptParams));
        }
        break;
    }
  };

  const handlePresetPostScriptClick = (key: string) => {
    switch (key) {
      case "sh_reload_nginx":
        {
          formInst.setFieldValue([parentNamePath, "postCommand"], initPresetScript(key));
        }
        break;

      case "sh_replace_synologydsm_ssl":
      case "sh_replace_fnos_ssl":
      case "sh_replace_qnap_ssl":
        {
          const presetScriptParams = {
            certPath: formInst.getFieldValue([parentNamePath, "certPath"]),
            certPathForServerOnly: formInst.getFieldValue([parentNamePath, "certPathForServerOnly"]),
            certPathForIntermediaOnly: formInst.getFieldValue([parentNamePath, "certPathForIntermediaOnly"]),
            keyPath: formInst.getFieldValue([parentNamePath, "keyPath"]),
          };
          formInst.setFieldValue([parentNamePath, "postCommand"], initPresetScript(key, presetScriptParams));
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
          formInst.setFieldValue([parentNamePath, "postCommand"], initPresetScript(key, presetScriptParams));
        }
        break;
    }
  };

  return (
    <>
      <Form.Item
        name={[parentNamePath, "useSCP"]}
        initialValue={initialValues.useSCP}
        label={t("workflow_node.deploy.form.ssh_use_scp.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_use_scp.tooltip") }}></span>}
      >
        <Switch />
      </Form.Item>

      <Form.Item
        name={[parentNamePath, "format"]}
        initialValue={initialValues.format}
        label={t("workflow_node.deploy.form.ssh_format.label")}
        rules={[formRule]}
      >
        <Select
          options={[FORMAT_PEM, FORMAT_PFX, FORMAT_JKS].map((s) => ({
            key: s,
            label: t(`workflow_node.deploy.form.ssh_format.option.${s.toLowerCase()}.label`),
            value: s,
          }))}
          placeholder={t("workflow_node.deploy.form.ssh_format.placeholder")}
          onSelect={handleFormatSelect}
        />
      </Form.Item>

      <Show when={fieldFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "keyPath"]}
          initialValue={initialValues.keyPath}
          label={t("workflow_node.deploy.form.ssh_key_path.label")}
          extra={t("workflow_node.deploy.form.ssh_key_path.help")}
          rules={[formRule]}
        >
          <Input placeholder={t("workflow_node.deploy.form.ssh_key_path.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item
        name={[parentNamePath, "certPath"]}
        initialValue={initialValues.certPath}
        label={t(`workflow_node.deploy.form.ssh_${fieldFormat === FORMAT_PEM ? "fullchaincert" : "cert"}_path.label`)}
        extra={t("workflow_node.deploy.form.ssh_cert_path.help")}
        rules={[formRule]}
      >
        <Input placeholder={t(`workflow_node.deploy.form.ssh_${fieldFormat === FORMAT_PEM ? "fullchaincert" : "cert"}_path.placeholder`)} />
      </Form.Item>

      <Show when={fieldFormat === FORMAT_PEM}>
        <Form.Item
          name={[parentNamePath, "certPathForServerOnly"]}
          initialValue={initialValues.certPathForServerOnly}
          label={t("workflow_node.deploy.form.ssh_servercert_path.label")}
          extra={t("workflow_node.deploy.form.ssh_servercert_path.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.ssh_servercert_path.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "certPathForIntermediaOnly"]}
          initialValue={initialValues.certPathForIntermediaOnly}
          label={t("workflow_node.deploy.form.ssh_intermediacert_path.label")}
          extra={t("workflow_node.deploy.form.ssh_intermediacert_path.help")}
          rules={[formRule]}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.ssh_intermediacert_path.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_PFX}>
        <Form.Item
          name={[parentNamePath, "pfxPassword"]}
          initialValue={initialValues.pfxPassword}
          label={t("workflow_node.deploy.form.ssh_pfx_password.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_pfx_password.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ssh_pfx_password.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_JKS}>
        <Form.Item
          name={[parentNamePath, "jksAlias"]}
          initialValue={initialValues.jksAlias}
          label={t("workflow_node.deploy.form.ssh_jks_alias.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_jks_alias.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ssh_jks_alias.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "jksKeypass"]}
          initialValue={initialValues.jksKeypass}
          label={t("workflow_node.deploy.form.ssh_jks_keypass.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_jks_keypass.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ssh_jks_keypass.placeholder")} />
        </Form.Item>

        <Form.Item
          name={[parentNamePath, "jksStorepass"]}
          initialValue={initialValues.jksStorepass}
          label={t("workflow_node.deploy.form.ssh_jks_storepass.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_jks_storepass.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ssh_jks_storepass.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item label={t("workflow_node.deploy.form.ssh_pre_command.label")}>
        <div className="absolute -top-1.5 right-0 -translate-y-full">
          <PresetScriptTemplatesPopselect
            options={["sh_backup_files", "ps_backup_files"].map((key) => ({
              key,
              label: t(`workflow_node.deploy.form.ssh_preset_scripts.${key}`),
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
            placeholder={t("workflow_node.deploy.form.ssh_pre_command.placeholder")}
          />
        </Form.Item>
      </Form.Item>

      <Form.Item label={t("workflow_node.deploy.form.ssh_post_command.label")}>
        <div className="absolute -top-1.5 right-0 -translate-y-full">
          <Space align="center" separator={<Divider orientation="vertical" />} size={0}>
            <Popover content={<div dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.shared_script_command.vartips") }} />} mouseEnterDelay={1}>
              <Button color="default" size="small" variant="link">
                <IconBulb size="1.25em" />
              </Button>
            </Popover>
            <PresetScriptTemplatesPopselect
              options={[
                "sh_reload_nginx",
                "sh_replace_synologydsm_ssl",
                "sh_replace_fnos_ssl",
                "sh_replace_qnap_ssl",
                "ps_binding_iis",
                "ps_binding_netsh",
                "ps_binding_rdp",
              ].map((key) => ({
                key,
                label: t(`workflow_node.deploy.form.ssh_preset_scripts.${key}`),
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
            placeholder={t("workflow_node.deploy.form.ssh_post_command.placeholder")}
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
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      useSCP: z.boolean().nullish(),
      format: z.literal([FORMAT_PEM, FORMAT_PFX, FORMAT_JKS], t("workflow_node.deploy.form.ssh_format.placeholder")),
      keyPath: z
        .string()
        .max(256, t("common.errmsg.string_max", { max: 256 }))
        .nullish(),
      certPath: z
        .string()
        .min(1, t("workflow_node.deploy.form.ssh_cert_path.placeholder"))
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
                message: t("workflow_node.deploy.form.ssh_key_path.placeholder"),
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
                message: t("workflow_node.deploy.form.ssh_pfx_password.placeholder"),
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
                message: t("workflow_node.deploy.form.ssh_jks_alias.placeholder"),
                path: ["jksAlias"],
              });
            }

            if (!values.jksKeypass?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.ssh_jks_keypass.placeholder"),
                path: ["jksKeypass"],
              });
            }

            if (!values.jksStorepass?.trim()) {
              ctx.addIssue({
                code: "custom",
                message: t("workflow_node.deploy.form.ssh_jks_storepass.placeholder"),
                path: ["jksStorepass"],
              });
            }
          }
          break;
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigFieldsProviderSSH, {
  getInitialValues,
  getSchema,
});

export default _default;
