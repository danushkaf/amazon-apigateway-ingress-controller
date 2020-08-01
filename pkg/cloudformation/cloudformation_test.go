package cloudformation

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/awslabs/amazon-apigateway-ingress-controller/pkg/network"
	cfn "github.com/awslabs/goformation/v4/cloudformation"
	"github.com/awslabs/goformation/v4/cloudformation/apigateway"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getUsagePlans() []UsagePlan {
	return []UsagePlan{
		{
			PlanName:    "Gold",
			Description: "20 requests for 1 min",
			APIKeys: []APIKey{
				{
					CustomerID:         "customer1",
					GenerateDistinctID: true,
					Name:               "cusKey1",
				},
				{
					CustomerID:         "customer2",
					GenerateDistinctID: true,
					Name:               "cusKey2",
				},
			},
			QuotaLimit:         100,
			QuotaPeriod:        "MONTH",
			ThrottleBurstLimit: 100,
			ThrottleRateLimit:  100,
			MethodThrottlingParameters: []MethodThrottlingParametersObject{
				{
					Path:       "/api/v1/foobar",
					BurstLimit: 100,
					RateLimit:  100,
				},
			},
		},
	}
}

func getUsagePlan() UsagePlan {
	return UsagePlan{
		PlanName:    "Gold",
		Description: "20 requests for 1 min",
		APIKeys: []APIKey{
			{
				CustomerID:         "customer1",
				GenerateDistinctID: true,
				Name:               "cusKey1",
			},
			{
				CustomerID:         "customer2",
				GenerateDistinctID: true,
				Name:               "cusKey2",
			},
		},
		QuotaLimit:         100,
		QuotaPeriod:        "MONTH",
		ThrottleBurstLimit: 100,
		ThrottleRateLimit:  100,
		MethodThrottlingParameters: []MethodThrottlingParametersObject{
			{
				Path:       "/api/v1/foobar",
				BurstLimit: 100,
				RateLimit:  100,
			},
		},
	}
}

func getSecondUsagePlans() []UsagePlan {
	return []UsagePlan{
		getSecondUsagePlan(),
	}
}

func getSecondUsagePlan() UsagePlan {
	return UsagePlan{
		PlanName:    "Gold",
		Description: "10 requests for 1 min",
		APIKeys: []APIKey{
			{
				CustomerID:         "customer1",
				GenerateDistinctID: true,
				Name:               "cusKey1",
			},
			{
				CustomerID:         "customer2",
				GenerateDistinctID: true,
				Name:               "cusKey2",
			},
		},
		QuotaLimit:         100,
		QuotaPeriod:        "MONTH",
		ThrottleBurstLimit: 100,
		ThrottleRateLimit:  100,
		MethodThrottlingParameters: []MethodThrottlingParametersObject{
			{
				Path:       "/api/v1/foobar",
				BurstLimit: 100,
				RateLimit:  100,
			},
		},
	}
}

func getUsagePlanBytes() string {
	usagePlan := getUsagePlans()
	usagePlanBytes, _ := json.Marshal(usagePlan)
	return string(usagePlanBytes)
}

func getAPIKeyMappingBuild(i int, k int, index int) *apigateway.UsagePlanKey {
	arr := buildUsagePlanAPIKeyMapping(getUsagePlan(), k, index)
	for k, key := range arr {
		if k == i {
			return key
		}
	}
	return nil
}

func getAPIKeyBuild(i int) *apigateway.ApiKey {
	arr := buildAPIKey(getUsagePlan(), 0)
	for k, key := range arr {
		if k == i {
			return key
		}
	}
	return nil
}

func getSecondAPIKeyMappingBuild(i int, k int, index int) *apigateway.UsagePlanKey {
	arr := buildUsagePlanAPIKeyMapping(getSecondUsagePlan(), k, index)
	for k, key := range arr {
		if k == i {
			return key
		}
	}
	return nil
}

func getSecondAPIKeyBuild(i int) *apigateway.ApiKey {
	arr := buildAPIKey(getSecondUsagePlan(), 0)
	for k, key := range arr {
		if k == i {
			return key
		}
	}
	return nil
}

func getAPIResources() []APIResource {
	return []APIResource{
		getAPIResource(),
	}
}

func getAPIDef() AWSAPIDefinition {
	return AWSAPIDefinition{
		Name:                         "api0",
		Context:                      "api0",
		IdentitySource:               "foo",
		AuthorizerType:               "REQUEST",
		AuthorizerAuthType:           "foo",
		AuthorizerName:               "foo",
		IdentityValidationExpression: "",
		AuthorizerUri:                "arn:bar",
		AuthenticationEnabled:        true,
		APIKeyEnabled:                true,
		Authorization_Enabled:        true,
		UsagePlans:                   getSecondUsagePlans(),
	}
}

func getAPIDefs() []AWSAPIDefinition {
	return []AWSAPIDefinition{
		getAPIDef(),
	}
}

func getAWSAPIDefBytes() string {
	awsDefs := getAPIDefs()
	awsDefsBytes, _ := json.Marshal(awsDefs)
	return string(awsDefsBytes)
}

