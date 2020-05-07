// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/internal/awsutil"
	"github.com/aws/aws-sdk-go-v2/private/protocol"
	"github.com/aws/aws-sdk-go-v2/private/protocol/restxml"
	"github.com/aws/aws-sdk-go-v2/service/s3/internal/arn"
)

type DeleteBucketLifecycleInput struct {
	_ struct{} `type:"structure"`

	// The bucket name of the lifecycle to delete.
	//
	// Bucket is a required field
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`
}

// String returns the string representation
func (s DeleteBucketLifecycleInput) String() string {
	return awsutil.Prettify(s)
}

// Validate inspects the fields of the type to determine if they are valid.
func (s *DeleteBucketLifecycleInput) Validate() error {
	invalidParams := aws.ErrInvalidParams{Context: "DeleteBucketLifecycleInput"}

	if s.Bucket == nil {
		invalidParams.Add(aws.NewErrParamRequired("Bucket"))
	}

	if invalidParams.Len() > 0 {
		return invalidParams
	}
	return nil
}

func (s *DeleteBucketLifecycleInput) getBucket() (v string) {
	if s.Bucket == nil {
		return v
	}
	return *s.Bucket
}

// MarshalFields encodes the AWS API shape using the passed in protocol encoder.
func (s DeleteBucketLifecycleInput) MarshalFields(e protocol.FieldEncoder) error {

	if s.Bucket != nil {
		v := *s.Bucket

		metadata := protocol.Metadata{}
		e.SetValue(protocol.PathTarget, "Bucket", protocol.StringValue(v), metadata)
	}
	return nil
}

func (s *DeleteBucketLifecycleInput) getEndpointARN() (arn.Resource, error) {
	if s.Bucket == nil {
		return nil, fmt.Errorf("member Bucket is nil")
	}
	return parseEndpointARN(*s.Bucket)
}

func (s *DeleteBucketLifecycleInput) hasEndpointARN() bool {
	if s.Bucket == nil {
		return false
	}
	return arn.IsARN(*s.Bucket)
}

type DeleteBucketLifecycleOutput struct {
	_ struct{} `type:"structure"`
}

// String returns the string representation
func (s DeleteBucketLifecycleOutput) String() string {
	return awsutil.Prettify(s)
}

// MarshalFields encodes the AWS API shape using the passed in protocol encoder.
func (s DeleteBucketLifecycleOutput) MarshalFields(e protocol.FieldEncoder) error {
	return nil
}

const opDeleteBucketLifecycle = "DeleteBucketLifecycle"

// DeleteBucketLifecycleRequest returns a request value for making API operation for
// Amazon Simple Storage Service.
//
// Deletes the lifecycle configuration from the specified bucket. Amazon S3
// removes all the lifecycle configuration rules in the lifecycle subresource
// associated with the bucket. Your objects never expire, and Amazon S3 no longer
// automatically deletes any objects on the basis of rules contained in the
// deleted lifecycle configuration.
//
// To use this operation, you must have permission to perform the s3:PutLifecycleConfiguration
// action. By default, the bucket owner has this permission and the bucket owner
// can grant this permission to others.
//
// There is usually some time lag before lifecycle configuration deletion is
// fully propagated to all the Amazon S3 systems.
//
// For more information about the object expiration, see Elements to Describe
// Lifecycle Actions (https://docs.aws.amazon.com/AmazonS3/latest/dev/intro-lifecycle-rules.html#intro-lifecycle-rules-actions).
//
// Related actions include:
//
//    * PutBucketLifecycleConfiguration
//
//    * GetBucketLifecycleConfiguration
//
//    // Example sending a request using DeleteBucketLifecycleRequest.
//    req := client.DeleteBucketLifecycleRequest(params)
//    resp, err := req.Send(context.TODO())
//    if err == nil {
//        fmt.Println(resp)
//    }
//
// Please also see https://docs.aws.amazon.com/goto/WebAPI/s3-2006-03-01/DeleteBucketLifecycle
func (c *Client) DeleteBucketLifecycleRequest(input *DeleteBucketLifecycleInput) DeleteBucketLifecycleRequest {
	op := &aws.Operation{
		Name:       opDeleteBucketLifecycle,
		HTTPMethod: "DELETE",
		HTTPPath:   "/{Bucket}?lifecycle",
	}

	if input == nil {
		input = &DeleteBucketLifecycleInput{}
	}

	req := c.newRequest(op, input, &DeleteBucketLifecycleOutput{})
	req.Handlers.Unmarshal.Remove(restxml.UnmarshalHandler)
	req.Handlers.Unmarshal.PushBackNamed(protocol.UnmarshalDiscardBodyHandler)
	return DeleteBucketLifecycleRequest{Request: req, Input: input, Copy: c.DeleteBucketLifecycleRequest}
}

// DeleteBucketLifecycleRequest is the request type for the
// DeleteBucketLifecycle API operation.
type DeleteBucketLifecycleRequest struct {
	*aws.Request
	Input *DeleteBucketLifecycleInput
	Copy  func(*DeleteBucketLifecycleInput) DeleteBucketLifecycleRequest
}

// Send marshals and sends the DeleteBucketLifecycle API request.
func (r DeleteBucketLifecycleRequest) Send(ctx context.Context) (*DeleteBucketLifecycleResponse, error) {
	r.Request.SetContext(ctx)
	err := r.Request.Send()
	if err != nil {
		return nil, err
	}

	resp := &DeleteBucketLifecycleResponse{
		DeleteBucketLifecycleOutput: r.Request.Data.(*DeleteBucketLifecycleOutput),
		response:                    &aws.Response{Request: r.Request},
	}

	return resp, nil
}

// DeleteBucketLifecycleResponse is the response type for the
// DeleteBucketLifecycle API operation.
type DeleteBucketLifecycleResponse struct {
	*DeleteBucketLifecycleOutput

	response *aws.Response
}

// SDKResponseMetdata returns the response metadata for the
// DeleteBucketLifecycle request.
func (r *DeleteBucketLifecycleResponse) SDKResponseMetdata() *aws.Response {
	return r.response
}
