package clusters

import (
	"github.com/TeaOSLab/EdgeAdmin/internal/utils"
	"github.com/TeaOSLab/EdgeAdmin/internal/web/actions/actionutils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

type ClusterAction struct {
	actionutils.ParentAction
}

func (this *ClusterAction) Init() {
	this.Nav("", "", "")
}

func (this *ClusterAction) RunGet(params struct {
	ClusterId int64
}) {
	// 集群信息
	clusterResp, err := this.RPC().NodeClusterRPC().FindEnabledNodeCluster(this.AdminContext(), &pb.FindEnabledNodeClusterRequest{NodeClusterId: params.ClusterId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	cluster := clusterResp.NodeCluster
	if cluster == nil {
		this.NotFound("nodeCluster", params.ClusterId)
		return
	}
	this.Data["cluster"] = maps.Map{
		"id":   cluster.Id,
		"name": cluster.Name,
	}

	// DNS信息
	dnsResp, err := this.RPC().NodeClusterRPC().FindEnabledNodeClusterDNS(this.AdminContext(), &pb.FindEnabledNodeClusterDNSRequest{NodeClusterId: params.ClusterId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	domainName := ""
	dnsMap := maps.Map{
		"dnsName":          dnsResp.Name,
		"domainId":         0,
		"domainName":       "",
		"providerId":       0,
		"providerName":     "",
		"providerTypeName": "",
	}
	if dnsResp.Domain != nil {
		domainName = dnsResp.Domain.Name
		dnsMap["domainId"] = dnsResp.Domain.Id
		dnsMap["domainName"] = dnsResp.Domain.Name
	}
	if dnsResp.Provider != nil {
		dnsMap["providerId"] = dnsResp.Provider.Id
		dnsMap["providerName"] = dnsResp.Provider.Name
		dnsMap["providerTypeName"] = dnsResp.Provider.TypeName
	}

	this.Data["dnsInfo"] = dnsMap

	// 节点DNS解析记录
	nodesResp, err := this.RPC().NodeRPC().FindAllEnabledNodesDNSWithNodeClusterId(this.AdminContext(), &pb.FindAllEnabledNodesDNSWithNodeClusterIdRequest{NodeClusterId: params.ClusterId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	nodeMaps := []maps.Map{}
	for _, node := range nodesResp.Nodes {
		if len(node.Routes) > 0 {
			for _, route := range node.Routes {
				// 检查是否已解析
				isResolved := false
				if cluster.DnsDomainId > 0 && len(cluster.DnsName) > 0 && len(node.IpAddr) > 0 {
					recordType := "A"
					if utils.IsIPv6(node.IpAddr) {
						recordType = "AAAA"
					}
					checkResp, err := this.RPC().DNSDomainRPC().ExistDNSDomainRecord(this.AdminContext(), &pb.ExistDNSDomainRecordRequest{
						DnsDomainId: cluster.DnsDomainId,
						Name:        cluster.DnsName,
						Type:        recordType,
						Route:       route.Code,
						Value:       node.IpAddr,
					})
					if err != nil {
						this.ErrorPage(err)
						return
					}
					isResolved = checkResp.IsOk
				}

				nodeMaps = append(nodeMaps, maps.Map{
					"id":     node.Id,
					"name":   node.Name,
					"ipAddr": node.IpAddr,
					"route": maps.Map{
						"name": route.Name,
						"code": route.Code,
					},
					"clusterId":  node.NodeClusterId,
					"isResolved": isResolved,
				})
			}
		} else {
			nodeMaps = append(nodeMaps, maps.Map{
				"id":     node.Id,
				"name":   node.Name,
				"ipAddr": node.IpAddr,
				"route": maps.Map{
					"name": "",
					"code": "",
				},
				"clusterId":  node.NodeClusterId,
				"isResolved": false,
			})
		}
	}
	this.Data["nodes"] = nodeMaps

	// 代理服务解析记录
	serversResp, err := this.RPC().ServerRPC().FindAllEnabledServersDNSWithNodeClusterId(this.AdminContext(), &pb.FindAllEnabledServersDNSWithNodeClusterIdRequest{NodeClusterId: params.ClusterId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	serverMaps := []maps.Map{}
	for _, server := range serversResp.Servers {
		// 检查是否已解析
		isResolved := false
		if cluster.DnsDomainId > 0 && len(cluster.DnsName) > 0 && len(server.DnsName) > 0 && len(domainName) > 0 {
			checkResp, err := this.RPC().DNSDomainRPC().ExistDNSDomainRecord(this.AdminContext(), &pb.ExistDNSDomainRecordRequest{
				DnsDomainId: cluster.DnsDomainId,
				Name:        server.DnsName,
				Type:        "CNAME",
				Value:       cluster.DnsName + "." + domainName,
			})
			if err != nil {
				this.ErrorPage(err)
				return
			}
			isResolved = checkResp.IsOk
		}

		serverMaps = append(serverMaps, maps.Map{
			"id":         server.Id,
			"name":       server.Name,
			"dnsName":    server.DnsName,
			"isResolved": isResolved,
		})
	}
	this.Data["servers"] = serverMaps

	// 检查解析记录是否有变化
	checkChangesResp, err := this.RPC().NodeClusterRPC().CheckNodeClusterDNSChanges(this.AdminContext(), &pb.CheckNodeClusterDNSChangesRequest{NodeClusterId: params.ClusterId})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	this.Data["dnsHasChanges"] = checkChangesResp.IsChanged

	// 需要解决的问题
	issuesResp, err := this.RPC().DNSRPC().FindAllDNSIssues(this.AdminContext(), &pb.FindAllDNSIssuesRequest{
		NodeClusterId: params.ClusterId,
	})
	if err != nil {
		this.ErrorPage(err)
		return
	}
	issueMaps := []maps.Map{}
	for _, issue := range issuesResp.Issues {
		issueMaps = append(issueMaps, maps.Map{
			"target":      issue.Target,
			"targetId":    issue.TargetId,
			"type":        issue.Type,
			"description": issue.Description,
			"params":      issue.Params,
		})
	}
	this.Data["issues"] = issueMaps

	this.Show()
}