func getAPIDefWOUsagePlans() AWSAPIDefinition {
	return AWSAPIDefinition{
		Name:                         "api1",
		Context:                      "api1",
		IdentitySource:               "foo",
		AuthorizerType:               "COGNITO_USER_POOLS",
		AuthorizerAuthType:           "foo",
		AuthorizerName:               "foo",
		IdentityValidationExpression: "",
		AuthorizerResultTtlInSeconds: 3600,
		AuthorizerUri:                "arn:bar",
		ProviderARNs: []string{
			"arn:foo",
		},
		AuthenticationEnabled: true,
		APIKeyEnabled:         true,
		Authorization_Enabled: true,
	}
}

func getAPIDefAPIKeyDisabled() AWSAPIDefinition {
	return AWSAPIDefinition{
		Name:                         "api2",
		Context:                      "api2",
		IdentitySource:               "foo",
		AuthorizerType:               "TOKEN",
		AuthorizerAuthType:           "foo",
		AuthorizerName:               "foo",
		IdentityValidationExpression: "",
		AuthorizerResultTtlInSeconds: 3600,
		AuthorizerUri:                "arn:bar",
		ProviderARNs: []string{
			"arn:foo",
		},
		AuthenticationEnabled: true,
		APIKeyEnabled:         false,
		Authorization_Enabled: true,
	}
}

func getAPIDefAuthDisabled() AWSAPIDefinition {
	return AWSAPIDefinition{
		Name:                         "api3",
		Context:                      "api3",
		IdentitySource:               "foo",
		AuthorizerType:               "REQUEST",
		AuthorizerAuthType:           "foo",
		AuthorizerName:               "foo",
		IdentityValidationExpression: "",
		AuthorizerResultTtlInSeconds: 3600,
		AuthorizerUri:                "arn:bar",
		ProviderARNs: []string{
			"arn:foo",
		},
		AuthenticationEnabled: false,
		APIKeyEnabled:         false,
		Authorization_Enabled: true,
	}
}

func getAPIDefAuthorizationDisabled() AWSAPIDefinition {
	return AWSAPIDefinition{
		Name:                         "api4",
		Context:                      "api4",
		IdentitySource:               "foo",
		AuthorizerType:               "REQUEST",
		AuthorizerAuthType:           "foo",
		AuthorizerName:               "foo",
		IdentityValidationExpression: "",
		AuthorizerResultTtlInSeconds: 3600,
		AuthorizerUri:                "arn:bar",
		ProviderARNs: []string{
			"arn:foo",
		},
		AuthenticationEnabled: false,
		APIKeyEnabled:         false,
		Authorization_Enabled: false,
	}
}

func getAPIDefsWOUsagePlans() []AWSAPIDefinition {
	return []AWSAPIDefinition{
		getAPIDefWOUsagePlans(),
		getAPIDefAPIKeyDisabled(),
		getAPIDefAuthDisabled(),
		getAPIDefAuthorizationDisabled(),
	}
}

func getAWSAPIDefWOUsagePlansBytes() string {
	awsDefs := getAPIDefsWOUsagePlans()
	awsDefsBytes, _ := json.Marshal(awsDefs)
	return string(awsDefsBytes)
}

func getAPIResource() APIResource {
	return APIResource{
		Path: "/api/v1/foobar",
		Methods: []string{
			"GET",
			"POST",
		},
		CachingEnabled: false,
		ProxyPathParams: []Param{
			{
				Param:    "fooid",
				Required: true,
			},
		},
		ProxyQueryParams: []Param{
			{
				Param:    "fooid",
				Required: true,
			},
		},
		ProxyHeaderParams: []Param{
			{
				Param:    "fooid",
				Required: true,
			},
		},
	}
}

func getAPIResourcesBytes() string {
	resourcesBytes, _ := json.Marshal(getAPIResources())
	return string(resourcesBytes)
}

