import { useEffect, useMemo, useState } from "react";
import { getI18n, useTranslation } from "react-i18next";
import { type FlowNodeEntity, getNodeForm } from "@flowgram.ai/fixed-layout-editor";
import { IconPlus } from "@tabler/icons-react";
import { type AnchorProps, Button, Divider, Flex, Form, type FormInstance, Select, Switch, Typography, theme } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import AccessEditDrawer from "@/components/access/AccessEditDrawer";
import AccessSelect from "@/components/access/AccessSelect";
import DeploymentProviderPicker from "@/components/provider/DeploymentProviderPicker";
import DeploymentProviderSelect from "@/components/provider/DeploymentProviderSelect";
import Show from "@/components/Show";
import { type AccessModel } from "@/domain/access";
import { DEPLOYMENT_PROVIDERS, deploymentProvidersMap } from "@/domain/provider";
import { type WorkflowNodeConfigForBizDeploy, defaultNodeConfigForBizDeploy } from "@/domain/workflow";
import { useAntdForm, useZustandShallowSelector } from "@/hooks";
import { useAccessesStore } from "@/stores/access";

import { getAllPreviousNodes } from "../_util";
import { FormNestedFieldsContextProvider, NodeFormContextProvider } from "./_context";
import BizDeployNodeConfigFieldsProvider1PanelConsole from "./BizDeployNodeConfigFieldsProvider1PanelConsole";
import BizDeployNodeConfigFieldsProvider1PanelSite from "./BizDeployNodeConfigFieldsProvider1PanelSite";
import BizDeployNodeConfigFieldsProviderAliyunALB from "./BizDeployNodeConfigFieldsProviderAliyunALB";
import BizDeployNodeConfigFieldsProviderAliyunAPIGW from "./BizDeployNodeConfigFieldsProviderAliyunAPIGW";
import BizDeployNodeConfigFieldsProviderAliyunCAS from "./BizDeployNodeConfigFieldsProviderAliyunCAS";
import BizDeployNodeConfigFieldsProviderAliyunCASDeploy from "./BizDeployNodeConfigFieldsProviderAliyunCASDeploy";
import BizDeployNodeConfigFieldsProviderAliyunCDN from "./BizDeployNodeConfigFieldsProviderAliyunCDN";
import BizDeployNodeConfigFieldsProviderAliyunCLB from "./BizDeployNodeConfigFieldsProviderAliyunCLB";
import BizDeployNodeConfigFieldsProviderAliyunDCDN from "./BizDeployNodeConfigFieldsProviderAliyunDCDN";
import BizDeployNodeConfigFieldsProviderAliyunDDoS from "./BizDeployNodeConfigFieldsProviderAliyunDDoS";
import BizDeployNodeConfigFieldsProviderAliyunESA from "./BizDeployNodeConfigFieldsProviderAliyunESA";
import BizDeployNodeConfigFieldsProviderAliyunFC from "./BizDeployNodeConfigFieldsProviderAliyunFC";
import BizDeployNodeConfigFieldsProviderAliyunGA from "./BizDeployNodeConfigFieldsProviderAliyunGA";
import BizDeployNodeConfigFieldsProviderAliyunLive from "./BizDeployNodeConfigFieldsProviderAliyunLive";
import BizDeployNodeConfigFieldsProviderAliyunNLB from "./BizDeployNodeConfigFieldsProviderAliyunNLB";
import BizDeployNodeConfigFieldsProviderAliyunOSS from "./BizDeployNodeConfigFieldsProviderAliyunOSS";
import BizDeployNodeConfigFieldsProviderAliyunVOD from "./BizDeployNodeConfigFieldsProviderAliyunVOD";
import BizDeployNodeConfigFieldsProviderAliyunWAF from "./BizDeployNodeConfigFieldsProviderAliyunWAF";
import BizDeployNodeConfigFieldsProviderAPISIX from "./BizDeployNodeConfigFieldsProviderAPISIX";
import BizDeployNodeConfigFieldsProviderAWSACM from "./BizDeployNodeConfigFieldsProviderAWSACM";
import BizDeployNodeConfigFieldsProviderAWSCloudFront from "./BizDeployNodeConfigFieldsProviderAWSCloudFront";
import BizDeployNodeConfigFieldsProviderAWSIAM from "./BizDeployNodeConfigFieldsProviderAWSIAM";
import BizDeployNodeConfigFieldsProviderAzureKeyVault from "./BizDeployNodeConfigFieldsProviderAzureKeyVault";
import BizDeployNodeConfigFieldsProviderBaiduCloudAppBLB from "./BizDeployNodeConfigFieldsProviderBaiduCloudAppBLB";
import BizDeployNodeConfigFieldsProviderBaiduCloudBLB from "./BizDeployNodeConfigFieldsProviderBaiduCloudBLB";
import BizDeployNodeConfigFieldsProviderBaiduCloudCDN from "./BizDeployNodeConfigFieldsProviderBaiduCloudCDN";
import BizDeployNodeConfigFieldsProviderBaishanCDN from "./BizDeployNodeConfigFieldsProviderBaishanCDN";
import BizDeployNodeConfigFieldsProviderBaotaPanelConsole from "./BizDeployNodeConfigFieldsProviderBaotaPanelConsole";
import BizDeployNodeConfigFieldsProviderBaotaPanelSite from "./BizDeployNodeConfigFieldsProviderBaotaPanelSite";
import BizDeployNodeConfigFieldsProviderBaotaPanelGoConsole from "./BizDeployNodeConfigFieldsProviderBaotaPanelGoConsole";
import BizDeployNodeConfigFieldsProviderBaotaPanelGoSite from "./BizDeployNodeConfigFieldsProviderBaotaPanelGoSite";
import BizDeployNodeConfigFieldsProviderBaotaWAFSite from "./BizDeployNodeConfigFieldsProviderBaotaWAFSite";
import BizDeployNodeConfigFieldsProviderBunnyCDN from "./BizDeployNodeConfigFieldsProviderBunnyCDN";
import BizDeployNodeConfigFieldsProviderBytePlusCDN from "./BizDeployNodeConfigFieldsProviderBytePlusCDN";
import BizDeployNodeConfigFieldsProviderCdnfly from "./BizDeployNodeConfigFieldsProviderCdnfly";
import BizDeployNodeConfigFieldsProviderCTCCCloudAO from "./BizDeployNodeConfigFieldsProviderCTCCCloudAO";
import BizDeployNodeConfigFieldsProviderCTCCCloudCDN from "./BizDeployNodeConfigFieldsProviderCTCCCloudCDN";
import BizDeployNodeConfigFieldsProviderCTCCCloudELB from "./BizDeployNodeConfigFieldsProviderCTCCCloudELB";
import BizDeployNodeConfigFieldsProviderCTCCCloudICDN from "./BizDeployNodeConfigFieldsProviderCTCCCloudICDN";
import BizDeployNodeConfigFieldsProviderCTCCCloudLVDN from "./BizDeployNodeConfigFieldsProviderCTCCCloudLVDN";
import BizDeployNodeConfigFieldsProviderDogeCloudCDN from "./BizDeployNodeConfigFieldsProviderDogeCloudCDN";
import BizDeployNodeConfigFieldsProviderFlexCDN from "./BizDeployNodeConfigFieldsProviderFlexCDN";
import BizDeployNodeConfigFieldsProviderGcoreCDN from "./BizDeployNodeConfigFieldsProviderGcoreCDN";
import BizDeployNodeConfigFieldsProviderGoEdge from "./BizDeployNodeConfigFieldsProviderGoEdge";
import BizDeployNodeConfigFieldsProviderHuaweiCloudCDN from "./BizDeployNodeConfigFieldsProviderHuaweiCloudCDN";
import BizDeployNodeConfigFieldsProviderHuaweiCloudELB from "./BizDeployNodeConfigFieldsProviderHuaweiCloudELB";
import BizDeployNodeConfigFieldsProviderHuaweiCloudOBS from "./BizDeployNodeConfigFieldsProviderHuaweiCloudOBS";
import BizDeployNodeConfigFieldsProviderHuaweiCloudWAF from "./BizDeployNodeConfigFieldsProviderHuaweiCloudWAF";
import BizDeployNodeConfigFieldsProviderJDCloudALB from "./BizDeployNodeConfigFieldsProviderJDCloudALB";
import BizDeployNodeConfigFieldsProviderJDCloudCDN from "./BizDeployNodeConfigFieldsProviderJDCloudCDN";
import BizDeployNodeConfigFieldsProviderJDCloudLive from "./BizDeployNodeConfigFieldsProviderJDCloudLive";
import BizDeployNodeConfigFieldsProviderJDCloudVOD from "./BizDeployNodeConfigFieldsProviderJDCloudVOD";
import BizDeployNodeConfigFieldsProviderKong from "./BizDeployNodeConfigFieldsProviderKong";
import BizDeployNodeConfigFieldsProviderKubernetesSecret from "./BizDeployNodeConfigFieldsProviderKubernetesSecret";
import BizDeployNodeConfigFieldsProviderLeCDN from "./BizDeployNodeConfigFieldsProviderLeCDN";
import BizDeployNodeConfigFieldsProviderLocal from "./BizDeployNodeConfigFieldsProviderLocal";
import BizDeployNodeConfigFieldsProviderNetlifySite from "./BizDeployNodeConfigFieldsProviderNetlifySite";
import BizDeployNodeConfigFieldsProviderProxmoxVE from "./BizDeployNodeConfigFieldsProviderProxmoxVE";
import BizDeployNodeConfigFieldsProviderQiniuCDN from "./BizDeployNodeConfigFieldsProviderQiniuCDN";
import BizDeployNodeConfigFieldsProviderQiniuKodo from "./BizDeployNodeConfigFieldsProviderQiniuKodo";
import BizDeployNodeConfigFieldsProviderQiniuPili from "./BizDeployNodeConfigFieldsProviderQiniuPili";
import BizDeployNodeConfigFieldsProviderRainYunRCDN from "./BizDeployNodeConfigFieldsProviderRainYunRCDN";
import BizDeployNodeConfigFieldsProviderRatPanelSite from "./BizDeployNodeConfigFieldsProviderRatPanelSite";
import BizDeployNodeConfigFieldsProviderSafeLine from "./BizDeployNodeConfigFieldsProviderSafeLine";
import BizDeployNodeConfigFieldsProviderSSH from "./BizDeployNodeConfigFieldsProviderSSH";
import BizDeployNodeConfigFieldsProviderTencentCloudCDN from "./BizDeployNodeConfigFieldsProviderTencentCloudCDN";
import BizDeployNodeConfigFieldsProviderTencentCloudCLB from "./BizDeployNodeConfigFieldsProviderTencentCloudCLB";
import BizDeployNodeConfigFieldsProviderTencentCloudCOS from "./BizDeployNodeConfigFieldsProviderTencentCloudCOS";
import BizDeployNodeConfigFieldsProviderTencentCloudCSS from "./BizDeployNodeConfigFieldsProviderTencentCloudCSS";
import BizDeployNodeConfigFieldsProviderTencentCloudECDN from "./BizDeployNodeConfigFieldsProviderTencentCloudECDN";
import BizDeployNodeConfigFieldsProviderTencentCloudEO from "./BizDeployNodeConfigFieldsProviderTencentCloudEO";
import BizDeployNodeConfigFieldsProviderTencentCloudGAAP from "./BizDeployNodeConfigFieldsProviderTencentCloudGAAP";
import BizDeployNodeConfigFieldsProviderTencentCloudSCF from "./BizDeployNodeConfigFieldsProviderTencentCloudSCF";
import BizDeployNodeConfigFieldsProviderTencentCloudSSL from "./BizDeployNodeConfigFieldsProviderTencentCloudSSL";
import BizDeployNodeConfigFieldsProviderTencentCloudSSLDeploy from "./BizDeployNodeConfigFieldsProviderTencentCloudSSLDeploy";
import BizDeployNodeConfigFieldsProviderTencentCloudSSLUpdate from "./BizDeployNodeConfigFieldsProviderTencentCloudSSLUpdate";
import BizDeployNodeConfigFieldsProviderTencentCloudVOD from "./BizDeployNodeConfigFieldsProviderTencentCloudVOD";
import BizDeployNodeConfigFieldsProviderTencentCloudWAF from "./BizDeployNodeConfigFieldsProviderTencentCloudWAF";
import BizDeployNodeConfigFieldsProviderUCloudUCDN from "./BizDeployNodeConfigFieldsProviderUCloudUCDN";
import BizDeployNodeConfigFieldsProviderUCloudUS3 from "./BizDeployNodeConfigFieldsProviderUCloudUS3";
import BizDeployNodeConfigFieldsProviderUniCloudWebHost from "./BizDeployNodeConfigFieldsProviderUniCloudWebHost";
import BizDeployNodeConfigFieldsProviderUpyunCDN from "./BizDeployNodeConfigFieldsProviderUpyunCDN";
import BizDeployNodeConfigFieldsProviderUpyunFile from "./BizDeployNodeConfigFieldsProviderUpyunFile";
import BizDeployNodeConfigFieldsProviderVolcEngineALB from "./BizDeployNodeConfigFieldsProviderVolcEngineALB";
import BizDeployNodeConfigFieldsProviderVolcEngineCDN from "./BizDeployNodeConfigFieldsProviderVolcEngineCDN";
import BizDeployNodeConfigFieldsProviderVolcEngineCertCenter from "./BizDeployNodeConfigFieldsProviderVolcEngineCertCenter";
import BizDeployNodeConfigFieldsProviderVolcEngineCLB from "./BizDeployNodeConfigFieldsProviderVolcEngineCLB";
import BizDeployNodeConfigFieldsProviderVolcEngineDCDN from "./BizDeployNodeConfigFieldsProviderVolcEngineDCDN";
import BizDeployNodeConfigFieldsProviderVolcEngineImageX from "./BizDeployNodeConfigFieldsProviderVolcEngineImageX";
import BizDeployNodeConfigFieldsProviderVolcEngineLive from "./BizDeployNodeConfigFieldsProviderVolcEngineLive";
import BizDeployNodeConfigFieldsProviderVolcEngineTOS from "./BizDeployNodeConfigFieldsProviderVolcEngineTOS";
import BizDeployNodeConfigFieldsProviderWangsuCDN from "./BizDeployNodeConfigFieldsProviderWangsuCDN";
import BizDeployNodeConfigFieldsProviderWangsuCDNPro from "./BizDeployNodeConfigFieldsProviderWangsuCDNPro";
import BizDeployNodeConfigFieldsProviderWangsuCertificate from "./BizDeployNodeConfigFieldsProviderWangsuCertificate";
import BizDeployNodeConfigFieldsProviderWebhook from "./BizDeployNodeConfigFieldsProviderWebhook";
import { NodeType } from "../nodes/typings";

