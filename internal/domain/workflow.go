package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/certimate-go/certimate/internal/domain/expr"
	xmaps "github.com/certimate-go/certimate/pkg/utils/maps"
)

const CollectionNameWorkflow = "workflow"

type Workflow struct {
	Meta
	Name          string                `json:"name" db:"name"`
	Description   string                `json:"description" db:"description"`
	Trigger       WorkflowTriggerType   `json:"trigger" db:"trigger"`
	TriggerCron   string                `json:"triggerCron" db:"triggerCron"`
	Enabled       bool                  `json:"enabled" db:"enabled"`
	GraphDraft    *WorkflowGraph        `json:"graphDraft" db:"graphDraft"`
	GraphContent  *WorkflowGraph        `json:"graphContent" db:"graphContent"`
	HasDraft      bool                  `json:"hasDraft" db:"hasDraft"`
	HasContent    bool                  `json:"hasContent" db:"hasContent"`
	LastRunId     string                `json:"lastRunId" db:"lastRunRef"`
	LastRunStatus WorkflowRunStatusType `json:"lastRunStatus" db:"lastRunStatus"`
	LastRunTime   time.Time             `json:"lastRunTime" db:"lastRunTime"`
}

type WorkflowGraph struct {
	Nodes []*WorkflowNode `json:"nodes"`
}

func (g *WorkflowGraph) GetNodeById(nodeId string) (*WorkflowNode, bool) {
	return g.getNodeInBlocksById(g.Nodes, nodeId)
}

func (g *WorkflowGraph) getNodeInBlocksById(blocks []*WorkflowNode, nodeId string) (*WorkflowNode, bool) {
	for _, node := range blocks {
		if node.Id == nodeId {
			return node, true
		}

		if len(node.Blocks) > 0 {
			if found, ok := g.getNodeInBlocksById(node.Blocks, nodeId); ok {
				return found, true
			}
		}
	}

	return nil, false
}

func (g *WorkflowGraph) Verify() error {
	if len(g.Nodes) < 2 {
		return fmt.Errorf("invalid nodes length of graph")
	} else if g.Nodes[0].Type != WorkflowNodeTypeStart {
		return fmt.Errorf("the first node is not a start node")
	} else if g.Nodes[len(g.Nodes)-1].Type != WorkflowNodeTypeEnd {
		return fmt.Errorf("the last node is not an end node")
	}

	return nil
}

func (g *WorkflowGraph) Clone() *WorkflowGraph {
	return &WorkflowGraph{
		Nodes: g.Nodes,
	}
}

type WorkflowTriggerType string

const (
	WorkflowTriggerTypeScheduled = WorkflowTriggerType("scheduled")
	WorkflowTriggerTypeManual    = WorkflowTriggerType("manual")
)

type WorkflowNode struct {
	Id     string           `json:"id"` // 节点 ID 只在该工作流中唯一，在全局中不保证唯一性
	Type   WorkflowNodeType `json:"type"`
	Data   WorkflowNodeData `json:"data"`
	Blocks []*WorkflowNode  `json:"blocks,omitempty"`
}

type WorkflowNodeType string

const (
	WorkflowNodeTypeStart       = WorkflowNodeType("start")
	WorkflowNodeTypeEnd         = WorkflowNodeType("end")
	WorkflowNodeTypeCondition   = WorkflowNodeType("condition")
	WorkflowNodeTypeBranchBlock = WorkflowNodeType("branchBlock")
	WorkflowNodeTypeTryCatch    = WorkflowNodeType("tryCatch")
	WorkflowNodeTypeTryBlock    = WorkflowNodeType("tryBlock")
	WorkflowNodeTypeCatchBlock  = WorkflowNodeType("catchBlock")
	WorkflowNodeTypeDelay       = WorkflowNodeType("delay")
	WorkflowNodeTypeBizApply    = WorkflowNodeType("bizApply")
	WorkflowNodeTypeBizUpload   = WorkflowNodeType("bizUpload")
	WorkflowNodeTypeBizMonitor  = WorkflowNodeType("bizMonitor")
	WorkflowNodeTypeBizDeploy   = WorkflowNodeType("bizDeploy")
	WorkflowNodeTypeBizNotify   = WorkflowNodeType("bizNotify")
)

type WorkflowNodeData struct {
	Name     string             `json:"name"`
	Disabled bool               `json:"disabled,omitempty,omitzero"`
	Config   WorkflowNodeConfig `json:"config,omitempty,omitzero"`
}

type WorkflowNodeConfig map[string]any

func (c WorkflowNodeConfig) AsDelay() WorkflowNodeConfigForDelay {
	return WorkflowNodeConfigForDelay{
		Wait: xmaps.GetInt(c, "wait"),
	}
}

func (c WorkflowNodeConfig) AsBranchBlock() WorkflowNodeConfigForBranchBlock {
	expression := c["expression"]
	if expression == nil {
		return WorkflowNodeConfigForBranchBlock{}
	}

	exprRaw, _ := json.Marshal(expression)
	expr, err := expr.UnmarshalExpr([]byte(exprRaw))
	if err != nil {
		return WorkflowNodeConfigForBranchBlock{}
	}

	return WorkflowNodeConfigForBranchBlock{
		Expression: expr,
	}
}

