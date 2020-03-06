/*
Copyright (c) 2020 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	sdk "github.com/openshift-online/ocm-sdk-go"

	"gitlab.cee.redhat.com/service/moactl/pkg/tags"
)

type Client interface {
	GetRegion() string
	ValidateCredentials() (bool, error)
	CreateUser(username string, clusterName string) error
	CreateAccessKey(username string) (*AWSAccessKey, error)
	GetCreator() (*AWSCreator, error)
	TagUser(username string, clusterID string) error
	ValidateSCP() (bool, error)
}

// ClientBuilder contains the information and logic needed to build a new AWS client.
type ClientBuilder struct {
	logger sdk.Logger
}

type awsClient struct {
	logger sdk.Logger
	// cloudformationClient cloudformationiface.CloudFormationAPI
	iamClient  iamiface.IAMAPI
	orgClient  organizationsiface.OrganizationsAPI
	stsClient  stsiface.STSAPI
	awsSession *session.Session
}

// NewClient creates a builder that can then be used to configure and build a new AWS client.
func NewClient() *ClientBuilder {
	return &ClientBuilder{}
}

// Logger sets the logger that the AWS client will use to send messages to the log.
func (b *ClientBuilder) Logger(value sdk.Logger) *ClientBuilder {
	b.logger = value
	return b
}

// Build uses the information stored in the builder to build a new AWS client.
func (b *ClientBuilder) Build() (result Client, err error) {
	// Create the AWS session:
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return
	}

	// Check that the region is set:
	region := aws.StringValue(sess.Config.Region)
	if region == "" {
		err = fmt.Errorf("region is not set")
		return
	}

	// Check that the AWS credentials are available:
	_, err = sess.Config.Credentials.Get()
	if err != nil {
		err = fmt.Errorf("can't find credentials: %v", err)
		return
	}

	// Create and populate the object:
	result = &awsClient{
		logger:     b.logger,
		iamClient:  iam.New(sess),
		orgClient:  organizations.New(sess),
		stsClient:  sts.New(sess),
		awsSession: sess,
	}

	return
}

func (c *awsClient) GetRegion() string {
	return aws.StringValue(c.awsSession.Config.Region)
}

type AWSCreator struct {
	ARN       string
	AccountID string
}

func (c *awsClient) GetCreator() (*AWSCreator, error) {
	getCallerIdentityOutput, err := c.stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	creatorARN := aws.StringValue(getCallerIdentityOutput.Arn)

	// Extract the account identifier from the ARN of the user:
	creatorParsedARN, err := arn.Parse(creatorARN)
	if err != nil {
		return nil, err
	}
	return &AWSCreator{
		ARN:       creatorARN,
		AccountID: creatorParsedARN.AccountID,
	}, nil
}

// Checks if given credentials are valid.
func (c *awsClient) ValidateCredentials() (bool, error) {
	_, err := c.stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return false, err
	}
	return true, nil
}

// Ensure osdCcsAdmin account
func (c *awsClient) CreateUser(username string, clusterName string) error {
	user, err := c.iamClient.CreateUser(&iam.CreateUserInput{
		UserName: aws.String(username),
		Tags: []*iam.Tag{
			{
				Key:   aws.String(tags.ClusterName),
				Value: aws.String(clusterName),
			},
		},
	})
	if err != nil {
		switch typed := err.(type) {
		case awserr.Error:
			if typed.Code() == iam.ErrCodeEntityAlreadyExistsException {
				return errors.New(fmt.Sprintf(
					"User '%s' already exists, which means that there is already a cluster created in the account",
					username,
				))
			}
		}
		return err
	}
	c.logger.Debug(nil, "CreateUser::CreateUser\n%+v\n", user)

	policy, err := c.iamClient.AttachUserPolicy(&iam.AttachUserPolicyInput{
		PolicyArn: aws.String("arn:aws:iam::aws:policy/AdministratorAccess"),
		UserName:  aws.String(username),
	})
	if err != nil {
		return err
	}
	c.logger.Debug(nil, "CreateUser::AttachUserPolicy\n%+v\n", policy)

	return nil
}

func (c *awsClient) TagUser(username string, clusterID string) error {
	_, err := c.iamClient.TagUser(&iam.TagUserInput{
		UserName: aws.String(username),
		Tags: []*iam.Tag{
			{
				Key:   aws.String(tags.ClusterID),
				Value: aws.String(clusterID),
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

type AWSAccessKey struct {
	AccessKeyID     string
	SecretAccessKey string
}

func (c *awsClient) CreateAccessKey(username string) (*AWSAccessKey, error) {
	createAccessKeyOutput, err := c.iamClient.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(username),
	})
	if err != nil {
		return nil, err
	}
	accessKey := createAccessKeyOutput.AccessKey
	return &AWSAccessKey{
		AccessKeyID:     aws.StringValue(accessKey.AccessKeyId),
		SecretAccessKey: aws.StringValue(accessKey.SecretAccessKey),
	}, nil
}

// Validate SCP...
func (c *awsClient) ValidateSCP() (bool, error) {
	policyType := aws.String("SERVICE_CONTROL_POLICY")
	policies, err := c.orgClient.ListPolicies(&organizations.ListPoliciesInput{
		Filter: policyType,
	})
	if err != nil {
		return false, err
	}
	c.logger.Debug(nil, "ValidateSCP::ListPolicies\n%+v\n", policies)
	return true, nil
}
