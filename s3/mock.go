package s3

import (
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type MockS3Client struct {
	s3iface.S3API
}

func (c *MockS3Client) GetObjectRequest(input *s3.GetObjectInput) (req *request.Request, output *s3.GetObjectOutput) {
	httpReq, _ := http.NewRequest("GET", "https://localhost:8080/check_alive", strings.NewReader("test"))
	return &request.Request{
		HTTPRequest: httpReq,
		Operation:   &request.Operation{},
	}, &s3.GetObjectOutput{}
}