func (c WorkflowNodeConfig) AsBizApply() WorkflowNodeConfigForBizApply {
	domains := lo.Filter(strings.Split(xmaps.GetString(c, "domains"), ";"), func(s string, _ int) bool { return s != "" })
	ipaddrs := lo.Filter(strings.Split(xmaps.GetString(c, "ipaddrs"), ";"), func(s string, _ int) bool { return s != "" })
	nameservers := lo.Filter(strings.Split(xmaps.GetString(c, "nameservers"), ";"), func(s string, _ int) bool { return s != "" })

	return WorkflowNodeConfigForBizApply{
		Domains:               domains,
		IPAddrs:               ipaddrs,
		ContactEmail:          xmaps.GetString(c, "contactEmail"),
		ChallengeType:         xmaps.GetString(c, "challengeType"),
		Provider:              xmaps.GetString(c, "provider"),
		ProviderAccessId:      xmaps.GetString(c, "providerAccessId"),
		ProviderConfig:        xmaps.GetKVMapAny(c, "providerConfig"),
		KeySource:             xmaps.GetOrDefaultString(c, "keySource", "auto"),
		KeyAlgorithm:          xmaps.GetOrDefaultString(c, "keyAlgorithm", string(CertificateKeyAlgorithmTypeRSA2048)),
		KeyContent:            xmaps.GetString(c, "keyContent"),
		CAProvider:            xmaps.GetString(c, "caProvider"),
		CAProviderAccessId:    xmaps.GetString(c, "caProviderAccessId"),
		CAProviderConfig:      xmaps.GetKVMapAny(c, "caProviderConfig"),
		ValidityLifetime:      xmaps.GetString(c, "validityLifetime"),
		PreferredChain:        xmaps.GetString(c, "preferredChain"),
		ACMEProfile:           xmaps.GetString(c, "acmeProfile"),
		Nameservers:           nameservers,
		DnsPropagationWait:    xmaps.GetInt(c, "dnsPropagationWait"),
		DnsPropagationTimeout: xmaps.GetInt(c, "dnsPropagationTimeout"),
		DnsTTL:                xmaps.GetInt(c, "dnsTTL"),
		HttpDelayWait:         xmaps.GetInt(c, "httpDelayWait"),
		DisableCommonName:     xmaps.GetBool(c, "disableCommonName"),
		DisableFollowCNAME:    xmaps.GetBool(c, "disableFollowCNAME"),
		DisableARI:            xmaps.GetBool(c, "disableARI"),
		SkipBeforeExpiryDays:  xmaps.GetInt(c, "skipBeforeExpiryDays"),
	}
}

func (c WorkflowNodeConfig) AsBizUpload() WorkflowNodeConfigForBizUpload {
	return WorkflowNodeConfigForBizUpload{
		Source:      xmaps.GetOrDefaultString(c, "source", "form"),
		Certificate: xmaps.GetString(c, "certificate"),
		PrivateKey:  xmaps.GetString(c, "privateKey"),
	}
}

func (c WorkflowNodeConfig) AsBizMonitor() WorkflowNodeConfigForBizMonitor {
	host := xmaps.GetString(c, "host")
	return WorkflowNodeConfigForBizMonitor{
		Host:        host,
		Port:        xmaps.GetOrDefaultInt32(c, "port", 443),
		Domain:      xmaps.GetOrDefaultString(c, "domain", host),
		RequestPath: xmaps.GetString(c, "path"),
	}
}

func (c WorkflowNodeConfig) AsBizDeploy() WorkflowNodeConfigForBizDeploy {
	return WorkflowNodeConfigForBizDeploy{
		CertificateOutputNodeId: xmaps.GetString(c, "certificateOutputNodeId"),
		Provider:                xmaps.GetString(c, "provider"),
		ProviderAccessId:        xmaps.GetString(c, "providerAccessId"),
		ProviderConfig:          xmaps.GetKVMapAny(c, "providerConfig"),
		SkipOnLastSucceeded:     xmaps.GetBool(c, "skipOnLastSucceeded"),
	}
}

func (c WorkflowNodeConfig) AsBizNotify() WorkflowNodeConfigForBizNotify {
	return WorkflowNodeConfigForBizNotify{
		Provider:             xmaps.GetString(c, "provider"),
		ProviderAccessId:     xmaps.GetString(c, "providerAccessId"),
		ProviderConfig:       xmaps.GetKVMapAny(c, "providerConfig"),
		Subject:              xmaps.GetString(c, "subject"),
		Message:              xmaps.GetString(c, "message"),
		SkipOnAllPrevSkipped: xmaps.GetBool(c, "skipOnAllPrevSkipped"),
	}
}

type WorkflowNodeConfigForDelay struct {
	Wait int `json:"wait"` // 等待时间
}

