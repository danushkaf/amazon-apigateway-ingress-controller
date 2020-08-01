package cloudformation

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/awslabs/amazon-apigateway-ingress-controller/pkg/network"
	cfn "github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/apigateway"
	"github.com/awslabs/goformation/v4/cloudformation/ec2"
	"github.com/awslabs/goformation/v4/cloudformation/elasticloadbalancingv2"
	"github.com/awslabs/goformation/v4/cloudformation/iam"
	"github.com/awslabs/goformation/v4/cloudformation/route53"
	"github.com/awslabs/goformation/v4/cloudformation/tags"
	"github.com/awslabs/goformation/v4/cloudformation/wafv2"

	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

//const is constance values for resource naming used to build cf templates
const (
	AWSStackName                            = "AWS::StackName"
	AWSRegion                               = "AWS::Region"
	APIMethodResourceID                     = "Method"
	APIKeyResourceName                      = "APIKey"
	APIKeyUsagePlanResourceName             = "APIKeyUsagePlan"
	APIRootResourceResourceID               = "RootResourceId"
	APIResourceResourceName                 = "Resource"
	APIResourceName                         = "RestAPI"
	APIAuthorizerResourceName               = "RestAPIAuthorizer"
	CustomDomainResourceName                = "CustomDomain"
	CustomDomainBasePathMappingResourceName = "CustomDomainBasePathMapping"
	DeploymentResourceName                  = "Deployment"
	DistributionDomainNameResourceName      = "DistributionDomainName"
	DistributionHostedZoneIdResourceName    = "DistributionHostedZoneId"
	LambdaInvokeRoleResourceName            = "LambdaInvokeRole"
	LoadBalancerResourceName                = "LoadBalancer"
	ListnerResourceName                     = "Listener"
	RegionalDomainNameResourceName          = "RegionalDomainName"
	RegionalHostedZoneIdResourceName        = "RegionalHostedZoneId"
	SecurityGroupIngressResourceName        = "SecurityGroupIngress"
	TargetGroupResourceName                 = "TargetGroup"
	UsagePlanResourceName                   = "UsagePlan"
	VPCLinkResourceName                     = "VPCLink"
	WAFACLResourceName                      = "WAFAcl"
	WAFAssociationResourceName              = "WAFAssociation"
	Route53RecordResourceName               = "Route53RecordSet"
	OutputKeyRestAPIID                      = "RestAPIID"
	OutputKeyAPIGatewayEndpoint             = "APIGatewayEndpoint"
	OutputKeyAPIEndpointType                = "APIGWEndpointType"
	OutputKeyClientARNS                     = "ClientARNS"
	OutputKeyCertARN                        = "SSLCertArn"
	OutputKeyCustomDomain                   = "CustomDomainName"
	OutputKeyCustomDomainBasePath           = "CustomDomainBasePath"
	OutputMinimumCompressionSize            = "MinimumCompressionSize"
	OutputKeyWAFEnabled                     = "WAFEnabled"
	OutputKeyWAFRules                       = "WAFRules"
	OutputKeyWAFScope                       = "WAFScope"
	OutputKeyWAFAssociationCreated          = "WAFAssociation"
	OutputKeyCustomDomainHostName           = "CustomDomainHostname"
	OutputKeyCustomDomainHostedZoneID       = "CustomDomainHostedZoneID"
	OutputKeyHostedZone                     = "HostedZone"
	OutputKeyRequestTimeout                 = "RequestTimeout"
	OutputKeyTLSPolicy                      = "TLSPolicy"
	OutputKeyUsagePlans                     = "UsagePlansData"
	OutputKeyCachingEnabled                 = "CachingEnabled"
	OutputKeyCacheClusterSize               = "CachingSize"
	OutputKeyAPIResources                   = "APIResources"
	OutputKeyAWSAPIConfigs                  = "AWSAPIConfigs"
)

func toLogicalName(idx int, parts []string) string {
	s := strings.Join(parts[:idx+1], "")
	remove := []string{"{", "}", "+", "-", "*", "_"}
	for _, char := range remove {
		s = strings.Replace(s, char, "", -1)
	}
	return s
}

func toPath(idx int, parts []string) string {
	if parts[idx] == "{proxy+}" {
		return strings.Join(parts[:idx], "/") + "/{proxy}"
	}
	return strings.Join(parts[:idx+1], "/")
}

func mapAPIGWMethodsAndResourcesFromDefinedPublicAPIs(resources []APIResource, requestTimeout int, authorizationType string, index int) map[string]cfn.Resource {
	m := map[string]cfn.Resource{}

	for _, resource := range resources {
		parts := strings.Split(resource.Path, "/")
		for idx, part := range parts {
			if idx == 0 {
				continue
			}
			ref := cfn.GetAtt(fmt.Sprintf("%s%d", APIResourceName, index), fmt.Sprintf("%s%d", APIRootResourceResourceID, index))
			if idx > 1 {
				ref = cfn.Ref(fmt.Sprintf("%s%s%d", APIResourceResourceName, toLogicalName(idx-1, parts), index))
			}

			resourceLogicalName := fmt.Sprintf("%s%s%d", APIResourceResourceName, toLogicalName(idx, parts), index)
			m[resourceLogicalName] = buildAWSApiGatewayResource(ref, part, index)
			if idx == len(parts)-1 {
				for _, method := range resource.Methods {
					m[fmt.Sprintf("%s%s%s%d", APIMethodResourceID, toLogicalName(idx, parts), method, index)] = buildAWSApiGatewayMethod(resourceLogicalName, toPath(idx, parts), requestTimeout, authorizationType, method, resource, index)
				}
			}
		}
	}

	return m
}

