package s3

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	ExpiredDays = 7
)

// //InitS3Svc to get s3 client
// func InitS3Svc(role string) *s3.S3 {
// 	session, cfg := NewSessionWithRole(role)
// 	svc := s3.New(session, cfg)
// 	return svc
// }

// //GenPresignURLForS3 to get presign URL
// func GenPresignURLForS3(svc s3iface.S3API, queryID, bucket, prefix string) (string, error) {
// 	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
// 		Bucket: aws.String(bucket),
// 		Key:    aws.String("/" + prefix + "/" + queryID + ".csv"),
// 	})
// 	urlStr, err := req.Presign(ExpiredDays * 24 * time.Hour)

// 	log.Infof("[Presign URL For S3 file] bucket=%s, prefix=%s, urlStr=%s", bucket, prefix, urlStr)
// 	if err != nil {
// 		return "", fmt.Errorf("Failed to sign request: %s", err.Error())
// 	}
// 	return urlStr, err
// }

const (
	defaultGrowthCoefficient = 1.618
)

type s3Client struct {
	s3iface.S3API
	*s3manager.Uploader
	*s3manager.Downloader
}

type UploadReq struct {
	Key           string //required
	MD5           string //required
	ContentType   string //optional
	ContentLength int64  //optional
}

func NewS3Client(sess *session.Session) *s3Client {
	c := &s3Client{
		S3API:      s3.New(sess),
		Uploader:   s3manager.NewUploader(sess),
		Downloader: s3manager.NewDownloader(sess),
	}
	return c
}

func (c *s3Client) Exists(bucket, key string) bool {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	res, err := c.HeadObject(input)
	return res != nil && err == nil
}

// touch an empty file on s3://bucket/key
func (c *s3Client) Create(bucket, key string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	_, err := c.PutObject(input)
	return err
}

// create a file with content on s3://bucketName/bucketKey
func (c *s3Client) CreateWithContent(bucket, key, content string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(content),
	}
	_, err := c.PutObject(input)
	return err
}

func (c *s3Client) Remove(bucket, key string, retryTimes int) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	for retry := 0; retry < retryTimes; retry++ {
		if _, err := c.DeleteObject(input); err != nil {
			if retry == (retryTimes - 1) {
				return err
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	return nil
}

func (c *s3Client) Upload(source, bucket, key string, retryTimes int, compressEnabled bool) error {
	return c.UploadWithAcl(source, bucket, key, "", retryTimes, compressEnabled)
}

func (c *s3Client) UploadWithAcl(source, bucket, key, acl string, retryTimes int, compressEnabled bool) error {
	fileInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	switch mode := fileInfo.Mode(); {
	case mode.IsDir():
		return c.UploadFolderWithAcl(source, bucket, key, acl, retryTimes, compressEnabled)
	case mode.IsRegular():
		return c.UploadFileWithAcl(source, bucket, key, acl, retryTimes, compressEnabled)
	}

	return nil
}

func (c *s3Client) UploadFolder(srcFolder, bucket, keyPrefix string, retryTimes int, compressEnabled bool) error {
	return c.UploadFolderWithAcl(srcFolder, bucket, keyPrefix, "", retryTimes, compressEnabled)
}

func (c *s3Client) UploadFolderWithAcl(srcFolder, bucket, keyPrefix, acl string, retryTimes int, compressEnabled bool) error {
	files, err := ioutil.ReadDir(srcFolder)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcFile := filepath.Join(srcFolder, file.Name())
		key := filepath.Join(keyPrefix, file.Name())
		if err := c.UploadFileWithAcl(srcFile, bucket, key, acl, retryTimes, compressEnabled); err != nil {
			return err
		}
	}

	return nil
}

func (c *s3Client) UploadFile(srcFile, bucket, key string, retryTimes int, compressEnabled bool) error {
	return c.UploadFileWithAcl(srcFile, bucket, key, "", retryTimes, compressEnabled)
}

func (c *s3Client) UploadFileWithAcl(srcFile, bucket, key, acl string, retryTimes int, compressEnabled bool) error {
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", srcFile)
	}

	if compressEnabled && !strings.HasSuffix(srcFile, ".gz") {
		gzipFile := srcFile + ".gz"

		if err := c.compressFile(srcFile, gzipFile); err != nil {
			return fmt.Errorf("compress file %s failed, %s", srcFile, err.Error())
		}

		srcFile = gzipFile
		if !strings.HasSuffix(key, ".gz") {
			key = key + ".gz"
		}
	}

	srcFileBody, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer srcFileBody.Close()

	if err := c.doUploadWithAcl(srcFileBody, bucket, key, acl, retryTimes); err != nil {
		return err
	}

	return nil
}