func TestBuildApiGatewayTemplateFromIngressRule(t *testing.T) {
	tests := []struct {
		name string
		args *TemplateConfig
		want *cfn.Template
	}{
		{
			name: "generates template without custom domain",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":          buildLambdaExecutionRole(),
					"Methodapi0":                buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":              buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":        buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":               buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                  buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":     buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":              buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                   buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":          Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0": Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":          Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":   Output{Value: "EDGE"},
					"RequestTimeout":      Output{Value: "10000"},
				},
			},
		},
		{
			name: "generates template with content encoding",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 1000000000,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":          buildLambdaExecutionRole(),
					"Methodapi0":                buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":              buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":        buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":               buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                  buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":     buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 1000000000, cfn.Ref("AWS::StackName")),
					"Deployment0":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":              buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                   buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":             Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0":    Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":             Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":      Output{Value: "EDGE"},
					"RequestTimeout":         Output{Value: "10000"},
					"MinimumCompressionSize": Output{Value: "1000000000"},
				},
			},
		},
		{
			name: "generates template with content encoding api keys",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				StageName:              "baz",
				NodePort:               30123,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				UsagePlans:             getUsagePlans(),
				MinimumCompressionSize: 1000000000,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":          buildLambdaExecutionRole(),
					"Methodapi0":                buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Methodapiv10":              buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":        buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Resourceapi0":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":               buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                  buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":     buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "NONE", 1000000000, cfn.Ref("AWS::StackName")),
					"Deployment0":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":              buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                   buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"APIKeyUsagePlan000":        getAPIKeyMappingBuild(0, 0, 0),
					"APIKeyUsagePlan010":        getAPIKeyMappingBuild(1, 0, 0),
					"UsagePlan0":                buildUsagePlan(getUsagePlan(), "baz", 0),
					"APIKey000":                 getAPIKeyBuild(0),
					"APIKey010":                 getAPIKeyBuild(1),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":             Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0":    Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"APIGWEndpointType":      Output{Value: "EDGE"},
					"RequestTimeout":         Output{Value: "10000"},
					"MinimumCompressionSize": Output{Value: "1000000000"},
					"UsagePlansData":         Output{Value: getUsagePlanBytes()},
				},
			},
		},
		{
			name: "generates template with usage plan",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				StageName:              "baz",
				NodePort:               30123,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				UsagePlans:             getUsagePlans(),
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":          buildLambdaExecutionRole(),
					"Methodapi0":                buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Methodapiv10":              buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":        buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Resourceapi0":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":               buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                  buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":     buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "NONE", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":              buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                   buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"APIKeyUsagePlan000":        getAPIKeyMappingBuild(0, 0, 0),
					"APIKeyUsagePlan010":        getAPIKeyMappingBuild(1, 0, 0),
					"UsagePlan0":                buildUsagePlan(getUsagePlan(), "baz", 0),
					"APIKey000":                 getAPIKeyBuild(0),
					"APIKey010":                 getAPIKeyBuild(1),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":          Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0": Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"APIGWEndpointType":   Output{Value: "EDGE"},
					"RequestTimeout":      Output{Value: "10000"},
					"UsagePlansData":      Output{Value: getUsagePlanBytes()},
				},
			},
		},
		{
			name: "generates template with waf",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				WAFEnabled:             true,
				WAFRulesJSON:           "[]",
				WAFAssociation:         true,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":          buildLambdaExecutionRole(),
					"Methodapi0":                buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":              buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":        buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":               buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                  buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":     buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":              buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                   buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"WAFAcl":                    buildAWSWAFWebACL("REGIONAL", "[]"),
					"WAFAssociation0":           buildAWSWAFWebACLAssociation("baz", 0),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":          Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0": Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":          Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":   Output{Value: "EDGE"},
					"WAFEnabled":          Output{Value: "true"},
					"WAFRules":            Output{Value: "[]"},
					"WAFScope":            Output{Value: "REGIONAL"},
					"WAFAssociation0":     Output{Value: cfn.Ref("WAFAssociation0")},
					"RequestTimeout":      Output{Value: "10000"},
				},
			},
		},
		{
			name: "generates template with waf regional api",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				WAFEnabled:             true,
				WAFRulesJSON:           "[]",
				WAFAssociation:         true,
				APIEndpointType:        "REGIONAL",
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":          buildLambdaExecutionRole(),
					"Methodapi0":                buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":              buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":        buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":               buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                  buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":     buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "REGIONAL", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":              buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                   buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"WAFAcl":                    buildAWSWAFWebACL("REGIONAL", "[]"),
					"WAFAssociation0":           buildAWSWAFWebACLAssociation("baz", 0),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":          Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0": Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":          Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":   Output{Value: "REGIONAL"},
					"WAFEnabled":          Output{Value: "true"},
					"WAFRules":            Output{Value: "[]"},
					"WAFScope":            Output{Value: "REGIONAL"},
					"WAFAssociation0":     Output{Value: cfn.Ref("WAFAssociation0")},
					"RequestTimeout":      Output{Value: "10000"},
				},
			},
		},
		{
			name: "generates template with waf null rules",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				WAFAssociation:         true,
				WAFEnabled:             true,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":          buildLambdaExecutionRole(),
					"Methodapi0":                buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":              buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":        buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":               buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                  buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":     buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":              buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                   buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"WAFAcl":                    buildAWSWAFWebACL("REGIONAL", ""),
					"WAFAssociation0":           buildAWSWAFWebACLAssociation("baz", 0),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":          Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0": Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":          Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":   Output{Value: "EDGE"},
					"WAFEnabled":          Output{Value: "true"},
					"WAFRules":            Output{Value: ""},
					"WAFScope":            Output{Value: "REGIONAL"},
					"WAFAssociation0":     Output{Value: cfn.Ref("WAFAssociation0")},
					"RequestTimeout":      Output{Value: "10000"},
				},
			},
		},
		{
			name: "generates template with waf error rules",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				WAFEnabled:             true,
				WAFRulesJSON:           "wrongjson",
				WAFAssociation:         true,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":          buildLambdaExecutionRole(),
					"Methodapi0":                buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":              buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":        buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":               buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                  buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":     buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":              buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                   buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"WAFAcl":                    buildAWSWAFWebACL("REGIONAL", ""),
					"WAFAssociation0":           buildAWSWAFWebACLAssociation("baz", 0),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":          Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0": Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":          Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":   Output{Value: "EDGE"},
					"WAFEnabled":          Output{Value: "true"},
					"WAFRules":            Output{Value: "wrongjson"},
					"WAFScope":            Output{Value: "REGIONAL"},
					"WAFAssociation0":     Output{Value: cfn.Ref("WAFAssociation0")},
					"RequestTimeout":      Output{Value: "10000"},
				},
			},
		},
		{
			name: "generates template with custom domain",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				CustomDomainName:       "example.com",
				CertificateArn:         "arn::foobar",
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":             buildLambdaExecutionRole(),
					"Methodapi0":                   buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":                 buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":           buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":      buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":                 buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":               buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":         buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0":    buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":                  buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                     buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":        buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                     buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":                  buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":                 buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                      buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"CustomDomain":                 buildCustomDomain("example.com", "arn::foobar", "EDGE", "TLS_1_2"),
					"CustomDomainBasePathMapping0": buildCustomDomainBasePathMapping("example.com", "baz", "", 0),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":               Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0":      Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":               Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":        Output{Value: "EDGE"},
					"SSLCertArn":               Output{Value: "arn::foobar"},
					"CustomDomainName":         Output{Value: "example.com"},
					"CustomDomainHostname":     Output{Value: cfn.GetAtt("CustomDomain", "DistributionDomainName")},
					"CustomDomainHostedZoneID": Output{Value: cfn.GetAtt("CustomDomain", "DistributionHostedZoneId")},
					"RequestTimeout":           Output{Value: "10000"},
					"TLSPolicy":                Output{Value: "TLS_1_2"},
					"CustomDomainBasePath":     Output{Value: ""},
				},
			},
		},
		{
			name: "generates template with custom domain with base path",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				CustomDomainName:       "example.com",
				CustomDomainBasePath:   "foo",
				CertificateArn:         "arn::foobar",
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":             buildLambdaExecutionRole(),
					"Methodapi0":                   buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":                 buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":           buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":      buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":                 buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":               buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":         buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0":    buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":                  buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                     buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":        buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                     buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":                  buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":                 buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                      buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"CustomDomain":                 buildCustomDomain("example.com", "arn::foobar", "EDGE", "TLS_1_2"),
					"CustomDomainBasePathMapping0": buildCustomDomainBasePathMapping("example.com", "baz", "foo", 0),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":               Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0":      Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":               Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":        Output{Value: "EDGE"},
					"SSLCertArn":               Output{Value: "arn::foobar"},
					"CustomDomainName":         Output{Value: "example.com"},
					"CustomDomainHostname":     Output{Value: cfn.GetAtt("CustomDomain", "DistributionDomainName")},
					"CustomDomainHostedZoneID": Output{Value: cfn.GetAtt("CustomDomain", "DistributionHostedZoneId")},
					"RequestTimeout":           Output{Value: "10000"},
					"TLSPolicy":                Output{Value: "TLS_1_2"},
					"CustomDomainBasePath":     Output{Value: "foo"},
				},
			},
		},
		{
			name: "generates template with custom domain regional api",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				CustomDomainName:       "example.com",
				CertificateArn:         "arn::foobar",
				APIEndpointType:        "REGIONAL",
				TLSPolicy:              "TLS_1_2",
				RequestTimeout:         10000,
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":             buildLambdaExecutionRole(),
					"Methodapi0":                   buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":                 buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":           buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":      buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":                 buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":               buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":         buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0":    buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":                  buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                     buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":        buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                     buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "REGIONAL", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":                  buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":                 buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                      buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"CustomDomain":                 buildCustomDomain("example.com", "arn::foobar", "REGIONAL", "TLS_1_2"),
					"CustomDomainBasePathMapping0": buildCustomDomainBasePathMapping("example.com", "baz", "", 0),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":               Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0":      Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":               Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":        Output{Value: "REGIONAL"},
					"SSLCertArn":               Output{Value: "arn::foobar"},
					"CustomDomainName":         Output{Value: "example.com"},
					"CustomDomainHostname":     Output{Value: cfn.GetAtt("CustomDomain", "RegionalDomainName")},
					"CustomDomainHostedZoneID": Output{Value: cfn.GetAtt("CustomDomain", "RegionalHostedZoneId")},
					"RequestTimeout":           Output{Value: "10000"},
					"TLSPolicy":                Output{Value: "TLS_1_2"},
					"CustomDomainBasePath":     Output{Value: ""},
				},
			},
		},
		{
			name: "generates template with custom domain edge api with WAF",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				CustomDomainName:       "example.com",
				CertificateArn:         "arn::foobar",
				WAFEnabled:             true,
				WAFRulesJSON:           "[]",
				WAFAssociation:         true,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":             buildLambdaExecutionRole(),
					"Methodapi0":                   buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":                 buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":           buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":      buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":                 buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":               buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":         buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0":    buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":                  buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                     buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":        buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                     buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":                  buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":                 buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                      buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"CustomDomain":                 buildCustomDomain("example.com", "arn::foobar", "EDGE", "TLS_1_2"),
					"CustomDomainBasePathMapping0": buildCustomDomainBasePathMapping("example.com", "baz", "", 0),
					"WAFAcl":                       buildAWSWAFWebACL("REGIONAL", "[]"),
					"WAFAssociation0":              buildAWSWAFWebACLAssociation("baz", 0),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":               Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0":      Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":               Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":        Output{Value: "EDGE"},
					"SSLCertArn":               Output{Value: "arn::foobar"},
					"CustomDomainName":         Output{Value: "example.com"},
					"CustomDomainHostname":     Output{Value: cfn.GetAtt("CustomDomain", "DistributionDomainName")},
					"CustomDomainHostedZoneID": Output{Value: cfn.GetAtt("CustomDomain", "DistributionHostedZoneId")},
					"WAFEnabled":               Output{Value: "true"},
					"WAFRules":                 Output{Value: "[]"},
					"WAFScope":                 Output{Value: "REGIONAL"},
					"WAFAssociation0":          Output{Value: cfn.Ref("WAFAssociation0")},
					"RequestTimeout":           Output{Value: "10000"},
					"TLSPolicy":                Output{Value: "TLS_1_2"},
					"CustomDomainBasePath":     Output{Value: ""},
				},
			},
		},
		{
			name: "generates template with custom domain edge api with WAF without association",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				CustomDomainName:       "example.com",
				CertificateArn:         "arn::foobar",
				WAFEnabled:             true,
				WAFRulesJSON:           "[]",
				WAFAssociation:         false,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":             buildLambdaExecutionRole(),
					"Methodapi0":                   buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":                 buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":           buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":      buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":                 buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":               buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":         buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0":    buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":                  buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                     buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":        buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                     buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":                  buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":                 buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                      buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"CustomDomain":                 buildCustomDomain("example.com", "arn::foobar", "EDGE", "TLS_1_2"),
					"CustomDomainBasePathMapping0": buildCustomDomainBasePathMapping("example.com", "baz", "", 0),
					"WAFAcl":                       buildAWSWAFWebACL("REGIONAL", "[]"),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":               Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0":      Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":               Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":        Output{Value: "EDGE"},
					"SSLCertArn":               Output{Value: "arn::foobar"},
					"CustomDomainName":         Output{Value: "example.com"},
					"CustomDomainHostname":     Output{Value: cfn.GetAtt("CustomDomain", "DistributionDomainName")},
					"CustomDomainHostedZoneID": Output{Value: cfn.GetAtt("CustomDomain", "DistributionHostedZoneId")},
					"WAFEnabled":               Output{Value: "true"},
					"WAFRules":                 Output{Value: "[]"},
					"WAFScope":                 Output{Value: "REGIONAL"},
					"WAFAssociation0":          Output{Value: cfn.Ref("WAFAssociation0")},
					"RequestTimeout":           Output{Value: "10000"},
					"TLSPolicy":                Output{Value: "TLS_1_2"},
					"CustomDomainBasePath":     Output{Value: ""},
				},
			},
		},
		{
			name: "generates template with custom domain regional api with WAF",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				CustomDomainName:       "example.com",
				CertificateArn:         "arn::foobar",
				APIEndpointType:        "REGIONAL",
				WAFEnabled:             true,
				WAFRulesJSON:           "[]",
				WAFAssociation:         true,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":             buildLambdaExecutionRole(),
					"Methodapi0":                   buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":                 buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":           buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":      buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":                 buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":               buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":         buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0":    buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":                  buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                     buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":        buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                     buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "REGIONAL", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":                  buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":                 buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                      buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"CustomDomain":                 buildCustomDomain("example.com", "arn::foobar", "REGIONAL", "TLS_1_2"),
					"CustomDomainBasePathMapping0": buildCustomDomainBasePathMapping("example.com", "baz", "", 0),
					"WAFAcl":                       buildAWSWAFWebACL("REGIONAL", "[]"),
					"WAFAssociation0":              buildAWSWAFWebACLAssociation("baz", 0),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":               Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0":      Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":               Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":        Output{Value: "REGIONAL"},
					"SSLCertArn":               Output{Value: "arn::foobar"},
					"CustomDomainName":         Output{Value: "example.com"},
					"CustomDomainHostname":     Output{Value: cfn.GetAtt("CustomDomain", "RegionalDomainName")},
					"CustomDomainHostedZoneID": Output{Value: cfn.GetAtt("CustomDomain", "RegionalHostedZoneId")},
					"WAFEnabled":               Output{Value: "true"},
					"WAFRules":                 Output{Value: "[]"},
					"WAFScope":                 Output{Value: "REGIONAL"},
					"WAFAssociation0":          Output{Value: cfn.Ref("WAFAssociation0")},
					"RequestTimeout":           Output{Value: "10000"},
					"TLSPolicy":                Output{Value: "TLS_1_2"},
					"CustomDomainBasePath":     Output{Value: ""},
				},
			},
		},
		{
			name: "generates template with defined public apis",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
				APIResources:           getAPIResources(),
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":       buildLambdaExecutionRole(),
					"Methodapiv1foobarGET0":  buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar"}), 10000, "AWS_IAM", "GET", getAPIResource(), 0),
					"Methodapiv1foobarPOST0": buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar"}), 10000, "AWS_IAM", "POST", getAPIResource(), 0),
					"Resourceapi0":           buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":         buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":   buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"TargetGroup":            buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":               buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":  buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":               buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":            buildAWSApiGatewayDeployment("baz", []string{"Methodapiv1foobarGET0", "Methodapiv1foobarPOST0"}, false, getAPIResources(), "", 0),
					"LoadBalancer":           buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":          Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0": Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":          Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":   Output{Value: "EDGE"},
					"RequestTimeout":      Output{Value: "10000"},
					"APIResources":        Output{Value: getAPIResourcesBytes()},
				},
			},
		},
		{
			name: "generates template with defined public apis with cache",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
				APIResources:           getAPIResources(),
				CachingEnabled:         true,
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":       buildLambdaExecutionRole(),
					"Methodapiv1foobarGET0":  buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar"}), 10000, "AWS_IAM", "GET", getAPIResource(), 0),
					"Methodapiv1foobarPOST0": buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar"}), 10000, "AWS_IAM", "POST", getAPIResource(), 0),
					"Resourceapi0":           buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":         buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":   buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"TargetGroup":            buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":               buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":  buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":               buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":            buildAWSApiGatewayDeployment("baz", []string{"Methodapiv1foobarGET0", "Methodapiv1foobarPOST0"}, true, getAPIResources(), "0.5", 0),
					"LoadBalancer":           buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":          Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0": Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":          Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":   Output{Value: "EDGE"},
					"RequestTimeout":      Output{Value: "10000"},
					"CachingEnabled":      Output{Value: "true"},
					"CachingSize":         Output{Value: "0.5"},
					"APIResources":        Output{Value: getAPIResourcesBytes()},
				},
			},
		},
		{
			name: "generates template with defined public apis with cache size only",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
				APIResources:           getAPIResources(),
				CachingSize:            "0.5",
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":       buildLambdaExecutionRole(),
					"Methodapiv1foobarGET0":  buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar"}), 10000, "AWS_IAM", "GET", getAPIResource(), 0),
					"Methodapiv1foobarPOST0": buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar"}), 10000, "AWS_IAM", "POST", getAPIResource(), 0),
					"Resourceapi0":           buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":         buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":   buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"TargetGroup":            buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":               buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":  buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":               buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, cfn.Ref("AWS::StackName")),
					"Deployment0":            buildAWSApiGatewayDeployment("baz", []string{"Methodapiv1foobarGET0", "Methodapiv1foobarPOST0"}, true, getAPIResources(), "0.5", 0),
					"LoadBalancer":           buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":          Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0": Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":          Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":   Output{Value: "EDGE"},
					"RequestTimeout":      Output{Value: "10000"},
					"CachingEnabled":      Output{Value: "true"},
					"CachingSize":         Output{Value: "0.5"},
					"APIResources":        Output{Value: getAPIResourcesBytes()},
				},
			},
		},
		{
			name: "generates template API Defs with Usage plans and auth enabled",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				Arns:                   []string{"arn::foo"},
				StageName:              "baz",
				NodePort:               30123,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				MinimumCompressionSize: 0,
				AWSAPIDefinitions:      getAPIDefs(),
				CustomDomainName:       "example.com",
				CertificateArn:         "arn::foobar",
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":             buildLambdaExecutionRole(),
					"Methodapi0":                   buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv10":                 buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":           buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":      buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "AWS_IAM", "ANY", APIResource{}, 0),
					"Resourceapi0":                 buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapiv10":               buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":         buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0":    buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"TargetGroup":                  buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                     buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":        buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                     buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "AWS_IAM", 0, "api0"),
					"Deployment0":                  buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"LoadBalancer":                 buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                      buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"APIKeyUsagePlan000":           getSecondAPIKeyMappingBuild(0, 0, 0),
					"APIKeyUsagePlan010":           getSecondAPIKeyMappingBuild(1, 0, 0),
					"UsagePlan0":                   buildUsagePlan(getSecondUsagePlan(), "baz", 0),
					"APIKey000":                    getSecondAPIKeyBuild(0),
					"APIKey010":                    getSecondAPIKeyBuild(1),
					"RestAPIAuthorizer0":           buildAuthorizer(getAPIDef(), 0),
					"CustomDomain":                 buildCustomDomain("example.com", "arn::foobar", "EDGE", "TLS_1_2"),
					"CustomDomainBasePathMapping0": buildCustomDomainBasePathMapping("example.com", "baz", "api0", 0),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":               Output{Value: cfn.Ref("RestAPI0")},
					"APIGatewayEndpoint0":      Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"ClientARNS":               Output{Value: strings.Join([]string{"arn::foo"}, ",")},
					"APIGWEndpointType":        Output{Value: "EDGE"},
					"RequestTimeout":           Output{Value: "10000"},
					"AWSAPIConfigs":            Output{Value: getAWSAPIDefBytes()},
					"SSLCertArn":               Output{Value: "arn::foobar"},
					"CustomDomainName":         Output{Value: "example.com"},
					"CustomDomainHostname":     Output{Value: cfn.GetAtt("CustomDomain", "DistributionDomainName")},
					"CustomDomainHostedZoneID": Output{Value: cfn.GetAtt("CustomDomain", "DistributionHostedZoneId")},
					"TLSPolicy":                Output{Value: "TLS_1_2"},
					"CustomDomainBasePath":     Output{Value: ""},
				},
			},
		},
		{
			name: "generates template API Defs without Usage plans and with auth enabled",
			args: &TemplateConfig{
				Rule: extensionsv1beta1.IngressRule{
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/api/v1/foobar",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: "foobar-service",
										ServicePort: intstr.FromInt(8080),
									},
								},
							},
						},
					},
				},
				Network: &network.Network{
					Vpc: &ec2.Vpc{
						VpcId:     aws.String("foo"),
						CidrBlock: aws.String("10.0.0.0/24"),
					},
					InstanceIDs:      []string{"i-foo"},
					SubnetIDs:        []string{"sn-foo"},
					SecurityGroupIDs: []string{"sg-foo"},
				},
				StageName:              "baz",
				NodePort:               30123,
				RequestTimeout:         10000,
				TLSPolicy:              "TLS_1_2",
				UsagePlans:             getUsagePlans(),
				MinimumCompressionSize: 1000000000,
				AWSAPIDefinitions:      getAPIDefsWOUsagePlans(),
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole":          buildLambdaExecutionRole(),
					"Methodapi0":                buildAWSApiGatewayMethod("Resourceapi0", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Methodapiv10":              buildAWSApiGatewayMethod("Resourceapiv10", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Methodapiv1foobar0":        buildAWSApiGatewayMethod("Resourceapiv1foobar0", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Methodapiv1foobarproxy0":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy0", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 0),
					"Methodapi1":                buildAWSApiGatewayMethod("Resourceapi1", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 1),
					"Methodapiv11":              buildAWSApiGatewayMethod("Resourceapiv11", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 1),
					"Methodapiv1foobar1":        buildAWSApiGatewayMethod("Resourceapiv1foobar1", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 1),
					"Methodapiv1foobarproxy1":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy1", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 1),
					"Methodapi2":                buildAWSApiGatewayMethod("Resourceapi2", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 2),
					"Methodapiv12":              buildAWSApiGatewayMethod("Resourceapiv12", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 2),
					"Methodapiv1foobar2":        buildAWSApiGatewayMethod("Resourceapiv1foobar2", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 2),
					"Methodapiv1foobarproxy2":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy2", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 2),
					"Methodapi3":                buildAWSApiGatewayMethod("Resourceapi3", toPath(1, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 3),
					"Methodapiv13":              buildAWSApiGatewayMethod("Resourceapiv13", toPath(2, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 3),
					"Methodapiv1foobar3":        buildAWSApiGatewayMethod("Resourceapiv1foobar3", toPath(3, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 3),
					"Methodapiv1foobarproxy3":   buildAWSApiGatewayMethod("Resourceapiv1foobarproxy3", toPath(4, []string{"", "api", "v1", "foobar", "{proxy+}"}), 10000, "NONE", "ANY", APIResource{}, 3),
					"Resourceapi0":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI0", "RootResourceId0"), "api", 0),
					"Resourceapi1":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI1", "RootResourceId1"), "api", 1),
					"Resourceapi2":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI2", "RootResourceId2"), "api", 2),
					"Resourceapi3":              buildAWSApiGatewayResource(cfn.GetAtt("RestAPI3", "RootResourceId3"), "api", 3),
					"Resourceapiv10":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi0"), "v1", 0),
					"Resourceapiv1foobar0":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv10"), "foobar", 0),
					"Resourceapiv1foobarproxy0": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar0"), "{proxy+}", 0),
					"Resourceapiv11":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi1"), "v1", 1),
					"Resourceapiv1foobar1":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv11"), "foobar", 1),
					"Resourceapiv1foobarproxy1": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar1"), "{proxy+}", 1),
					"Resourceapiv12":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi2"), "v1", 2),
					"Resourceapiv1foobar2":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv12"), "foobar", 2),
					"Resourceapiv1foobarproxy2": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar2"), "{proxy+}", 2),
					"Resourceapiv13":            buildAWSApiGatewayResource(cfn.Ref("Resourceapi3"), "v1", 3),
					"Resourceapiv1foobar3":      buildAWSApiGatewayResource(cfn.Ref("Resourceapiv13"), "foobar", 3),
					"Resourceapiv1foobarproxy3": buildAWSApiGatewayResource(cfn.Ref("Resourceapiv1foobar3"), "{proxy+}", 3),
					"TargetGroup":               buildAWSElasticLoadBalancingV2TargetGroup("foo", []string{"i-foo"}, 30123, []string{"LoadBalancer"}),
					"Listener":                  buildAWSElasticLoadBalancingV2Listener(),
					"SecurityGroupIngress0":     buildAWSEC2SecurityGroupIngresses([]string{"sg-foo"}, "10.0.0.0/24", 30123)[0],
					"RestAPI0":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "NONE", 1000000000, "api1"),
					"RestAPI1":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "NONE", 1000000000, "api2"),
					"RestAPI2":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "NONE", 1000000000, "api3"),
					"RestAPI3":                  buildAWSApiGatewayRestAPI([]string{"arn::foo"}, "EDGE", "NONE", 1000000000, "api4"),
					"Deployment0":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi0", "Methodapiv10", "Methodapiv1foobar0", "Methodapiv1foobarproxy0"}, false, nil, "", 0),
					"Deployment1":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi1", "Methodapiv11", "Methodapiv1foobar1", "Methodapiv1foobarproxy1"}, false, nil, "", 1),
					"Deployment2":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi2", "Methodapiv12", "Methodapiv1foobar2", "Methodapiv1foobarproxy2"}, false, nil, "", 2),
					"Deployment3":               buildAWSApiGatewayDeployment("baz", []string{"Methodapi3", "Methodapiv13", "Methodapiv1foobar3", "Methodapiv1foobarproxy3"}, false, nil, "", 3),
					"LoadBalancer":              buildAWSElasticLoadBalancingV2LoadBalancer([]string{"sn-foo"}),
					"VPCLink":                   buildAWSApiGatewayVpcLink([]string{"LoadBalancer"}),
					"APIKeyUsagePlan000":        getAPIKeyMappingBuild(0, 0, 0),
					"APIKeyUsagePlan010":        getAPIKeyMappingBuild(1, 0, 0),
					"UsagePlan0":                buildUsagePlan(getUsagePlan(), "baz", 0),
					"APIKey000":                 getAPIKeyBuild(0),
					"APIKey010":                 getAPIKeyBuild(1),
					"RestAPIAuthorizer0":        buildAuthorizer(getAPIDefWOUsagePlans(), 0),
					"RestAPIAuthorizer1":        buildAuthorizer(getAPIDefAPIKeyDisabled(), 1),
					"RestAPIAuthorizer2":        buildAuthorizer(getAPIDefAuthDisabled(), 2),
				},
				Outputs: map[string]interface{}{
					"RestAPIID0":             Output{Value: cfn.Ref("RestAPI0")},
					"RestAPIID1":             Output{Value: cfn.Ref("RestAPI1")},
					"RestAPIID2":             Output{Value: cfn.Ref("RestAPI2")},
					"RestAPIID3":             Output{Value: cfn.Ref("RestAPI3")},
					"APIGatewayEndpoint0":    Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI0"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"APIGatewayEndpoint1":    Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI1"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"APIGatewayEndpoint2":    Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI2"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"APIGatewayEndpoint3":    Output{Value: cfn.Join("", []string{"https://", cfn.Ref("RestAPI3"), ".execute-api.", cfn.Ref("AWS::Region"), ".amazonaws.com/", "baz"})},
					"APIGWEndpointType":      Output{Value: "EDGE"},
					"RequestTimeout":         Output{Value: "10000"},
					"MinimumCompressionSize": Output{Value: "1000000000"},
					"UsagePlansData":         Output{Value: getUsagePlanBytes()},
					"AWSAPIConfigs":          Output{Value: getAWSAPIDefWOUsagePlansBytes()},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildAPIGatewayTemplateFromIngressRule(tt.args)
			for k, resource := range got.Resources {
				if !reflect.DeepEqual(resource, tt.want.Resources[k]) {
					t.Errorf("Got Resources.%s = %v, want %v", k, got.Resources, tt.want.Resources)
				}
			}
			for k, resource := range got.Outputs {
				if !reflect.DeepEqual(resource, tt.want.Outputs[k]) {
					t.Errorf("Got Outputs.%s = %v, want %v", k, got.Outputs, tt.want.Outputs)
				}
			}
		})
	}
}