func mapApiGatewayMethodsAndResourcesFromPaths(paths []extensionsv1beta1.HTTPIngressPath, requestTimeout int, authorizationType string, index int) map[string]cfn.Resource {
	m := map[string]cfn.Resource{}

	for _, path := range paths {
		parts := strings.Split(path.Path, "/")
		parts = append(parts, "{proxy+}")
		for idx, part := range parts {
			if idx == 0 {
				continue
			}
			ref := cfn.GetAtt(fmt.Sprintf("%s%d", APIResourceName, index), fmt.Sprintf("%s%d", APIRootResourceResourceID, index))
			if idx > 1 {
				ref = cfn.Ref(fmt.Sprintf("%s%s%d", APIResourceResourceName, toLogicalName(idx-1, parts), index))
			}

			resourceLogicalName := fmt.Sprintf("%s%s%d", APIResourceResourceName, toLogicalName(idx, parts), index)
			m[resourceLogicalName] = buildAWSApiGatewayResource(ref, part, index)
			m[fmt.Sprintf("%s%s%d", APIMethodResourceID, toLogicalName(idx, parts), index)] = buildAWSApiGatewayMethod(resourceLogicalName, toPath(idx, parts), requestTimeout, authorizationType, "ANY", APIResource{}, index)
		}
	}

	return m
}

func buildAWSApiGatewayResource(ref, part string, index int) *apigateway.Resource {
	return &apigateway.Resource{
		ParentId:  ref,
		PathPart:  part,
		RestApiId: cfn.Ref(fmt.Sprintf("%s%d", APIResourceName, index)),
	}
}

func buildAWSApiGatewayRestAPI(arns []string, apiEPType string, authorizationType string, minimumCompressionSize int, apiName string) *apigateway.RestApi {
	if authorizationType == "AWS_IAM" && minimumCompressionSize > 0 {
		return &apigateway.RestApi{
			MinimumCompressionSize: minimumCompressionSize,
			ApiKeySourceType:       "HEADER",
			EndpointConfiguration: &apigateway.RestApi_EndpointConfiguration{
				Types: []string{apiEPType},
			},
			Name: apiName,
			Policy: &PolicyDocument{
				Version: "2012-10-17",
				Statement: []Statement{
					{
						Action:    []string{"execute-api:Invoke"},
						Effect:    "Allow",
						Principal: map[string][]string{"AWS": arns},
						Resource:  []string{"*"},
					},
				},
			},
		}
	} else if authorizationType == "AWS_IAM" && minimumCompressionSize == 0 {
		return &apigateway.RestApi{
			ApiKeySourceType: "HEADER",
			EndpointConfiguration: &apigateway.RestApi_EndpointConfiguration{
				Types: []string{apiEPType},
			},
			Name: apiName,
			Policy: &PolicyDocument{
				Version: "2012-10-17",
				Statement: []Statement{
					{
						Action:    []string{"execute-api:Invoke"},
						Effect:    "Allow",
						Principal: map[string][]string{"AWS": arns},
						Resource:  []string{"*"},
					},
				},
			},
		}
	} else if minimumCompressionSize > 0 {
		return &apigateway.RestApi{
			MinimumCompressionSize: minimumCompressionSize,
			ApiKeySourceType:       "HEADER",
			EndpointConfiguration: &apigateway.RestApi_EndpointConfiguration{
				Types: []string{apiEPType},
			},
			Name: apiName,
			Policy: &AllPrinciplesPolicyDocument{
				Version: "2012-10-17",
				Statement: []AllPrinciplesStatement{
					{
						Action:    []string{"execute-api:Invoke"},
						Effect:    "Allow",
						Principal: "*",
						Resource:  []string{"*"},
					},
				},
			},
		}
	} else {
		return &apigateway.RestApi{
			ApiKeySourceType: "HEADER",
			EndpointConfiguration: &apigateway.RestApi_EndpointConfiguration{
				Types: []string{apiEPType},
			},
			Name: apiName,
			Policy: &AllPrinciplesPolicyDocument{
				Version: "2012-10-17",
				Statement: []AllPrinciplesStatement{
					{
						Action:    []string{"execute-api:Invoke"},
						Effect:    "Allow",
						Principal: "*",
						Resource:  []string{"*"},
					},
				},
			},
		}
	}
}

type EmptyAction struct{}

func buildAWSWAFWebACL(webACLScope string, rules string) *wafv2.WebACL {
	waf := &wafv2.WebACL{
		Name:        cfn.Ref(AWSStackName),
		Scope:       webACLScope,
		Description: "This is an example WebACL",
		DefaultAction: &wafv2.WebACL_DefaultAction{
			Allow: EmptyAction{},
		},
		VisibilityConfig: &wafv2.WebACL_VisibilityConfig{
			SampledRequestsEnabled:   true,
			CloudWatchMetricsEnabled: true,
			MetricName:               cfn.Sub(fmt.Sprintf("${%s}WebACLMetric", AWSStackName)),
		},
	}

	if rules == "" {
		return waf
	}
	var wafRules []wafv2.WebACL_Rule
	if err := json.Unmarshal([]byte(rules), &wafRules); err != nil {
		return waf
	}
	waf.Rules = wafRules

	return waf
}

