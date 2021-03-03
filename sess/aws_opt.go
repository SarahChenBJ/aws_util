package sess

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

//NewSessionWithKeys for aws opt
func NewSessionWithKeys(region, accessKey, secretAccessKey string) (*session.Session, error) {
	log.Printf("[New AWS Session with keys] region=%s, accessKey=%s, secretAccessKey=%s", region, accessKey, secretAccessKey)
	config := aws.NewConfig().WithRegion(region).
		WithCredentials(credentials.NewStaticCredentials(accessKey, secretAccessKey, ""))
	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

//NewSession for aws opt
func NewSession() (*session.Session, error) {
	sess := session.Must(session.NewSession())
	return sess, nil
}

//NewSessionWithRole for aws opt
func NewSessionWithRole(role string) (*session.Session, *aws.Config) {
	log.Printf("[New AWS Session with Role] role=%s", role)
	sess, _ := NewSession()
	creds := stscreds.NewCredentials(sess, role)
	return sess, &aws.Config{Credentials: creds}
}

//NewSessionWithRegion for aws opt
func NewSessionWithRegion(region string) (*session.Session, error) {
	config := aws.NewConfig().WithRegion(region)
	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

//NewSessionWithSharedConfig creates a session that gets credential values from ~/.aws/credentials
// and the default region from ~/.aws/config
func NewSessionWithSharedConfig() *session.Session {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return sess
}