func TestBuildApiGatewayTemplateForRoute53(t *testing.T) {
	tests := []struct {
		name string
		args *Route53TemplateConfig
		want *cfn.Template
	}{
		{
			name: "generates template for edge hosted zone",
			args: &Route53TemplateConfig{
				CustomDomainName:         "example.com",
				CustomDomainHostName:     "d-example.aws.com",
				CustomDomainHostedZoneID: "123234",
				HostedZoneName:           "example.com",
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole": buildLambdaExecutionRole(),
					"Route53RecordSet": buildCustomDomainRoute53Record("example.com", "example.com", "d-example.aws.com", "123234"),
				},
				Outputs: map[string]interface{}{
					"CustomDomainHostname":     Output{Value: "d-example.aws.com"},
					"CustomDomainHostedZoneID": Output{Value: "123234"},
					"CustomDomainName":         Output{Value: "example.com"},
					"HostedZone":               Output{Value: "example.com"},
				},
			},
		},
		{
			name: "generates template for regional hosted zone",
			args: &Route53TemplateConfig{
				CustomDomainName:         "example.com",
				CustomDomainHostName:     "d-example.aws.com",
				CustomDomainHostedZoneID: "123234",
				HostedZoneName:           "example.com",
			},
			want: &cfn.Template{
				Resources: cfn.Resources{
					"LambdaInvokeRole": buildLambdaExecutionRole(),
					"Route53RecordSet": buildCustomDomainRoute53Record("example.com", "example.com", "d-example.aws.com", "123234"),
				},
				Outputs: map[string]interface{}{
					"CustomDomainHostname":     Output{Value: "d-example.aws.com"},
					"CustomDomainHostedZoneID": Output{Value: "123234"},
					"CustomDomainName":         Output{Value: "example.com"},
					"HostedZone":               Output{Value: "example.com"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildAPIGatewayRoute53Template(tt.args)
			for k, resource := range got.Resources {
				if !reflect.DeepEqual(resource, tt.want.Resources[k]) {
					t.Errorf("Got Resources.%s = %v, want %v", k, got.Resources, tt.want.Resources)
				}
			}
			for k, resource := range got.Outputs {
				if !reflect.DeepEqual(resource, tt.want.Outputs[k]) {
					t.Errorf("Got Outputs.%s = %v, want %v", k, got.Outputs, tt.want.Outputs)
				}
			}
		})
	}
}