export interface BizDeployNodeConfigFormProps {
  form: FormInstance;
  node: FlowNodeEntity;
}

const BizDeployNodeConfigForm = ({ node, ...props }: BizDeployNodeConfigFormProps) => {
  if (node.flowNodeType !== NodeType.BizDeploy) {
    console.warn(`[certimate] current workflow node type is not: ${NodeType.BizDeploy}`);
  }

  const { i18n, t } = useTranslation();

  const { token: themeToken } = theme.useToken();

  const { accesses } = useAccessesStore(useZustandShallowSelector("accesses"));
  const accessOptionFilter = (_: string, option: AccessModel) => {
    if (option.reserve) return false;
    return deploymentProvidersMap.get(fieldProvider)?.provider === option.provider;
  };

  const initialValues = useMemo(() => {
    return getNodeForm(node)?.getValueIn("config") as WorkflowNodeConfigForBizDeploy | undefined;
  }, [node]);

  const formSchema = getSchema({ i18n }).superRefine((values, ctx) => {
    if (values.certificateOutputNodeId) {
      if (!certificateOutputNodeIdOptions.some((option) => option.value === values.certificateOutputNodeId)) {
        ctx.addIssue({
          code: "custom",
          message: t("workflow_node.deploy.form.certificate_output_node_id.placeholder"),
          path: ["certificateOutputNodeId"],
        });
      }
    }
  });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    form: props.form,
    name: "workflowNodeBizDeployConfigForm",
    initialValues: initialValues ?? getInitialValues(),
  });

  const fieldProvider = Form.useWatch<string>("provider", { form: formInst, preserve: true });
  const fieldProviderAccessId = Form.useWatch<string>("providerAccessId", { form: formInst, preserve: true });

  const certificateOutputNodeIdOptions = useMemo(() => {
    return getAllPreviousNodes(node)
      .filter((node) => node.flowNodeType === NodeType.BizApply || node.flowNodeType === NodeType.BizUpload)
      .map((node) => {
        return {
          label: getNodeForm(node)?.getValueIn("name"),
          value: node.id,
        };
      });
  }, [node]);

  const [showProviderAccess, setShowProviderAccess] = useState(false);
  useEffect(() => {
    // 内置的部署提供商（如本地部署）无需显示授权信息字段
    if (fieldProvider) {
      const provider = deploymentProvidersMap.get(fieldProvider);
      setShowProviderAccess(!provider?.builtin);
    } else {
      setShowProviderAccess(false);
    }
  }, [fieldProvider]);

  const NestedProviderConfigFields = useMemo(() => {
    /*
      注意：如果追加新的子组件，请保持以 ASCII 排序。
      NOTICE: If you add new child component, please keep ASCII order.
     */
    switch (fieldProvider) {
      case DEPLOYMENT_PROVIDERS["1PANEL_CONSOLE"]: {
        return BizDeployNodeConfigFieldsProvider1PanelConsole;
      }
      case DEPLOYMENT_PROVIDERS["1PANEL_SITE"]: {
        return BizDeployNodeConfigFieldsProvider1PanelSite;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_ALB: {
        return BizDeployNodeConfigFieldsProviderAliyunALB;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_APIGW: {
        return BizDeployNodeConfigFieldsProviderAliyunAPIGW;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_CAS: {
        return BizDeployNodeConfigFieldsProviderAliyunCAS;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_CAS_DEPLOY: {
        return BizDeployNodeConfigFieldsProviderAliyunCASDeploy;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_CLB: {
        return BizDeployNodeConfigFieldsProviderAliyunCLB;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_CDN: {
        return BizDeployNodeConfigFieldsProviderAliyunCDN;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_DCDN: {
        return BizDeployNodeConfigFieldsProviderAliyunDCDN;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_DDOS: {
        return BizDeployNodeConfigFieldsProviderAliyunDDoS;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_ESA: {
        return BizDeployNodeConfigFieldsProviderAliyunESA;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_FC: {
        return BizDeployNodeConfigFieldsProviderAliyunFC;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_GA: {
        return BizDeployNodeConfigFieldsProviderAliyunGA;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_LIVE: {
        return BizDeployNodeConfigFieldsProviderAliyunLive;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_NLB: {
        return BizDeployNodeConfigFieldsProviderAliyunNLB;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_OSS: {
        return BizDeployNodeConfigFieldsProviderAliyunOSS;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_VOD: {
        return BizDeployNodeConfigFieldsProviderAliyunVOD;
      }
      case DEPLOYMENT_PROVIDERS.ALIYUN_WAF: {
        return BizDeployNodeConfigFieldsProviderAliyunWAF;
      }
      case DEPLOYMENT_PROVIDERS.APISIX: {
        return BizDeployNodeConfigFieldsProviderAPISIX;
      }
      case DEPLOYMENT_PROVIDERS.AWS_ACM: {
        return BizDeployNodeConfigFieldsProviderAWSACM;
      }
      case DEPLOYMENT_PROVIDERS.AWS_CLOUDFRONT: {
        return BizDeployNodeConfigFieldsProviderAWSCloudFront;
      }
      case DEPLOYMENT_PROVIDERS.AWS_IAM: {
        return BizDeployNodeConfigFieldsProviderAWSIAM;
      }
      case DEPLOYMENT_PROVIDERS.AZURE_KEYVAULT: {
        return BizDeployNodeConfigFieldsProviderAzureKeyVault;
      }
      case DEPLOYMENT_PROVIDERS.BAIDUCLOUD_APPBLB: {
        return BizDeployNodeConfigFieldsProviderBaiduCloudAppBLB;
      }
      case DEPLOYMENT_PROVIDERS.BAIDUCLOUD_BLB: {
        return BizDeployNodeConfigFieldsProviderBaiduCloudBLB;
      }
      case DEPLOYMENT_PROVIDERS.BAIDUCLOUD_CDN: {
        return BizDeployNodeConfigFieldsProviderBaiduCloudCDN;
      }
      case DEPLOYMENT_PROVIDERS.BAISHAN_CDN: {
        return BizDeployNodeConfigFieldsProviderBaishanCDN;
      }
      case DEPLOYMENT_PROVIDERS.BAOTAPANEL_CONSOLE: {
        return BizDeployNodeConfigFieldsProviderBaotaPanelConsole;
      }
      case DEPLOYMENT_PROVIDERS.BAOTAPANEL_SITE: {
        return BizDeployNodeConfigFieldsProviderBaotaPanelSite;
      }
      case DEPLOYMENT_PROVIDERS.BAOTAPANELGO_CONSOLE: {
        return BizDeployNodeConfigFieldsProviderBaotaPanelGoConsole;
      }
      case DEPLOYMENT_PROVIDERS.BAOTAPANELGO_SITE: {
        return BizDeployNodeConfigFieldsProviderBaotaPanelGoSite;
      }
      case DEPLOYMENT_PROVIDERS.BAOTAWAF_SITE: {
        return BizDeployNodeConfigFieldsProviderBaotaWAFSite;
      }
      case DEPLOYMENT_PROVIDERS.BUNNY_CDN: {
        return BizDeployNodeConfigFieldsProviderBunnyCDN;
      }
      case DEPLOYMENT_PROVIDERS.BYTEPLUS_CDN: {
        return BizDeployNodeConfigFieldsProviderBytePlusCDN;
      }
      case DEPLOYMENT_PROVIDERS.CDNFLY: {
        return BizDeployNodeConfigFieldsProviderCdnfly;
      }
      case DEPLOYMENT_PROVIDERS.CTCCCLOUD_AO: {
        return BizDeployNodeConfigFieldsProviderCTCCCloudAO;
      }
      case DEPLOYMENT_PROVIDERS.CTCCCLOUD_CDN: {
        return BizDeployNodeConfigFieldsProviderCTCCCloudCDN;
      }
      case DEPLOYMENT_PROVIDERS.CTCCCLOUD_ELB: {
        return BizDeployNodeConfigFieldsProviderCTCCCloudELB;
      }
      case DEPLOYMENT_PROVIDERS.CTCCCLOUD_ICDN: {
        return BizDeployNodeConfigFieldsProviderCTCCCloudICDN;
      }
      case DEPLOYMENT_PROVIDERS.CTCCCLOUD_LVDN: {
        return BizDeployNodeConfigFieldsProviderCTCCCloudLVDN;
      }
      case DEPLOYMENT_PROVIDERS.DOGECLOUD_CDN: {
        return BizDeployNodeConfigFieldsProviderDogeCloudCDN;
      }
      case DEPLOYMENT_PROVIDERS.FLEXCDN: {
        return BizDeployNodeConfigFieldsProviderFlexCDN;
      }
      case DEPLOYMENT_PROVIDERS.GCORE_CDN: {
        return BizDeployNodeConfigFieldsProviderGcoreCDN;
      }
      case DEPLOYMENT_PROVIDERS.GOEDGE: {
        return BizDeployNodeConfigFieldsProviderGoEdge;
      }
      case DEPLOYMENT_PROVIDERS.HUAWEICLOUD_CDN: {
        return BizDeployNodeConfigFieldsProviderHuaweiCloudCDN;
      }
      case DEPLOYMENT_PROVIDERS.HUAWEICLOUD_ELB: {
        return BizDeployNodeConfigFieldsProviderHuaweiCloudELB;
      }
      case DEPLOYMENT_PROVIDERS.HUAWEICLOUD_OBS: {
        return BizDeployNodeConfigFieldsProviderHuaweiCloudOBS;
      }
      case DEPLOYMENT_PROVIDERS.HUAWEICLOUD_WAF: {
        return BizDeployNodeConfigFieldsProviderHuaweiCloudWAF;
      }
      case DEPLOYMENT_PROVIDERS.JDCLOUD_ALB: {
        return BizDeployNodeConfigFieldsProviderJDCloudALB;
      }
      case DEPLOYMENT_PROVIDERS.JDCLOUD_CDN: {
        return BizDeployNodeConfigFieldsProviderJDCloudCDN;
      }
      case DEPLOYMENT_PROVIDERS.JDCLOUD_LIVE: {
        return BizDeployNodeConfigFieldsProviderJDCloudLive;
      }
      case DEPLOYMENT_PROVIDERS.JDCLOUD_VOD: {
        return BizDeployNodeConfigFieldsProviderJDCloudVOD;
      }
      case DEPLOYMENT_PROVIDERS.KONG: {
        return BizDeployNodeConfigFieldsProviderKong;
      }
      case DEPLOYMENT_PROVIDERS.KUBERNETES_SECRET: {
        return BizDeployNodeConfigFieldsProviderKubernetesSecret;
      }
      case DEPLOYMENT_PROVIDERS.LECDN: {
        return BizDeployNodeConfigFieldsProviderLeCDN;
      }
      case DEPLOYMENT_PROVIDERS.LOCAL: {
        return BizDeployNodeConfigFieldsProviderLocal;
      }
      case DEPLOYMENT_PROVIDERS.NETLIFY_SITE: {
        return BizDeployNodeConfigFieldsProviderNetlifySite;
      }
      case DEPLOYMENT_PROVIDERS.PROXMOXVE: {
        return BizDeployNodeConfigFieldsProviderProxmoxVE;
      }
      case DEPLOYMENT_PROVIDERS.QINIU_CDN: {
        return BizDeployNodeConfigFieldsProviderQiniuCDN;
      }
      case DEPLOYMENT_PROVIDERS.QINIU_KODO: {
        return BizDeployNodeConfigFieldsProviderQiniuKodo;
      }
      case DEPLOYMENT_PROVIDERS.QINIU_PILI: {
        return BizDeployNodeConfigFieldsProviderQiniuPili;
      }
      case DEPLOYMENT_PROVIDERS.RAINYUN_RCDN: {
        return BizDeployNodeConfigFieldsProviderRainYunRCDN;
      }
      case DEPLOYMENT_PROVIDERS.RATPANEL_SITE: {
        return BizDeployNodeConfigFieldsProviderRatPanelSite;
      }
      case DEPLOYMENT_PROVIDERS.SAFELINE: {
        return BizDeployNodeConfigFieldsProviderSafeLine;
      }
      case DEPLOYMENT_PROVIDERS.SSH: {
        return BizDeployNodeConfigFieldsProviderSSH;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CDN: {
        return BizDeployNodeConfigFieldsProviderTencentCloudCDN;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CLB: {
        return BizDeployNodeConfigFieldsProviderTencentCloudCLB;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_COS: {
        return BizDeployNodeConfigFieldsProviderTencentCloudCOS;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_CSS: {
        return BizDeployNodeConfigFieldsProviderTencentCloudCSS;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_ECDN: {
        return BizDeployNodeConfigFieldsProviderTencentCloudECDN;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_EO: {
        return BizDeployNodeConfigFieldsProviderTencentCloudEO;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_GAAP: {
        return BizDeployNodeConfigFieldsProviderTencentCloudGAAP;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SCF: {
        return BizDeployNodeConfigFieldsProviderTencentCloudSCF;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL: {
        return BizDeployNodeConfigFieldsProviderTencentCloudSSL;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL_DEPLOY: {
        return BizDeployNodeConfigFieldsProviderTencentCloudSSLDeploy;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_SSL_UPDATE: {
        return BizDeployNodeConfigFieldsProviderTencentCloudSSLUpdate;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_VOD: {
        return BizDeployNodeConfigFieldsProviderTencentCloudVOD;
      }
      case DEPLOYMENT_PROVIDERS.TENCENTCLOUD_WAF: {
        return BizDeployNodeConfigFieldsProviderTencentCloudWAF;
      }
      case DEPLOYMENT_PROVIDERS.UCLOUD_UCDN: {
        return BizDeployNodeConfigFieldsProviderUCloudUCDN;
      }
      case DEPLOYMENT_PROVIDERS.UCLOUD_US3: {
        return BizDeployNodeConfigFieldsProviderUCloudUS3;
      }
      case DEPLOYMENT_PROVIDERS.UNICLOUD_WEBHOST: {
        return BizDeployNodeConfigFieldsProviderUniCloudWebHost;
      }
      case DEPLOYMENT_PROVIDERS.UPYUN_CDN: {
        return BizDeployNodeConfigFieldsProviderUpyunCDN;
      }
      case DEPLOYMENT_PROVIDERS.UPYUN_FILE: {
        return BizDeployNodeConfigFieldsProviderUpyunFile;
      }
      case DEPLOYMENT_PROVIDERS.VOLCENGINE_ALB: {
        return BizDeployNodeConfigFieldsProviderVolcEngineALB;
      }
      case DEPLOYMENT_PROVIDERS.VOLCENGINE_CDN: {
        return BizDeployNodeConfigFieldsProviderVolcEngineCDN;
      }
      case DEPLOYMENT_PROVIDERS.VOLCENGINE_CERTCENTER: {
        return BizDeployNodeConfigFieldsProviderVolcEngineCertCenter;
      }
      case DEPLOYMENT_PROVIDERS.VOLCENGINE_CLB: {
        return BizDeployNodeConfigFieldsProviderVolcEngineCLB;
      }
      case DEPLOYMENT_PROVIDERS.VOLCENGINE_DCDN: {
        return BizDeployNodeConfigFieldsProviderVolcEngineDCDN;
      }
      case DEPLOYMENT_PROVIDERS.VOLCENGINE_IMAGEX: {
        return BizDeployNodeConfigFieldsProviderVolcEngineImageX;
      }
      case DEPLOYMENT_PROVIDERS.VOLCENGINE_LIVE: {
        return BizDeployNodeConfigFieldsProviderVolcEngineLive;
      }
      case DEPLOYMENT_PROVIDERS.VOLCENGINE_TOS: {
        return BizDeployNodeConfigFieldsProviderVolcEngineTOS;
      }
      case DEPLOYMENT_PROVIDERS.WANGSU_CDN: {
        return BizDeployNodeConfigFieldsProviderWangsuCDN;
      }
      case DEPLOYMENT_PROVIDERS.WANGSU_CDNPRO: {
        return BizDeployNodeConfigFieldsProviderWangsuCDNPro;
      }
      case DEPLOYMENT_PROVIDERS.WANGSU_CERTIFICATE: {
        return BizDeployNodeConfigFieldsProviderWangsuCertificate;
      }
      case DEPLOYMENT_PROVIDERS.WEBHOOK: {
        return BizDeployNodeConfigFieldsProviderWebhook;
      }
    }
  }, [fieldProvider]);

  useEffect(() => {
    // 如果未选择部署目标，则清空授权信息
    if (!fieldProvider && fieldProviderAccessId) {
      formInst.setFieldValue("providerAccessId", void 0);
      return;
    }

    // 如果已选择部署目标只有一个授权信息，则自动选择该授权信息
    if (fieldProvider && !fieldProviderAccessId) {
      const availableAccesses = accesses
        .filter((access) => accessOptionFilter(access.provider, access))
        .filter((access) => deploymentProvidersMap.get(fieldProvider)?.provider === access.provider);
      if (availableAccesses.length === 1) {
        formInst.setFieldValue("providerAccessId", availableAccesses[0].id);
      }
    }
  }, [fieldProvider, fieldProviderAccessId]);

  const handleProviderPick = (value: string) => {
    formInst.setFieldValue("provider", value);
    formInst.setFieldValue("providerAccessId", void 0);
    formInst.setFieldValue("providerConfig", void 0);
  };

  const handleProviderSelect = (value?: string | undefined) => {
    // 切换部署目标时重置表单，避免其他部署目标的配置字段影响当前部署目标
    if (initialValues?.provider === value) {
      formInst.setFieldValue("providerAccessId", void 0);
      formInst.resetFields(["providerConfig"]);
    } else {
      formInst.setFieldValue("providerAccessId", void 0);
      formInst.setFieldValue("providerConfig", void 0);
    }
  };

  return (
    <NodeFormContextProvider value={{ node }}>
      <Form {...formProps} clearOnDestroy={true} form={formInst} layout="vertical" preserve={false} scrollToFirstError>
        <Show when={!fieldProvider}>
          <DeploymentProviderPicker
            autoFocus
            placeholder={t("workflow_node.deploy.form.provider.search.placeholder")}
            showAvailability
            showSearch
            onSelect={handleProviderPick}
          />
        </Show>

        <div style={{ display: fieldProvider ? "block" : "none" }}>
          <div id="parameters" data-anchor="parameters">
            <Form.Item
              name="certificateOutputNodeId"
              label={t("workflow_node.deploy.form.certificate_output_node_id.label")}
              extra={t("workflow_node.deploy.form.certificate_output_node_id.help")}
              rules={[formRule]}
            >
              <Select
                optionRender={({ label, value }) => {
                  return (
                    <div className="flex items-center justify-between gap-4 overflow-hidden">
                      <div className="flex-1 truncate">{label}</div>
                      <div className="origin-right scale-90 font-mono text-xs" style={{ color: themeToken.colorTextSecondary }}>
                        (NodeID: {value})
                      </div>
                    </div>
                  );
                }}
                options={certificateOutputNodeIdOptions}
                placeholder={t("workflow_node.deploy.form.certificate_output_node_id.placeholder")}
              />
            </Form.Item>
          </div>

          <div id="deployment" data-anchor="deployment">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.deploy.form_anchor.deployment.title")}
              </Typography.Text>
            </Divider>

            <Form.Item name="provider" label={t("workflow_node.deploy.form.provider.label")} rules={[formRule]}>
              <DeploymentProviderSelect
                allowClear
                disabled={!!initialValues?.provider}
                placeholder={t("workflow_node.deploy.form.provider.placeholder")}
                showAvailability
                showSearch
                onSelect={handleProviderSelect}
                onClear={handleProviderSelect}
              />
            </Form.Item>

            <Form.Item className="relative" hidden={!showProviderAccess} label={t("workflow_node.deploy.form.provider_access.label")}>
              <div className="absolute -top-[6px] right-0 -translate-y-full">
                <AccessEditDrawer
                  data={{ provider: deploymentProvidersMap.get(fieldProvider!)?.provider }}
                  mode="create"
                  trigger={
                    <Button size="small" type="link">
                      {t("workflow_node.deploy.form.provider_access.button")}
                      <IconPlus size="1.25em" />
                    </Button>
                  }
                  usage="hosting"
                  afterSubmit={(record) => {
                    if (!accessOptionFilter(record.provider, record)) return;
                    if (deploymentProvidersMap.get(fieldProvider!)?.provider !== record.provider) return;
                    formInst.setFieldValue("providerAccessId", record.id);
                  }}
                />
              </div>
              <Form.Item name="providerAccessId" dependencies={["provider"]} rules={[formRule]} noStyle>
                <AccessSelect
                  disabled={!fieldProvider}
                  placeholder={t("workflow_node.deploy.form.provider_access.placeholder")}
                  showSearch
                  onFilter={accessOptionFilter}
                />
              </Form.Item>
            </Form.Item>

            <FormNestedFieldsContextProvider value={{ parentNamePath: "providerConfig" }}>
              {NestedProviderConfigFields && <NestedProviderConfigFields />}
            </FormNestedFieldsContextProvider>
          </div>

          <div id="strategy" data-anchor="strategy">
            <Divider size="small">
              <Typography.Text className="text-xs font-normal" type="secondary">
                {t("workflow_node.deploy.form_anchor.strategy.title")}
              </Typography.Text>
            </Divider>

            <Form.Item label={t("workflow_node.deploy.form.skip_on_last_succeeded.label")}>
              <Flex align="center" gap={8} wrap="wrap">
                <div>{t("workflow_node.deploy.form.skip_on_last_succeeded.prefix")}</div>
                <Form.Item name="skipOnLastSucceeded" noStyle rules={[formRule]}>
                  <Switch
                    checkedChildren={t("workflow_node.deploy.form.skip_on_last_succeeded.switch.on")}
                    unCheckedChildren={t("workflow_node.deploy.form.skip_on_last_succeeded.switch.off")}
                  />
                </Form.Item>
                <div>{t("workflow_node.deploy.form.skip_on_last_succeeded.suffix")}</div>
              </Flex>
            </Form.Item>
          </div>
        </div>
      </Form>
    </NodeFormContextProvider>
  );
};

const getAnchorItems = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }): Required<AnchorProps>["items"] => {
  const { t } = i18n;

  return ["parameters", "deployment", "strategy"].map((key) => ({
    key: key,
    title: t(`workflow_node.deploy.form_anchor.${key}.tab`),
    href: "#" + key,
  }));
};

const getInitialValues = (): Nullish<z.infer<ReturnType<typeof getSchema>>> => {
  return {
    ...defaultNodeConfigForBizDeploy(),
  };
};

const getSchema = ({ i18n = getI18n() }: { i18n?: ReturnType<typeof getI18n> }) => {
  const { t } = i18n;

  return z
    .object({
      certificateOutputNodeId: z
        .string(t("workflow_node.deploy.form.certificate_output_node_id.placeholder"))
        .nonempty(t("workflow_node.deploy.form.certificate_output_node_id.placeholder")),
      provider: z.string(t("workflow_node.deploy.form.provider.placeholder")).nonempty(t("workflow_node.deploy.form.provider.placeholder")),
      providerAccessId: z.string().nullish(),
      providerConfig: z.any().nullish(),
      skipOnLastSucceeded: z.boolean().nullish(),
    })
    .superRefine((values, ctx) => {
      if (values.provider) {
        const provider = deploymentProvidersMap.get(values.provider);
        if (!provider?.builtin && !values.providerAccessId) {
          ctx.addIssue({
            code: "custom",
            message: t("workflow_node.deploy.form.provider_access.placeholder"),
            path: ["providerAccessId"],
          });
        }
      }
    });
};

const _default = Object.assign(BizDeployNodeConfigForm, {
  getAnchorItems,
  getSchema,
});

export default _default;