func buildAWSWAFWebACLAssociation(stage string, index int) *wafv2.WebACLAssociation {
	wafAssociation := &wafv2.WebACLAssociation{
		WebACLArn:   cfn.GetAtt(WAFACLResourceName, "Arn"),
		ResourceArn: cfn.Sub(fmt.Sprintf("arn:aws:apigateway:${%s}::/restapis/${%s}/stages/%s", AWSRegion, APIResourceName, stage)),
	}

	dependsOn := []string{fmt.Sprintf("%s%d", DeploymentResourceName, index), fmt.Sprintf("%s%d", WAFACLResourceName, index)}
	sort.Strings(dependsOn)
	wafAssociation.AWSCloudFormationDependsOn = dependsOn

	return wafAssociation
}

func buildResourcePath(path string) string {
	path = strings.Replace(path, "/", "~1", -1)
	path = fmt.Sprintf("%s%s", "/", path)
	return path
}

func buildAWSAPIGWDeploymentMethodSettings(cachingEnabled bool, apiResources []APIResource) []apigateway.Deployment_MethodSetting {
	methodSettings := make([]apigateway.Deployment_MethodSetting, 1)
	if cachingEnabled && apiResources != nil && len(apiResources) > 0 {
		for i, resource := range apiResources {
			for j, method := range resource.Methods {
				methodSetting := apigateway.Deployment_MethodSetting{
					ResourcePath:   buildResourcePath(resource.Path),
					HttpMethod:     method,
					CachingEnabled: resource.CachingEnabled,
				}
				if i == 0 && j == 0 {
					methodSettings[0] = methodSetting
				} else {
					methodSettings = append(methodSettings, methodSetting)
				}
			}
		}
	}
	return methodSettings
}

func buildAWSApiGatewayDeployment(stageName string, dependsOn []string, cachingEnabled bool, apiResources []APIResource, cacheSize string, index int) *apigateway.Deployment {
	d := &apigateway.Deployment{
		RestApiId: cfn.Ref(fmt.Sprintf("%s%d", APIResourceName, index)),
		StageName: stageName,
		StageDescription: &apigateway.Deployment_StageDescription{
			CacheClusterEnabled: cachingEnabled,
			CacheClusterSize:    cacheSize,
			CacheDataEncrypted:  cachingEnabled,
			MethodSettings:      buildAWSAPIGWDeploymentMethodSettings(cachingEnabled, apiResources),
		},
	}

	// Since we construct a map of in `mapApiGatewayMethodsAndResourcesFromPaths` we can't determine the order
	// that this list will be in - making it difficult to test - the order isn't important - but passing tests are.
	// This isn't the worst thing in the world - and - I considered refactoring - but I like how simple this is for now.
	// Also the order doesn't matter to CFN in the end.
	sort.Strings(dependsOn)
	d.AWSCloudFormationDependsOn = dependsOn

	return d
}

func buildAWSElasticLoadBalancingV2Listener() *elasticloadbalancingv2.Listener {
	return &elasticloadbalancingv2.Listener{
		LoadBalancerArn: cfn.Ref(LoadBalancerResourceName),
		Protocol:        "TCP",
		Port:            80,
		DefaultActions: []elasticloadbalancingv2.Listener_Action{
			elasticloadbalancingv2.Listener_Action{
				TargetGroupArn: cfn.Ref(TargetGroupResourceName),
				Type:           "forward",
			},
		},
	}
}

func buildAWSElasticLoadBalancingV2LoadBalancer(subnetIDs []string) *elasticloadbalancingv2.LoadBalancer {
	return &elasticloadbalancingv2.LoadBalancer{
		IpAddressType: "ipv4",
		Scheme:        "internal",
		Subnets:       subnetIDs,
		Tags: []tags.Tag{
			{
				Key:   "com.github.amazon-apigateway-ingress-controller/stack",
				Value: cfn.Ref(AWSStackName),
			},
		},
		Type: "network",
	}
}

func buildAWSElasticLoadBalancingV2TargetGroup(vpcID string, instanceIDs []string, nodePort int, dependsOn []string) *elasticloadbalancingv2.TargetGroup {
	targets := make([]elasticloadbalancingv2.TargetGroup_TargetDescription, len(instanceIDs))
	for i, instanceID := range instanceIDs {
		targets[i] = elasticloadbalancingv2.TargetGroup_TargetDescription{Id: instanceID}
	}

	return &elasticloadbalancingv2.TargetGroup{
		HealthCheckIntervalSeconds: 30,
		HealthCheckPort:            "traffic-port",
		HealthCheckProtocol:        "TCP",
		HealthCheckTimeoutSeconds:  10,
		HealthyThresholdCount:      3,
		Port:                       nodePort,
		Protocol:                   "TCP",
		Tags: []tags.Tag{
			{
				Key:   "com.github.amazon-apigateway-ingress-controller/stack",
				Value: cfn.Ref(AWSStackName),
			},
		},
		TargetType:              "instance",
		Targets:                 targets,
		UnhealthyThresholdCount: 3,
		VpcId:                   vpcID,
	}

}

func buildAWSApiGatewayVpcLink(dependsOn []string) *apigateway.VpcLink {
	r := &apigateway.VpcLink{
		Name:       cfn.Ref(AWSStackName),
		TargetArns: []string{cfn.Ref(LoadBalancerResourceName)},
	}

	r.AWSCloudFormationDependsOn = dependsOn

	return r
}