func (c *s3Client) UploadContent(content []byte, bucket, key string, retryTimes int) error {
	return c.UploadContentWithAcl(content, bucket, key, "", retryTimes)
}

func (c *s3Client) UploadContentWithAcl(content []byte, bucket, key, acl string, retryTimes int) error {
	if err := c.doUploadWithAcl(bytes.NewReader(content), bucket, key, acl, retryTimes); err != nil {
		return err
	}

	return nil
}

func (c *s3Client) doUploadWithAcl(body io.Reader, bucket, key, acl string, retryTimes int) error {
	input := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	}

	if acl != "" {
		input.ACL = &acl
	}

	for retry := 0; retry < retryTimes; retry++ {
		if _, err := c.Uploader.Upload(input); err != nil {
			if retry == (retryTimes - 1) {
				return err
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	return nil
}

func (c *s3Client) DownloadFile(dst, bucket, key string, retryTimes int) error {
	fileDir := filepath.Dir(dst)
	if err := os.MkdirAll(fileDir, 0777); err != nil {
		return fmt.Errorf("failed to mkdir '%s': %s", fileDir, err.Error())
	}

	file, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %s", dst, err.Error())
	}
	defer file.Close()

	return c.doDownload(file, bucket, key, retryTimes)
}

func (c *s3Client) ReadFile(bucket, key string, retryTimes int) (string, error) {
	buff := &aws.WriteAtBuffer{
		GrowthCoeff: defaultGrowthCoefficient,
	}

	if err := c.doDownload(buff, bucket, key, retryTimes); err != nil {
		return "", err
	}

	return string(buff.Bytes()[:]), nil
}

func (c *s3Client) Copy(srcBucket, srcKey, dstBucket, dstKey string, retryTimes int) error {
	return c.CopyWithAcl(srcBucket, srcKey, dstBucket, dstKey, "", retryTimes)
}

func (c *s3Client) CopyWithAcl(srcBucket, srcKey, dstBucket, dstKey, acl string, retryTimes int) error {
	input := &s3.CopyObjectInput{
		CopySource: aws.String(srcBucket + "/" + srcKey),
		Bucket:     aws.String(dstBucket),
		Key:        aws.String(dstKey),
	}

	if acl != "" {
		input.SetACL(acl)
	}

	for retry := 0; retry < retryTimes; retry++ {
		if _, err := c.CopyObject(input); err != nil {
			if retry == (retryTimes - 1) {
				return err
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	return nil
}

// "S3.CopyObject" can not copy objects which size is more than 5GB.
// If we want to copy an object more than 5GB, we must use multipart copy.
// We can copy all objects using this operation.
// This operation is more complex and needs more user-side control than "S3.CopyObject"
func (c *s3Client) MultipartCopyObject(srcBucket, srcKey, dstBucket, dstKey string, retryTimes int, partSize int64, nCur int) error {
	// TODO
	return nil
}

func (c *s3Client) List(bucket, keyPrefix string) ([]string, error) {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(keyPrefix),
	}
	output, err := c.ListObjects(input)
	if err != nil {
		return nil, err
	}

	rst := make([]string, 0)
	for {
		for _, object := range output.Contents {
			rst = append(rst, aws.StringValue(object.Key))
		}

		if !aws.BoolValue(output.IsTruncated) {
			break
		}

		lastKey := aws.StringValue(output.Contents[len(output.Contents)-1].Key)
		input.SetMarker(lastKey)
		output, err = c.ListObjects(input)
		if err != nil {
			return nil, err
		}
	}

	return rst, nil
}

// Return the "sub-dir" of given prefix.
// For example:
//   s3://bucket/a/b/f1.file
//   s3://bucket/a/b/f2.file
//   s3://bucket/a/c/f3.file
//   s3://bucket/a/f4.file
// ListSubDir(bucket, "a/") -> []string{"b", "c"}
func (c *s3Client) ListSubDir(bucket, prefix string) ([]string, error) {
	if prefix != "" && prefix[len(prefix)-1:] != "/" {
		prefix += "/"
	}

	input := &s3.ListObjectsInput{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}
	output, err := c.ListObjects(input)
	if err != nil {
		return nil, err
	}

	rst := make([]string, 0)
	for {
		for _, commonPrefix := range output.CommonPrefixes {
			pSplit := strings.Split(commonPrefix.String(), "/")
			rst = append(rst, pSplit[len(pSplit)-2])
		}

		if !aws.BoolValue(output.IsTruncated) || nil == output.NextMarker {
			break
		}

		input.SetMarker(aws.StringValue(output.NextMarker))
		output, err = c.ListObjects(input)
		if err != nil {
			return nil, err
		}
	}

	return rst, nil
}

func (c *s3Client) PathExists(bucket, prefix string) (bool, error) {
	if prefix != "" && prefix[len(prefix)-1:] != "/" {
		prefix += "/"
	}

	input := &s3.ListObjectsInput{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}
	output, err := c.ListObjects(input)
	if err != nil {
		return false, err
	}
	return len(output.CommonPrefixes) > 0, nil
}

func (c *s3Client) HeadObj(bucket, key string) (*s3.HeadObjectOutput, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	output, err := c.S3API.HeadObject(input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *s3Client) doDownload(w io.WriterAt, bucket, key string, retryTimes int) error {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	for retry := 0; retry < retryTimes; retry++ {
		if _, err := c.Downloader.Download(w, input); err != nil {
			if retry == (retryTimes - 1) {
				return err
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	return nil
}

func (c *s3Client) compressFile(srcFile, dstFile string) error {
	srcFileHandler, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer srcFileHandler.Close()

	dstFileHandler, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer dstFileHandler.Close()

	gzipWriter := gzip.NewWriter(dstFileHandler)
	defer gzipWriter.Close()

	const BufferSize = 1 << 26 // 64M
	buffer := make([]byte, BufferSize)
	for {
		nBytes, err := srcFileHandler.Read(buffer)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		gzipWriter.Write(buffer[:nBytes])
	}

	return nil
}

func (c *s3Client) GetDownloadURL(bucket, key string, expireTimeSecond int64, retryTimes int) (string, error) {
	getReq, _ := c.S3API.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	var url string
	for retry := 0; retry < retryTimes; retry++ {
		if link, _, err := getReq.PresignRequest(time.Duration(expireTimeSecond) * time.Second); err != nil {
			if retry == (retryTimes - 1) {
				return "", err
			}
			time.Sleep(time.Second)
		} else {
			url = link
			break
		}
	}
	return url, nil
}

func (c *s3Client) GetUploadURL(bucket string, req *UploadReq, expireTimeSecond int64, retryTimes int) (string, error) {
	if req == nil {
		return "", fmt.Errorf("invalid upload request")
	}

	var url string
	in := &s3.PutObjectInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(req.Key),
		ContentMD5: aws.String(req.MD5),
	}
	if len(req.ContentType) > 0 {
		in.ContentType = aws.String(req.ContentType)
	}
	if req.ContentLength >= 0 {
		in.ContentLength = aws.Int64(req.ContentLength)
	}

	putReq, _ := c.S3API.PutObjectRequest(in)
	for retry := 0; retry < retryTimes; retry++ {
		if link, _, err := putReq.PresignRequest(time.Duration(expireTimeSecond) * time.Second); err != nil {
			if retry == (retryTimes - 1) {
				return "", err
			}
			time.Sleep(time.Second)
		} else {
			url = link
			break
		}
	}
	return url, nil
}

func (c *s3Client) GetObjectTagging(bucket, key string, retryTimes int) (map[string]string, error) {
	getReq := &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	tagMap := make(map[string]string)
	for retry := 0; retry < retryTimes; retry++ {
		if tagging, err := c.S3API.GetObjectTagging(getReq); err != nil {
			if retry == (retryTimes - 1) {
				return nil, err
			}
			time.Sleep(time.Second)
		} else {
			tagMap = convertS3TagsToMap(tagging.TagSet)
			break
		}
	}

	return tagMap, nil
}

func (c *s3Client) PutObjectTagging(bucket, key string, tags map[string]string, retryTimes int) error {
	putReq := &s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(key),
		Tagging: convertMapToS3Tagging(tags),
	}

	for retry := 0; retry < retryTimes; retry++ {
		if _, err := c.S3API.PutObjectTagging(putReq); err != nil {
			if retry == (retryTimes - 1) {
				return err
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	return nil
}

func (c *s3Client) DeleteObjectTagging(bucket, key string, retryTimes int) error {
	delReq := &s3.DeleteObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	for retry := 0; retry < retryTimes; retry++ {
		if _, err := c.S3API.DeleteObjectTagging(delReq); err != nil {
			if retry == (retryTimes - 1) {
				return err
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return nil
}

func convertS3TagsToMap(tags []*s3.Tag) map[string]string {
	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[*tag.Key] = *tag.Value
	}
	return tagMap
}

func convertMapToS3Tagging(tags map[string]string) *s3.Tagging {
	tagSet := []*s3.Tag{}
	for k, v := range tags {
		tagSet = append(tagSet, &s3.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	return &s3.Tagging{TagSet: tagSet}
}
