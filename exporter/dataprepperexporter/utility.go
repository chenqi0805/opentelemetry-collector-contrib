package dataprepperexporter

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
)

func getDataPrepperHeader(pipelineArn string) (string, error) {
	pipelineArnObj, err := arn.Parse(pipelineArn)
	if err != nil {
		return "", fmt.Errorf("failed to parse pipeline_arn: %s", pipelineArn)
	}
	var resource = pipelineArnObj.Resource
	names := strings.Split(resource, "/")
	if (len(names) != 2) || (names[0] != "pipeline") || (names[1] == "") {
		return "", fmt.Errorf("invalid resource format in pipeline_arn: %s", resource)
	}
	header := names[1]
	return header, nil
}

func getHostHeader(pipelineArn string) (string, error) {
	pipelineArnObj, err := arn.Parse(pipelineArn)
	if err != nil {
		return "", fmt.Errorf("failed to parse pipeline_arn: %s", pipelineArn)
	}
	header := fmt.Sprintf("%s.ingest.%s.amazonaws.com", pipelineArnObj.AccountID, pipelineArnObj.Region)
	return header, nil
}