func buildAWSApiGatewayMethod(resourceLogicalName, path string, timeout int, authorizationType string, method string, resource APIResource, index int) *apigateway.Method {
	var apiKeyRequired bool

	if authorizationType == "AWS_IAM" {
		apiKeyRequired = false
	} else {
		apiKeyRequired = true
	}

	requestParams := make(map[string]bool)
	requestParams["method.request.path.proxy"] = true
	integrationRequestParams := make(map[string]string)
	integrationRequestParams["integration.request.path.proxy"] = "method.request.path.proxy"
	integrationRequestParams["integration.request.header.Accept-Encoding"] = "'identity'"
	if resource.Path != "" {
		if resource.ProxyPathParams != nil && len(resource.ProxyPathParams) > 0 {
			for _, param := range resource.ProxyPathParams {
				mathodVarName := fmt.Sprintf("method.request.path.%s", param.Param)
				intVarName := fmt.Sprintf("integration.request.path.%s", param.Param)
				requestParams[mathodVarName] = true
				integrationRequestParams[intVarName] = mathodVarName
			}
		}
		if resource.ProxyQueryParams != nil && len(resource.ProxyQueryParams) > 0 {
			for _, param := range resource.ProxyQueryParams {
				mathodVarName := fmt.Sprintf("method.request.query.%s", param.Param)
				intVarName := fmt.Sprintf("integration.request.query.%s", param.Param)
				requestParams[mathodVarName] = true
				integrationRequestParams[intVarName] = mathodVarName
			}
		}
		if resource.ProxyHeaderParams != nil && len(resource.ProxyHeaderParams) > 0 {
			for _, param := range resource.ProxyHeaderParams {
				mathodVarName := fmt.Sprintf("method.request.header.%s", param.Param)
				intVarName := fmt.Sprintf("integration.request.header.%s", param.Param)
				requestParams[mathodVarName] = true
				integrationRequestParams[intVarName] = mathodVarName
			}
		}
	}

	m := &apigateway.Method{
		RequestParameters: requestParams,
		AuthorizationType: authorizationType,
		ApiKeyRequired:    apiKeyRequired,
		HttpMethod:        method,
		ResourceId:        cfn.Ref(resourceLogicalName),
		RestApiId:         cfn.Ref(fmt.Sprintf("%s%d", APIResourceName, index)),
		Integration: &apigateway.Method_Integration{
			ConnectionId:          cfn.Ref(VPCLinkResourceName),
			ConnectionType:        "VPC_LINK",
			IntegrationHttpMethod: "ANY",
			PassthroughBehavior:   "WHEN_NO_MATCH",
			RequestParameters:     integrationRequestParams,
			Type:                  "HTTP_PROXY",
			TimeoutInMillis:       timeout,
			Uri:                   cfn.Join("", []string{"http://", cfn.GetAtt(LoadBalancerResourceName, "DNSName"), path}),
		},
	}

	m.AWSCloudFormationDependsOn = []string{LoadBalancerResourceName}
	return m
}

func buildAWSEC2SecurityGroupIngresses(securityGroupIds []string, cidr string, nodePort int) []*ec2.SecurityGroupIngress {
	sgIngresses := make([]*ec2.SecurityGroupIngress, len(securityGroupIds))
	for i, sgID := range securityGroupIds {
		sgIngresses[i] = &ec2.SecurityGroupIngress{
			IpProtocol: "TCP",
			CidrIp:     cidr,
			FromPort:   nodePort,
			ToPort:     nodePort,
			GroupId:    sgID,
		}
	}

	return sgIngresses
}

func buildCustomDomainBasePathMapping(domainName string, stageName string, basePath string, index int) *apigateway.BasePathMapping {
	var r *apigateway.BasePathMapping
	if basePath == "" {
		r = &apigateway.BasePathMapping{
			DomainName: domainName,
			RestApiId:  cfn.Ref(fmt.Sprintf("%s%d", APIResourceName, index)),
			Stage:      stageName,
		}
	} else {
		r = &apigateway.BasePathMapping{
			DomainName: domainName,
			RestApiId:  cfn.Ref(APIResourceName),
			Stage:      stageName,
			BasePath:   basePath,
		}
	}

	r.AWSCloudFormationDependsOn = []string{DeploymentResourceName}
	return r
}

func buildCustomDomain(domainName string, certificateArn string, apiEPType string, secPolicy string) *apigateway.DomainName {
	if apiEPType == "REGIONAL" {
		return &apigateway.DomainName{
			RegionalCertificateArn: certificateArn,
			DomainName:             domainName,
			SecurityPolicy:         secPolicy,
			EndpointConfiguration: &apigateway.DomainName_EndpointConfiguration{
				Types: []string{apiEPType},
			},
		}
	}
	return &apigateway.DomainName{
		CertificateArn: certificateArn,
		DomainName:     domainName,
		EndpointConfiguration: &apigateway.DomainName_EndpointConfiguration{
			Types: []string{apiEPType},
		},
	}

}

func buildUsagePlanAPIKeyMapping(usagePlan UsagePlan, i int, index int) []*apigateway.UsagePlanKey {
	if usagePlan.APIKeys == nil {
		return nil
	}
	arr := make([]*apigateway.UsagePlanKey, len(usagePlan.APIKeys))
	for k, _ := range usagePlan.APIKeys {
		arr[k] = &apigateway.UsagePlanKey{
			KeyId:       cfn.Ref(fmt.Sprintf("%s%d%d", fmt.Sprintf("%s%d", APIKeyResourceName, index), i, k)),
			KeyType:     "API_KEY",
			UsagePlanId: cfn.Ref(fmt.Sprintf("%s%d", UsagePlanResourceName, i)),
		}
	}
	return arr
}

func buildAPIKey(usagePlan UsagePlan, index int) []*apigateway.ApiKey {
	if usagePlan.APIKeys == nil {
		return nil
	}
	arr := make([]*apigateway.ApiKey, len(usagePlan.APIKeys))
	for k, key := range usagePlan.APIKeys {
		arr[k] = &apigateway.ApiKey{
			CustomerId:         key.CustomerID,
			GenerateDistinctId: key.GenerateDistinctID,
			Name:               fmt.Sprintf("%s%d", key.Name, index),
			Enabled:            true,
		}
	}
	return arr
}

