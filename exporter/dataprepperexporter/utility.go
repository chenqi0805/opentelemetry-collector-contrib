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
	if (len(names) != 3) || (names[0] != "es") || (names[1] != "dataprepper") || (names[2] == "") {
		return "", fmt.Errorf("invalid resource format in pipeline_arn: %s", resource)
	}
	header := names[2]
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