type WorkflowNodeConfigForBranchBlock struct {
	Expression expr.Expr `json:"expression"` // 条件表达式
}

type WorkflowNodeConfigForBizApply struct {
	Domains               []string       `json:"domains"`                         // 域名列表，以半角分号分隔
	IPAddrs               []string       `json:"ipaddrs"`                         // IP 地址列表，以半角分号分隔
	ContactEmail          string         `json:"contactEmail"`                    // 联系邮箱
	ChallengeType         string         `json:"challengeType"`                   // 质询方式
	Provider              string         `json:"provider"`                        // 质询提供商
	ProviderAccessId      string         `json:"providerAccessId"`                // 质询提供商授权记录 ID
	ProviderConfig        map[string]any `json:"providerConfig,omitempty"`        // 质询提供商额外配置
	CAProvider            string         `json:"caProvider,omitempty"`            // CA 提供商（零值时使用全局配置）
	CAProviderAccessId    string         `json:"caProviderAccessId,omitempty"`    // CA 提供商授权记录 ID
	CAProviderConfig      map[string]any `json:"caProviderConfig,omitempty"`      // CA 提供商额外配置
	KeySource             string         `json:"keySource"`                       // 私钥来源，可取值 "auto"、"reuse"、"custom"（零值时默认值 "auto"）
	KeyAlgorithm          string         `json:"keyAlgorithm,omitempty"`          // 私钥算法
	KeyContent            string         `json:"keyContent,omitempty"`            // 私钥内容
	ValidityLifetime      string         `json:"validityLifetime,omitempty"`      // 有效期，形如 "30d"、"6h"
	PreferredChain        string         `json:"preferredChain,omitempty"`        // 首选证书链
	ACMEProfile           string         `json:"acmeProfile,omitempty"`           // ACME Profiles Extension
	Nameservers           []string       `json:"nameservers,omitempty"`           // DNS 服务器列表，以半角分号分隔
	DnsPropagationWait    int            `json:"dnsPropagationWait,omitempty"`    // DNS 传播等待时间，等同于 lego 的 `--dns.propagation-wait` 参数
	DnsPropagationTimeout int            `json:"dnsPropagationTimeout,omitempty"` // DNS 传播检查超时时间（零值时使用提供商的默认值）
	DnsTTL                int            `json:"dnsTTL,omitempty"`                // DNS 解析记录 TTL（零值时使用提供商的默认值）
	HttpDelayWait         int            `json:"httpDelayWait,omitempty"`         // HTTP 等待时间，等同于 lego 的 `--http.delay` 参数
	DisableCommonName     bool           `json:"disableCommonName,omitempty"`     // 是否不包含 CommonName
	DisableFollowCNAME    bool           `json:"disableFollowCNAME,omitempty"`    // 是否关闭 CNAME 跟随
	DisableARI            bool           `json:"disableARI,omitempty"`            // 是否关闭 ARI
	SkipBeforeExpiryDays  int            `json:"skipBeforeExpiryDays,omitempty"`  // 证书到期前多少天前跳过续期
}

type WorkflowNodeConfigForBizUpload struct {
	Source      string `json:"source"`      // 证书来源，可取值 "form"、"local"、"url"（零值时默认值 "form"）
	Certificate string `json:"certificate"` // 证书，根据证书来源决定是 PEM 内容 / 文件路径 / URL
	PrivateKey  string `json:"privateKey"`  // 私钥，根据证书来源决定是 PEM 内容 / 文件路径 / URL
}

type WorkflowNodeConfigForBizMonitor struct {
	Host        string `json:"host"`                  // 主机地址
	Port        int32  `json:"port,omitempty"`        // 端口（零值时默认值 443）
	Domain      string `json:"domain,omitempty"`      // 域名（零值时默认值 [Host]）
	RequestPath string `json:"requestPath,omitempty"` // 请求路径
}

type WorkflowNodeConfigForBizDeploy struct {
	CertificateOutputNodeId string         `json:"certificateOutputNodeId"`    // 前序证书输出节点 ID
	Provider                string         `json:"provider"`                   // 主机提供商
	ProviderAccessId        string         `json:"providerAccessId,omitempty"` // 主机提供商授权记录 ID
	ProviderConfig          map[string]any `json:"providerConfig,omitempty"`   // 主机提供商额外配置
	SkipOnLastSucceeded     bool           `json:"skipOnLastSucceeded"`        // 上次部署成功时是否跳过
}

type WorkflowNodeConfigForBizNotify struct {
	Provider             string         `json:"provider"`                 // 通知提供商
	ProviderAccessId     string         `json:"providerAccessId"`         // 通知提供商授权记录 ID
	ProviderConfig       map[string]any `json:"providerConfig,omitempty"` // 通知提供商额外配置
	Subject              string         `json:"subject"`                  // 通知主题
	Message              string         `json:"message"`                  // 通知内容
	SkipOnAllPrevSkipped bool           `json:"skipOnAllPrevSkipped"`     // 前序节点均已跳过时是否跳过
}