func buildUsagePlan(usagePlan UsagePlan, stage string, index int) *apigateway.UsagePlan {
	r := &apigateway.UsagePlan{
		UsagePlanName: usagePlan.PlanName,
		Description:   usagePlan.Description,
		Quota: &apigateway.UsagePlan_QuotaSettings{
			Limit:  usagePlan.QuotaLimit,
			Offset: usagePlan.QuotaOffset,
			Period: usagePlan.QuotaPeriod,
		},
		Throttle: &apigateway.UsagePlan_ThrottleSettings{
			BurstLimit: usagePlan.ThrottleBurstLimit,
			RateLimit:  usagePlan.ThrottleRateLimit,
		},
	}

	r.ApiStages = buildMethodThrottling(usagePlan.MethodThrottlingParameters, stage, index)

	r.AWSCloudFormationDependsOn = []string{DeploymentResourceName}

	return r
}

func buildLambdaExecutionRole() *iam.Role {
	r := &iam.Role{
		RoleName:    cfn.Sub(fmt.Sprintf("${%s}-LambdaExecutionRole", AWSStackName)),
		Description: cfn.Sub(fmt.Sprintf("Lambda Execution Role for stack ${%s}", AWSStackName)),
		Path:        "/",
		ManagedPolicyArns: []string{
			"arn:aws:iam::aws:policy/service-role/AWSLambdaRole",
		},
		AssumeRolePolicyDocument: PolicyDocument{
			Version: "2012-10-17",
			Statement: []Statement{
				{
					Effect: "Allow",
					Principal: map[string][]string{
						"Service": []string{
							"apigateway.amazonaws.com",
							"lambda.amazonaws.com",
						},
					},
					Action: []string{
						"sts:AssumeRole",
					},
				},
			},
		},
	}

	return r
}

func buildAuthorizer(apiDef AWSAPIDefinition, index int) *apigateway.Authorizer {
	if apiDef.AuthorizerResultTtlInSeconds == 0 {
		apiDef.AuthorizerResultTtlInSeconds = 300
	}
	if apiDef.AuthorizerType == "COGNITO_USER_POOLS" {
		return &apigateway.Authorizer{
			AuthorizerResultTtlInSeconds: apiDef.AuthorizerResultTtlInSeconds,
			AuthType:                     apiDef.AuthorizerAuthType,
			IdentitySource:               apiDef.IdentitySource,
			Name:                         apiDef.Name,
			ProviderARNs:                 apiDef.ProviderARNs,
			RestApiId:                    cfn.Ref(fmt.Sprintf("%s%d", APIResourceName, index)),
			Type:                         apiDef.AuthorizerType,
		}
	} else if apiDef.AuthorizerType == "TOKEN" {
		return &apigateway.Authorizer{
			AuthorizerCredentials:        cfn.GetAtt(LambdaInvokeRoleResourceName, "Arn"),
			AuthorizerResultTtlInSeconds: apiDef.AuthorizerResultTtlInSeconds,
			AuthorizerUri:                cfn.Join("", []string{"arn:aws:apigateway:", cfn.Ref("AWS::Region"), ":lambda:path/2015-03-31/functions/", apiDef.AuthorizerUri, "/invocations"}),
			AuthType:                     apiDef.AuthorizerAuthType,
			IdentitySource:               apiDef.IdentitySource,
			IdentityValidationExpression: apiDef.IdentityValidationExpression,
			Name:                         apiDef.Name,
			RestApiId:                    cfn.Ref(fmt.Sprintf("%s%d", APIResourceName, index)),
			Type:                         apiDef.AuthorizerType,
		}
	} else {
		return &apigateway.Authorizer{
			AuthorizerCredentials:        cfn.GetAtt(LambdaInvokeRoleResourceName, "Arn"),
			AuthorizerResultTtlInSeconds: apiDef.AuthorizerResultTtlInSeconds,
			AuthorizerUri:                cfn.Join("", []string{"arn:aws:apigateway:", cfn.Ref("AWS::Region"), ":lambda:path/2015-03-31/functions/", apiDef.AuthorizerUri, "/invocations"}),
			AuthType:                     apiDef.AuthorizerAuthType,
			IdentitySource:               apiDef.IdentitySource,
			IdentityValidationExpression: apiDef.AuthorizerAuthType,
			Name:                         apiDef.Name,
			RestApiId:                    cfn.Ref(fmt.Sprintf("%s%d", APIResourceName, index)),
			Type:                         apiDef.AuthorizerType,
		}
	}
}

func buildMethodThrottling(methodThrottlingParameters []MethodThrottlingParametersObject, stage string, index int) []apigateway.UsagePlan_ApiStage {
	rmap := make(map[string]apigateway.UsagePlan_ThrottleSettings)
	if methodThrottlingParameters != nil && len(methodThrottlingParameters) > 0 {
		for _, methodThrottlingParameter := range methodThrottlingParameters {
			r := apigateway.UsagePlan_ThrottleSettings{
				BurstLimit: methodThrottlingParameter.BurstLimit,
				RateLimit:  methodThrottlingParameter.RateLimit,
			}
			var key string
			if strings.HasSuffix(methodThrottlingParameter.Path, "/") {
				key = fmt.Sprintf("%sANY", methodThrottlingParameter.Path)
			} else {
				key = fmt.Sprintf("%s/ANY", methodThrottlingParameter.Path)
			}
			rmap[key] = r
		}
	}

	stageResource := apigateway.UsagePlan_ApiStage{
		ApiId:    cfn.Ref(fmt.Sprintf("%s%d", APIResourceName, index)),
		Stage:    stage,
		Throttle: rmap,
	}

	stageResources := []apigateway.UsagePlan_ApiStage{stageResource}

	return stageResources
}

//TemplateConfig is the structure of configuration used to provide data to build the cf template
type TemplateConfig struct {
	Network                *network.Network
	Rule                   extensionsv1beta1.IngressRule
	NodePort               int
	StageName              string
	Arns                   []string
	CustomDomainName       string
	CustomDomainBasePath   string
	CertificateArn         string
	APIEndpointType        string
	WAFEnabled             bool
	WAFRulesJSON           string
	WAFScope               string
	WAFAssociation         bool
	RequestTimeout         int
	TLSPolicy              string
	UsagePlans             []UsagePlan
	MinimumCompressionSize int
	CachingEnabled         bool
	CachingSize            string
	APIResources           []APIResource
	AWSAPIDefinitions      []AWSAPIDefinition
}

// BuildAPIGatewayTemplateFromIngressRule generates the cloudformation template according to the config provided
func BuildAPIGatewayTemplateFromIngressRule(cfg *TemplateConfig) *cfn.Template {
	template := cfn.NewTemplate()
	paths := cfg.Rule.IngressRuleValue.HTTP.Paths
	publicAPIs := cfg.APIResources

	//Making default type edge
	if cfg.APIEndpointType == "" {
		cfg.APIEndpointType = "EDGE"
	}

	//Making default regional as cloudfront is not supported in all regions
	if cfg.WAFEnabled && cfg.WAFScope == "" {
		cfg.WAFScope = "REGIONAL"
	}

	if !cfg.CachingEnabled && cfg.CachingSize != "" {
		cfg.CachingEnabled = true
	}

	if cfg.CachingEnabled && cfg.CachingSize == "" {
		cfg.CachingSize = "0.5"
	}

	var authorizationType string
	if cfg.Arns != nil && len(cfg.Arns) > 0 {
		authorizationType = "AWS_IAM"
	} else {
		authorizationType = "NONE"
	}

	var apiSize int
	if cfg.AWSAPIDefinitions != nil && len(cfg.AWSAPIDefinitions) > 0 {
		apiSize = len(cfg.AWSAPIDefinitions)
	} else {
		apiSize = 1
	}

	if cfg.CustomDomainName != "" && cfg.CertificateArn != "" {
		customDomain := buildCustomDomain(cfg.CustomDomainName, cfg.CertificateArn, cfg.APIEndpointType, cfg.TLSPolicy)
		template.Resources[CustomDomainResourceName] = customDomain
	}

	lambdaInvokeRole := buildLambdaExecutionRole()
	template.Resources[LambdaInvokeRoleResourceName] = lambdaInvokeRole

	targetGroup := buildAWSElasticLoadBalancingV2TargetGroup(*cfg.Network.Vpc.VpcId, cfg.Network.InstanceIDs, cfg.NodePort, []string{LoadBalancerResourceName})
	template.Resources[TargetGroupResourceName] = targetGroup

	listener := buildAWSElasticLoadBalancingV2Listener()
	template.Resources[ListnerResourceName] = listener

	securityGroupIngresses := buildAWSEC2SecurityGroupIngresses(cfg.Network.SecurityGroupIDs, *cfg.Network.Vpc.CidrBlock, cfg.NodePort)
	for i, sgI := range securityGroupIngresses {
		template.Resources[fmt.Sprintf("%s%d", SecurityGroupIngressResourceName, i)] = sgI
	}

	if cfg.WAFEnabled {
		webACL := buildAWSWAFWebACL(cfg.WAFScope, cfg.WAFRulesJSON)
		template.Resources[WAFACLResourceName] = webACL
	}

	for i := 0; i < apiSize; i++ {

		methodLogicalNames := []string{}
		var resourceMap map[string]cfn.Resource
		if publicAPIs != nil && len(publicAPIs) > 0 {
			resourceMap = mapAPIGWMethodsAndResourcesFromDefinedPublicAPIs(publicAPIs, cfg.RequestTimeout, authorizationType, i)
		} else {
			resourceMap = mapApiGatewayMethodsAndResourcesFromPaths(paths, cfg.RequestTimeout, authorizationType, i)
		}

		for k, resource := range resourceMap {
			if _, ok := resource.(*apigateway.Method); ok {
				methodLogicalNames = append(methodLogicalNames, k)
			}
			template.Resources[k] = resource
		}

		if cfg.AWSAPIDefinitions != nil && len(cfg.AWSAPIDefinitions) > 0 && !cfg.AWSAPIDefinitions[i].AuthenticationEnabled {
			restAPI := buildAWSApiGatewayRestAPI(cfg.Arns, cfg.APIEndpointType, "NONE", cfg.MinimumCompressionSize, cfg.AWSAPIDefinitions[i].Name)
			template.Resources[fmt.Sprintf("%s%d", APIResourceName, i)] = restAPI
		} else if cfg.AWSAPIDefinitions != nil && len(cfg.AWSAPIDefinitions) > 0 {
			restAPI := buildAWSApiGatewayRestAPI(cfg.Arns, cfg.APIEndpointType, authorizationType, cfg.MinimumCompressionSize, cfg.AWSAPIDefinitions[i].Name)
			template.Resources[fmt.Sprintf("%s%d", APIResourceName, i)] = restAPI
		} else {
			restAPI := buildAWSApiGatewayRestAPI(cfg.Arns, cfg.APIEndpointType, authorizationType, cfg.MinimumCompressionSize, cfn.Ref(AWSStackName))
			template.Resources[fmt.Sprintf("%s%d", APIResourceName, i)] = restAPI
		}

		if cfg.AWSAPIDefinitions != nil && len(cfg.AWSAPIDefinitions) > 0 && cfg.AWSAPIDefinitions[i].Authorization_Enabled {
			//Authorizer Implementation
			authorizer := buildAuthorizer(cfg.AWSAPIDefinitions[i], i)
			template.Resources[fmt.Sprintf("%s%d", APIAuthorizerResourceName, i)] = authorizer
		}

		deployment := buildAWSApiGatewayDeployment(cfg.StageName, methodLogicalNames, cfg.CachingEnabled, cfg.APIResources, cfg.CachingSize, i)
		template.Resources[fmt.Sprintf("%s%d", DeploymentResourceName, i)] = deployment

		if cfg.CustomDomainName != "" && cfg.CertificateArn != "" {
			if cfg.AWSAPIDefinitions != nil && len(cfg.AWSAPIDefinitions) > 0 {
				basePathMapping := buildCustomDomainBasePathMapping(cfg.CustomDomainName, cfg.StageName, cfg.AWSAPIDefinitions[i].Context, i)
				template.Resources[fmt.Sprintf("%s%d", CustomDomainBasePathMappingResourceName, i)] = basePathMapping
			} else {
				basePathMapping := buildCustomDomainBasePathMapping(cfg.CustomDomainName, cfg.StageName, cfg.CustomDomainBasePath, i)
				template.Resources[fmt.Sprintf("%s%d", CustomDomainBasePathMappingResourceName, i)] = basePathMapping
			}
		}

		if cfg.WAFEnabled {
			if cfg.WAFAssociation {
				webACLAssociation := buildAWSWAFWebACLAssociation(cfg.StageName, i)
				template.Resources[fmt.Sprintf("%s%d", WAFAssociationResourceName, i)] = webACLAssociation
			}
		}

		if cfg.AWSAPIDefinitions != nil && len(cfg.AWSAPIDefinitions) > 0 {
			if cfg.AWSAPIDefinitions[i].APIKeyEnabled {
				if cfg.AWSAPIDefinitions[i].UsagePlans != nil && len(cfg.AWSAPIDefinitions[i].UsagePlans) > 0 {
					for j, usagePlan := range cfg.AWSAPIDefinitions[i].UsagePlans {
						keyArr := buildAPIKey(usagePlan, i)
						for k, key := range keyArr {
							template.Resources[fmt.Sprintf("%s%d%d%d", APIKeyResourceName, j, k, i)] = key
						}
						template.Resources[fmt.Sprintf("%s%d", UsagePlanResourceName, i)] = buildUsagePlan(usagePlan, cfg.StageName, i)
						mapArr := buildUsagePlanAPIKeyMapping(usagePlan, j, i)
						for k, key := range mapArr {
							template.Resources[fmt.Sprintf("%s%d%d%d", APIKeyUsagePlanResourceName, j, k, i)] = key
						}
					}
				} else {
					for j, usagePlan := range cfg.UsagePlans {
						keyArr := buildAPIKey(usagePlan, i)
						for k, key := range keyArr {
							template.Resources[fmt.Sprintf("%s%d%d%d", APIKeyResourceName, j, k, i)] = key
						}
						template.Resources[fmt.Sprintf("%s%d", UsagePlanResourceName, i)] = buildUsagePlan(usagePlan, cfg.StageName, i)
						mapArr := buildUsagePlanAPIKeyMapping(usagePlan, j, i)
						for k, key := range mapArr {
							template.Resources[fmt.Sprintf("%s%d%d%d", APIKeyUsagePlanResourceName, j, k, i)] = key
						}
					}
				}
			}
		} else if cfg.UsagePlans != nil && len(cfg.UsagePlans) > 0 {
			for j, usagePlan := range cfg.UsagePlans {
				keyArr := buildAPIKey(usagePlan, i)
				for k, key := range keyArr {
					template.Resources[fmt.Sprintf("%s%d%d%d", APIKeyResourceName, j, k, i)] = key
				}
				template.Resources[fmt.Sprintf("%s%d", UsagePlanResourceName, i)] = buildUsagePlan(usagePlan, cfg.StageName, i)
				mapArr := buildUsagePlanAPIKeyMapping(usagePlan, j, i)
				for k, key := range mapArr {
					template.Resources[fmt.Sprintf("%s%d%d%d", APIKeyUsagePlanResourceName, j, k, i)] = key
				}
			}
		}

	}

	loadBalancer := buildAWSElasticLoadBalancingV2LoadBalancer(cfg.Network.SubnetIDs)
	template.Resources[LoadBalancerResourceName] = loadBalancer

	vPCLink := buildAWSApiGatewayVpcLink([]string{LoadBalancerResourceName})
	template.Resources[VPCLinkResourceName] = vPCLink

	template.Outputs = map[string]interface{}{
		OutputKeyAPIEndpointType: Output{Value: cfg.APIEndpointType},
		OutputKeyRequestTimeout:  Output{Value: fmt.Sprintf("%d", cfg.RequestTimeout)},
	}

	for i := 0; i < apiSize; i++ {
		template.Outputs[fmt.Sprintf("%s%d", OutputKeyRestAPIID, i)] = Output{Value: cfn.Ref(fmt.Sprintf("%s%d", APIResourceName, i))}
		template.Outputs[fmt.Sprintf("%s%d", OutputKeyAPIGatewayEndpoint, i)] = Output{Value: cfn.Join("", []string{"https://", cfn.Ref(fmt.Sprintf("%s%d", APIResourceName, i)), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", cfg.StageName})}

		if cfg.WAFAssociation {
			template.Outputs[fmt.Sprintf("%s%d", OutputKeyWAFAssociationCreated, i)] = Output{Value: cfn.Ref(fmt.Sprintf("%s%d", WAFAssociationResourceName, i))}
		}
	}

	if cfg.UsagePlans != nil && len(cfg.UsagePlans) > 0 {
		val, _ := json.Marshal(cfg.UsagePlans)
		template.Outputs[OutputKeyUsagePlans] = Output{Value: string(val)}
	}

	if cfg.Arns != nil && len(cfg.Arns) > 0 {
		template.Outputs[OutputKeyClientARNS] = Output{Value: strings.Join(cfg.Arns, ",")}
	}

	if cfg.MinimumCompressionSize > 0 {
		template.Outputs[OutputMinimumCompressionSize] = Output{fmt.Sprintf("%d", cfg.MinimumCompressionSize)}
	}

	if cfg.AWSAPIDefinitions != nil && len(cfg.AWSAPIDefinitions) > 0 {
		val, _ := json.Marshal(cfg.AWSAPIDefinitions)
		template.Outputs[OutputKeyAWSAPIConfigs] = Output{Value: string(val)}
	}

	if cfg.WAFEnabled {
		template.Outputs[OutputKeyWAFEnabled] = Output{Value: fmt.Sprintf("%t", cfg.WAFEnabled)}
		template.Outputs[OutputKeyWAFRules] = Output{Value: cfg.WAFRulesJSON}
		template.Outputs[OutputKeyWAFScope] = Output{Value: cfg.WAFScope}
	}

	if cfg.APIEndpointType == "REGIONAL" && cfg.CustomDomainName != "" {
		template.Outputs[OutputKeyCertARN] = Output{Value: cfg.CertificateArn}
		template.Outputs[OutputKeyCustomDomain] = Output{Value: cfg.CustomDomainName}
		template.Outputs[OutputKeyCustomDomainHostName] = Output{Value: cfn.GetAtt(CustomDomainResourceName, RegionalDomainNameResourceName)}
		template.Outputs[OutputKeyCustomDomainHostedZoneID] = Output{Value: cfn.GetAtt(CustomDomainResourceName, RegionalHostedZoneIdResourceName)}
		template.Outputs[OutputKeyTLSPolicy] = Output{Value: cfg.TLSPolicy}
		template.Outputs[OutputKeyCustomDomainBasePath] = Output{Value: cfg.CustomDomainBasePath}
	}

	if cfg.APIEndpointType == "EDGE" && cfg.CustomDomainName != "" {
		template.Outputs[OutputKeyCertARN] = Output{Value: cfg.CertificateArn}
		template.Outputs[OutputKeyCustomDomain] = Output{Value: cfg.CustomDomainName}
		template.Outputs[OutputKeyCustomDomainHostName] = Output{Value: cfn.GetAtt(CustomDomainResourceName, DistributionDomainNameResourceName)}
		template.Outputs[OutputKeyCustomDomainHostedZoneID] = Output{Value: cfn.GetAtt(CustomDomainResourceName, DistributionHostedZoneIdResourceName)}
		template.Outputs[OutputKeyTLSPolicy] = Output{Value: cfg.TLSPolicy}
		template.Outputs[OutputKeyCustomDomainBasePath] = Output{Value: cfg.CustomDomainBasePath}
	}

	if cfg.CachingEnabled && cfg.CachingSize != "" {
		template.Outputs[OutputKeyCachingEnabled] = Output{Value: fmt.Sprintf("%t", cfg.CachingEnabled)}
		template.Outputs[OutputKeyCacheClusterSize] = Output{Value: cfg.CachingSize}
	}

	if cfg.APIResources != nil && len(cfg.APIResources) > 0 {
		val, _ := json.Marshal(cfg.APIResources)
		template.Outputs[OutputKeyAPIResources] = Output{Value: string(val)}
	}

	return template
}

func buildCustomDomainRoute53Record(domainName string, hostedZoneName string, dnsName string, hostedZoneID string) *route53.RecordSet {
	return &route53.RecordSet{
		Name:           domainName,
		HostedZoneName: hostedZoneName,
		Type:           "A",
		AliasTarget: &route53.RecordSet_AliasTarget{
			DNSName:      dnsName,
			HostedZoneId: hostedZoneID,
		},
	}
}

//Route53TemplateConfig is the structure of configuration used to provide data to build the cf template of route53
type Route53TemplateConfig struct {
	CustomDomainName         string
	CustomDomainHostName     string
	CustomDomainHostedZoneID string
	HostedZoneName           string
}

// BuildAPIGatewayRoute53Template generates the cloudformation template according to the config provided
func BuildAPIGatewayRoute53Template(cfg *Route53TemplateConfig) *cfn.Template {
	route53Template := cfn.NewTemplate()

	if cfg.HostedZoneName != "" {
		recordSet := buildCustomDomainRoute53Record(cfg.CustomDomainName, cfg.HostedZoneName, cfg.CustomDomainHostName, cfg.CustomDomainHostedZoneID)
		route53Template.Resources[Route53RecordResourceName] = recordSet
	}

	route53Template.Outputs = map[string]interface{}{
		OutputKeyCustomDomainHostName:     Output{Value: cfg.CustomDomainHostName},
		OutputKeyCustomDomainHostedZoneID: Output{Value: cfg.CustomDomainHostedZoneID},
		OutputKeyCustomDomain:             Output{Value: cfg.CustomDomainName},
		OutputKeyHostedZone:               Output{Value: cfg.HostedZoneName},
	}

	return route53Template
}
