package s3

// import (
// 	"archive/zip"
// 	"bytes"
// 	"context"
// 	"encoding/csv"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"path"
// 	"reflect"
// 	"runtime/debug"
// 	"strings"
// 	"time"

// 	"github.com/aws-opt-go/utils/util"
// 	"github.com/grpc-ecosystem/grpc-gateway/runtime"
// 	"github.com/jmoiron/sqlx/reflectx"
// 	"github.com/tealeg/xlsx"
// )

// const (
// 	//LINK is a link for file name
// 	LINK      = "_"
// 	TagJson   = "json"
// 	TagExport = "export"
// 	TagEmbed  = "embed"
// )

// //Jsonpb is a shared variable for all of json converting
// var Jsonpb runtime.JSONPb

// // DeepReflectData <fieldName,fieldValue> pair
// type DeepReflectData struct {
// 	V reflect.Value
// 	F reflect.StructField
// }

// //Exporter interface, xls and csv use different lib
// //BaseExporter provide a common method, service can use it directly after import
// type Exporter interface {
// 	ExportXLSX() (string, error)
// 	ExportCSV() (string, error)
// 	AppendCSV() (string, error)
// 	AsyncExport()
// 	ConvertToData() ([]byte, error)
// 	ExportS3() error
// 	AppendS3() error
// 	AppendS3Done() error
// 	DownloadS3(ctx context.Context) error
// }

// //BaseExporter implements Exporter and provide basic export methods
// type BaseExporter struct {
// 	S3Config
// 	FileName        string
// 	FilePath        string
// 	SheetName       string
// 	Title           string   // Metric name exported
// 	SubTitle        string   //generally it is runtime
// 	ExtraInfos      []string // extra information below title
// 	Headers         [][]string
// 	DataHeader      []string
// 	Data            *bytes.Buffer
// 	Values          [][]string //detail records to fill in
// 	IsUploadCreated bool
// 	zipWriter       *zip.Writer
// }

// type S3Config struct {
// 	Bucket     string
// 	S3Client   *S3Client
// 	ExpireTime string
// }

// // InitS3Client init export s3 client, can optional set expireTime(default 7 days)
// func (e *BaseExporter) InitS3Client(bucket string, s3 *S3Client, expireTime ...string) {
// 	// default expire time 7 days
// 	expire := S3Expire7Days
// 	// if client set expire time, use this param.
// 	if len(expireTime) > 0 {
// 		expire = expireTime[0]
// 	}
// 	e.S3Config = S3Config{
// 		Bucket:     bucket,
// 		S3Client:   s3,
// 		ExpireTime: expire,
// 	}
// }

// // CompressData compress data and append upload
// func (e *BaseExporter) CompressData(data CompressDataItem) error {
// 	if e.Data == nil {
// 		e.Data = bytes.NewBuffer([]byte{})
// 	} else {
// 		e.Data.Reset()
// 	}
// 	if e.zipWriter == nil {
// 		e.zipWriter = zip.NewWriter(e.Data)
// 	}
// 	fileHeader := &zip.FileHeader{
// 		Name:   data.Name,
// 		Method: zip.Deflate,
// 	}
// 	fileHeader.SetModTime(time.Now())
// 	writer, err := e.zipWriter.CreateHeader(fileHeader)
// 	if err != nil {
// 		return err
// 	}
// 	io.Copy(writer, bytes.NewReader(data.Data))
// 	e.zipWriter.Flush()
// 	s3Client := e.S3Client
// 	bucket := e.Bucket
// 	if s3Client == nil || bucket == "" {
// 		return fmt.Errorf("Exporter S3 Client is nil or Bucket is empty", StructToString(s3Client), bucket)
// 	}
// 	key := path.Join(e.FilePath, e.FileName)
// 	if key == "" {
// 		return fmt.Errorf("Filepath & FileName is empty")
// 	}
// 	return s3Client.AppendFile(e.Data.Bytes(), bucket, key)
// }

// // CompressDataDone compress and append file done
// func (e *BaseExporter) CompressDataDone() error {
// 	if e.zipWriter == nil {
// 		return nil
// 	}
// 	if e.Data == nil {
// 		e.Data = bytes.NewBuffer([]byte{})
// 	} else {
// 		e.Data.Reset()
// 	}
// 	e.zipWriter.Close()
// 	s3Client := e.S3Client
// 	bucket := e.Bucket
// 	if s3Client == nil || bucket == "" {
// 		return fmt.Errorf("Exporter S3 Client is nil or Bucket is empty", StructToString(s3Client), bucket)
// 	}
// 	key := path.Join(e.FilePath, e.FileName)
// 	if key == "" {
// 		return fmt.Errorf("Filepath & FileName is empty")
// 	}
// 	s3Client.AppendFile(e.Data.Bytes(), bucket, key)
// 	err := s3Client.AppendFileDone(bucket, key)
// 	if err != nil {
// 		return err
// 	}
// 	return e.PutExpireTagging(e.ExpireTime)
// }

// // BuildFilePath build export file path
// func (e *BaseExporter) BuildFilePath(filePath string) {
// 	e.FilePath = filePath
// }

// // BuildFileName build export file path
// func (e *BaseExporter) BuildFileName(fileName string) {
// 	e.FileName = fileName
// }

// // GenerateDataByHeaders get data from bean by header array(field tag) and append to data array.
// func (e *BaseExporter) GenerateDataByHeaders(data interface{}, headers []string, inTag ...string) []string {
// 	tag := TagJson
// 	if len(inTag) > 0 {
// 		tag = inTag[0]
// 	}
// 	re := make([]string, 0)
// 	for _, v := range headers {
// 		value := getField(data, v, tag)
// 		var item string
// 		switch value.(type) {
// 		case string:
// 			item = value
// 		case time.Time:
// 			item = util.DateToString(value)

// 		default:
// 			item = util.GetString(value)
// 		}
// 		re = append(re, item)
// 	}
// 	return re
// }

// // get all <fieldName,fieldValue> pair of a struct
// // will deep search sub struct if field has "embed" tag
// func deepFields(iface interface{}) []*DeepReflectData {
// 	fields := make([]*DeepReflectData, 0)
// 	ifv := reflect.Indirect(reflect.ValueOf(iface))
// 	ift := reflect.TypeOf(iface)
// 	if ift.Kind() == reflect.Ptr {
// 		ift = ift.Elem()
// 	}
// 	if ifv.Kind() != reflect.Struct {
// 		log.Errorf("Can not get deep field of non struct")
// 		return fields
// 	}
// 	for i := 0; i < ift.NumField(); i++ {
// 		v := ifv.Field(i)
// 		f := ift.Field(i)
// 		rd := &DeepReflectData{
// 			v,
// 			f,
// 		}
// 		switch v.Kind() {
// 		case reflect.Struct:
// 			_, hasEmbed := f.Tag.Lookup(TagEmbed)
// 			if hasEmbed {
// 				fields = append(fields, deepFields(v.Interface())...)
// 			} else {
// 				fields = append(fields, rd)
// 			}
// 		default:
// 			fields = append(fields, rd)
// 		}
// 	}

// 	return fields
// }

// // get bean's field value by specific tag
// func getField(s interface{}, field, tag string) interface{} {
// 	embedFields := deepFields(s)
// 	for _, ef := range embedFields {
// 		tag, hasTag := ef.F.Tag.Lookup(tag)
// 		if !hasTag {
// 			continue
// 		}
// 		if !ef.V.IsValid() {
// 			continue
// 		}
// 		if tag == field {
// 			return ef.V.Interface()
// 		}
// 	}
// 	return "N/A"
// }

// // GenerateDataArrayByHeaders get data from bean by header array(field tag) and append to data array.
// // only can be used by struct without nested structure
// func (e *BaseExporter) GenerateDataArrayByHeaders(data interface{}, headers []string, inTag ...string) [][]string {
// 	tag := TagJson
// 	if len(inTag) > 0 {
// 		tag = inTag[0]
// 	}
// 	re := make([][]string, 0)
// 	v := reflect.ValueOf(data)
// 	direct := reflect.Indirect(v)
// 	switch direct.Kind() {
// 	case reflect.Slice:
// 		for i := 0; i < direct.Len(); i++ {
// 			items := getDataByHeaders(direct.Index(i), headers, tag)
// 			re = append(re, items)
// 		}
// 	case reflect.Struct:
// 		items := getDataByHeaders(direct, headers, tag)
// 		re = append(re, items)
// 	}
// 	return re
// }

// func getDataByHeaders(direct reflect.Value, headers []string, tag string) []string {
// 	items := make([]string, 0)
// 	dataItem := direct.Interface()
// 	t := reflectx.Deref(reflect.TypeOf(dataItem))
// 	v := reflect.Indirect(reflect.ValueOf(dataItem))
// 	if v.Kind() != reflect.Struct {
// 		return nil
// 	}
// 	headerMap := make(map[string]interface{})
// 	for i := 0; i < t.NumField(); i++ {
// 		f := t.Field(i)
// 		d := v.Field(i)
// 		tag, hasTag := f.Tag.Lookup(tag)
// 		if !hasTag {
// 			continue
// 		}
// 		for _, field := range headers {
// 			if tag == field || strings.Split(tag, ",")[0] == field {
// 				headerMap[field] = d.Interface()
// 			}
// 		}
// 	}
// 	for _, header := range headers {
// 		value := headerMap[header]
// 		var item string
// 		switch value.(type) {
// 		// case bool:
// 		// 	item = util.GetString(value)
// 		// case int64:
// 		// 	item = util.GetString(value)
// 		// case int32:
// 		// 	item = util.GetString(value)
// 		// case int:
// 		// 	item = util.GetString(value)
// 		// case float64:
// 		// 	item = util.GetString(value)
// 		case string:
// 			item = value
// 		case time.Time:
// 			item = util.DateToString(value.(time.Time))

// 		default:
// 			item = util.GetString(value)
// 		}
// 		items = append(items, item)
// 	}
// 	return items
// }

// //ExportXLSX generate a xls file and return its path, data format is [][]string
// func (e *BaseExporter) ExportXLSX() (string, error) {
// 	//validation
// 	err := e.isInvalid()
// 	if err != nil {
// 		return "", err
// 	}
// 	// create file and sheet
// 	file := xlsx.NewFile()
// 	sheet, err := file.AddSheet(e.buildSheetName())
// 	if err != nil {
// 		log.Errorln("excel file add sheet err:", err)
// 		return "", err
// 	}

// 	row := sheet.AddRow()
// 	// fill title
// 	row.AddCell().Value = e.Title
// 	// fill run time
// 	row.AddCell().Value = e.SubTitle

// 	// fill extra info
// 	if e.ExtraInfos != nil && len(e.ExtraInfos) > 0 {
// 		for _, info := range e.ExtraInfos {
// 			sheet.AddRow().AddCell().Value = info
// 		}
// 	}
// 	// fill header
// 	row = sheet.AddRow()
// 	for _, head := range e.Headers {
// 		for _, cell := range head {
// 			row.AddCell().Value = cell
// 		}
// 	}
// 	//fill data
// 	for line := 0; line < len(e.Values); line++ {
// 		row = sheet.AddRow()
// 		for _, value := range e.Values[line] {
// 			row.AddCell().Value = value
// 		}
// 	}
// 	// save file
// 	url := path.Join(e.GetExportPath(), e.buildFileName("xlsx"))
// 	err = file.Save(url)
// 	if err != nil {
// 		log.Errorln("save excel file err:", err)
// 	}
// 	return url, err
// }

// //ExportXLS generate a xls file and return its path
// func (e *BaseExporter) ExportXLS() (string, error) {
// 	//validation
// 	err := e.isInvalid()
// 	if err != nil {
// 		return "", err
// 	}
// 	// create file and sheet
// 	file := xlsx.NewFile()
// 	sheet, err := file.AddSheet(e.buildSheetName())
// 	if err != nil {
// 		log.Errorln("excel file add sheet err:", err)
// 		return "", err
// 	}

// 	row := sheet.AddRow()
// 	// fill title
// 	if e.Title != "" {
// 		row.AddCell().Value = e.Title
// 	}
// 	// fill run time
// 	row.AddCell().Value = util.Now()
// 	// fill extra info
// 	if e.ExtraInfos != nil && len(e.ExtraInfos) > 0 {
// 		for _, info := range e.ExtraInfos {
// 			sheet.AddRow().AddCell().Value = info
// 		}
// 	}
// 	// fill header
// 	row = sheet.AddRow()
// 	for _, head := range e.Headers {
// 		for _, cell := range head {
// 			row.AddCell().Value = cell
// 		}
// 	}
// 	// fill data
// 	values := e.Values
// 	for line := 0; line < len(values); line++ {
// 		row = sheet.AddRow()
// 		for _, value := range values[line] {
// 			row.AddCell().Value = value
// 		}
// 	}
// 	// save file
// 	url := path.Join(e.GetExportPath(), e.buildFileName("xlsx"))
// 	err = file.Save(url)
// 	if err != nil {
// 		log.Errorln("save excel file err:", err)
// 	}
// 	return url, err
// }

// //ExportCSV generate csv file and return its path
// func (e *BaseExporter) ExportCSV() (string, error) {
// 	//validation
// 	err := e.isInvalid()
// 	if err != nil {
// 		return "", err
// 	}
// 	// create file and sheet
// 	url := path.Join(e.GetExportPath(), e.FileName)
// 	fout, err := os.Create(url)
// 	if err != nil {
// 		log.Errorln("csv create err:", err)
// 		return "", err
// 	}
// 	defer func() {
// 		_ = fout.Close()
// 	}()
// 	//set UTF8
// 	_, _ = fout.WriteString("\xEF\xBB\xBF")
// 	// fill title
// 	if e.Title != "" {
// 		_, _ = fout.WriteString(e.Title + ",")
// 	}
// 	// fill sub title
// 	if e.SubTitle != "" {
// 		_, _ = fout.WriteString(e.SubTitle + "\n")
// 	}

// 	// fill extra info
// 	if e.ExtraInfos != nil && len(e.ExtraInfos) > 0 {
// 		for _, info := range e.ExtraInfos {
// 			_, _ = fout.WriteString(info + "\n")
// 		}
// 	}

// 	w := csv.NewWriter(fout)
// 	// fill header
// 	_ = w.WriteAll(e.Headers)
// 	//convert and fill data
// 	if len(e.DataHeader) > 0 {
// 		_ = w.Write(e.DataHeader)
// 	}
// 	_ = w.WriteAll(e.Values)
// 	w.Flush()
// 	if err = w.Error(); err != nil {
// 		log.Errorln(err)
// 	}
// 	return url, err
// }

// // AppendCSV append csv file and return its path
// func (e *BaseExporter) AppendCSV() (string, error) {
// 	//validation
// 	err := e.isInvalid()
// 	if err != nil {
// 		log.Errorf("AppendCSV Err:%v", err)
// 		return "", err
// 	}
// 	// create file and sheet
// 	url := path.Join(e.GetExportPath(), e.FileName)
// 	if !pathExist(url) {
// 		log.Warnf("File %s not exist", url)
// 		return e.ExportCSV()
// 	}
// 	fout, err := os.OpenFile(url, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
// 	if err != nil {
// 		log.Errorln("csv create err:", err)
// 		return "", err
// 	}
// 	defer func() {
// 		_ = fout.Close()
// 	}()
// 	// fill title
// 	if e.Title != "" {
// 		_, _ = fout.WriteString(e.Title + ",")
// 	}
// 	// fill sub title
// 	if e.SubTitle != "" {
// 		_, _ = fout.WriteString(e.SubTitle + "\n")
// 	}

// 	// fill extra info
// 	if e.ExtraInfos != nil && len(e.ExtraInfos) > 0 {
// 		for _, info := range e.ExtraInfos {
// 			_, _ = fout.WriteString(info + "\n")
// 		}
// 	}

// 	w := csv.NewWriter(fout)
// 	// fill header
// 	_ = w.WriteAll(e.Headers)
// 	if len(e.DataHeader) > 0 {
// 		_ = w.Write(e.DataHeader)
// 	}
// 	_ = w.WriteAll(e.Values)
// 	w.Flush()
// 	if err = w.Error(); err != nil {
// 		log.Errorln(err)
// 	}
// 	return url, err
// }

// // AddBOM add UTF-8 BOM to file
// func (e *BaseExporter) AddBOM() ([]byte, error) {
// 	err := e.isInvalid()
// 	if err != nil {
// 		log.Errorf("AddBOM Err:%v", err)
// 		return nil, err
// 	}
// 	bf := bytes.NewBuffer([]byte{})
// 	bf.WriteString("\xEF\xBB\xBF")
// 	return bf.Bytes(), err
// }

// // ConvertToData append csv file and return its path
// func (e *BaseExporter) ConvertToData() ([]byte, error) {
// 	//validation
// 	err := e.isInvalid()
// 	if err != nil {
// 		log.Errorf("ConvertToData Err:%v", err)
// 		return nil, err
// 	}
// 	// if e.Data == nil {
// 	// 	e.Data = bytes.NewBuffer([]byte{})
// 	// }
// 	bf := bytes.NewBuffer([]byte{})
// 	if e.Title != "" {
// 		_, _ = bf.WriteString(e.Title + ",")
// 	}
// 	// fill sub title
// 	if e.SubTitle != "" {
// 		_, _ = bf.WriteString(e.SubTitle + "\n")
// 	}
// 	// fill extra info
// 	if e.ExtraInfos != nil && len(e.ExtraInfos) > 0 {
// 		for _, info := range e.ExtraInfos {
// 			_, _ = bf.WriteString(info + "\n")
// 		}
// 	}
// 	w := csv.NewWriter(bf)
// 	// fill header
// 	_ = w.WriteAll(e.Headers)
// 	if len(e.DataHeader) > 0 {
// 		_ = w.Write(e.DataHeader)
// 	}
// 	_ = w.WriteAll(e.Values)
// 	w.Flush()
// 	if err = w.Error(); err != nil {
// 		log.Errorln(err)
// 	}
// 	return bf.Bytes(), err
// }

// // ConvertToDataWithBOM convert data to []byte and add BOM
// func (e *BaseExporter) ConvertToDataWithBOM() ([]byte, error) {
// 	bom, err := e.AddBOM()
// 	if err != nil {
// 		return nil, err
// 	}
// 	values, err := e.ConvertToData()
// 	if err != nil {
// 		return nil, err
// 	}
// 	data := append(bom, values...)
// 	return data, nil
// }

// // ExportS3 export data to s3
// func (e *BaseExporter) ExportS3() error {
// 	s3Client := e.S3Client
// 	bucket := e.Bucket
// 	if s3Client == nil || bucket == "" {
// 		return fmt.Errorf("Exporter S3 Client is nil or Bucket is empty", StructToString(s3Client), bucket)
// 	}
// 	data, _ := e.ConvertToDataWithBOM()
// 	key := path.Join(e.FilePath, e.FileName)
// 	if key == "" {
// 		return fmt.Errorf("Filepath & FileName is empty")
// 	}
// 	err := s3Client.UploadData(data, bucket, key)
// 	if err != nil {
// 		return err
// 	}
// 	return e.PutExpireTagging(e.ExpireTime)
// }

// func (e *BaseExporter) MultiUpload(dateRange []map[string]string) error {
// 	//AppendS3
// 	for i, v := range dateRange {
// 		param.StartDate = v["StartDate"]
// 		param.EndDate = v["EndDate"]
// 		e := exportInstance.GenerateData(s.ctx, param)
// 		if e != nil {
// 			return e
// 		}
// 		err := exportInstance.Sort()
// 		if err != nil {
// 			log.Warnf("Breakdown Export Sort Err: %v", err)
// 		}
// 		// TO DO need refactor
// 		if i == 0 {
// 			be.Headers = exportInstance.BuildHeader(param)
// 			be.DataHeader = exportInstance.BuildDataHeader()
// 		} else {
// 			be.Headers = make([][]string, 0)
// 			be.DataHeader = make([]string, 0)
// 		}

// 		be.Values = be.ConvertToArray(exportInstance.ConvertToRenderData(param))

// 		_ = be.AppendS3()

// 		// for gc
// 		be.Headers = nil
// 		be.Values = nil
// 		debug.FreeOSMemory()
// 	}
// 	err = be.AppendS3Done()
// 	return err
// }

// // AppendS3 append data to S3(multi part upload)
// func (e *BaseExporter) AppendS3() error {
// 	s3Client := e.S3Client
// 	bucket := e.Bucket
// 	if s3Client == nil || bucket == "" {
// 		return fmt.Errorf("Exporter S3 Client is nil or Bucket is empty", StructToString(s3Client), bucket)
// 	}
// 	data := make([]byte, 0)
// 	if !e.IsUploadCreated {
// 		data, _ = e.ConvertToDataWithBOM()
// 		e.IsUploadCreated = true
// 	} else {
// 		data, _ = e.ConvertToData()
// 	}
// 	key := path.Join(e.FilePath, e.FileName)
// 	if key == "" {
// 		return fmt.Errorf("Filepath & FileName is empty")
// 	}
// 	return s3Client.AppendFile(data, bucket, key)
// }

// // AppendS3Done finish append data to S3(complete multi part upload)
// func (e *BaseExporter) AppendS3Done() error {
// 	s3Client := e.S3Client
// 	bucket := e.Bucket
// 	if s3Client == nil || bucket == "" {
// 		return fmt.Errorf("Exporter S3 Client is nil or Bucket is empty", StructToString(s3Client), bucket)
// 	}
// 	key := path.Join(e.FilePath, e.FileName)
// 	if key == "" {
// 		return fmt.Errorf("Filepath & FileName is empty")
// 	}
// 	err := s3Client.AppendFileDone(bucket, key)
// 	if err != nil {
// 		return err
// 	}
// 	return e.PutExpireTagging(e.ExpireTime)
// }

// // PutExpireTagging put object's expire tag
// func (e *BaseExporter) PutExpireTagging(expireTime string) error {
// 	s3Client := e.S3Client
// 	bucket := e.Bucket
// 	if s3Client == nil || bucket == "" {
// 		return fmt.Errorf("Exporter S3 Client is nil or Bucket is empty", StructToString(s3Client), bucket)
// 	}
// 	key := path.Join(e.FilePath, e.FileName)
// 	if key == "" {
// 		return fmt.Errorf("Filepath & FileName is empty")
// 	}
// 	return s3Client.PutObjectTagging(bucket, key, map[string]string{S3TagExpire: expireTime}, 3)
// }

// // DownloadS3 download file from S3
// func (e *BaseExporter) DownloadS3(ctx context.Context) error {
// 	s3Client := e.S3Client
// 	bucket := e.Bucket
// 	if s3Client == nil || bucket == "" {
// 		return fmt.Errorf("Exporter S3 Client is nil or Bucket is empty", StructToString(s3Client), bucket)
// 	}
// 	key := path.Join(e.FilePath, e.FileName)
// 	if key == "" {
// 		return fmt.Errorf("Filepath & FileName is empty")
// 	}
// 	_, err := s3Client.GetDownloadURL(ctx, bucket, key, S3DownloadTimeOut, 3)
// 	return err
// }

// func pathExist(_path string) bool {
// 	_, err := os.Stat(_path)
// 	if err != nil && os.IsNotExist(err) {
// 		return false
// 	}
// 	return true
// }

// //AsyncExport async build xls file and send to inbox
// func (e *BaseExporter) AsyncExport() {
// 	go e.ExportXLS()
// }
// func (e *BaseExporter) isInvalid() error {
// 	if e.FileName == "" || (len(e.Values) == 0 && len(e.Headers) == 0) {
// 		return fmt.Error(errcode.CODE_DATA_INVALID_ARG_VALUE, "export filename or header+data is empty!")
// 	}
// 	return nil
// }
// func (e *BaseExporter) buildFileName(ext string) string {
// 	str := fmt.Sprintf("%d_%s_%s.%s", e.NetworkID, e.FileName, util.NowWithUnderLine(), ext)
// 	str = strings.Replace(str, " ", LINK, -1)
// 	return str
// }

// func (e *BaseExporter) buildSheetName() string {
// 	str := fmt.Sprintf("%d_%s", e.NetworkID, e.SheetName)
// 	str = strings.Replace(str, " ", LINK, -1)
// 	return str
// }

// // ConvertToArray convert data to array
// func (e *BaseExporter) ConvertToArray(src interface{}) [][]string {
// 	return convertToArray(src)
// }

// func convertToArray(src interface{}) [][]string {
// 	arr := convertInterfaceToArray(src)
// 	dest := make([][]string, len(arr))
// 	for k, obj := range arr {
// 		srcValue := reflect.Indirect(reflect.ValueOf(obj))
// 		count := 0
// 		for i := 0; i < srcValue.NumField(); i++ {
// 			field := srcValue.Field(i)
// 			// process array data, such as metrics data of breakdown export
// 			if field.Kind() == reflect.Slice {
// 				count += field.Len()

// 			} else {
// 				if field.CanInterface() && field.IsValid() {
// 					count++
// 				}
// 			}

// 		}
// 		subArr := make([]string, count)
// 		c := 0
// 		for i := 0; i < srcValue.NumField(); i++ {
// 			field := srcValue.Field(i)
// 			// fieldVal := field.Interface()
// 			// process array data, such as metrics data of breakdown export
// 			if field.Kind() == reflect.Slice {
// 				for j := 0; j < field.Len(); j++ {
// 					subArr[c] = util.GetString(field.Index(j).Interface())
// 					c++
// 				}

// 			} else {
// 				if field.CanInterface() && field.IsValid() {
// 					subArr[c] = util.GetString(field.Interface())
// 					c++
// 				}
// 			}
// 		}
// 		dest[k] = subArr
// 	}
// 	return dest
// }

// // the result of biz method is interface{}, make sure convertObjectsToArray can work fine, have to convert interface to []interface
// func convertInterfaceToArray(slice interface{}) []interface{} {
// 	s := reflect.ValueOf(slice)
// 	if s.Kind() != reflect.Slice {
// 		log.Errorln("convertInterfaceToArray error: param is not slice. ")
// 		return nil
// 	}
// 	ret := make([]interface{}, s.Len())
// 	for i := 0; i < s.Len(); i++ {
// 		ret[i] = s.Index(i).Interface()
// 	}
// 	return ret
// }

// //RenderExportData convert []*bean to [][]string for export lib use, src must be bean's array
// func RenderExportData(headers []string, src interface{}) [][]string {
// 	var data [][]string
// 	arr := convertInterfaceToArray(src)
// 	for _, plc := range arr {
// 		b, _ := Jsonpb.Marshal(plc)
// 		//mapping plc json to map
// 		var beanMap map[string]interface{}
// 		//json.Unmarshal(b, &beanMap)
// 		//In Golang, if map's value type is interface,when you get a value which is int originally and word length>6,
// 		// the value will be convert to a float64 with scientific notation. eg: 15467732 => 1.5467732e+07
// 		dec := json.NewDecoder(bytes.NewBuffer(b))
// 		dec.UseNumber()
// 		_ = dec.Decode(&beanMap)
// 		var record []string
// 		//get matched data based on head
// 		for _, head := range headers {
// 			//1. get value, map's key is camel case, head is snake case, need to convert firstly
// 			//2. make sure convert correctly for long id
// 			value := fmt.Sprint(beanMap[util.SnakeToCamel(head)])
// 			record = append(record, value)
// 		}
// 		data = append(data, record)
// 	}
// 	return data
// }

// // Export for common export
// func Export(networkID int, title, subTitle, fileType string, headers [][]string, data [][]string) (url string, err error) {
// 	//Export is a common function
// 	// build export data
// 	exporter := new(BaseExporter)
// 	exporter.FileName = T("export.export_file_name")
// 	exporter.SheetName = "sheet"
// 	exporter.Title = title
// 	exporter.SubTitle = subTitle
// 	exporter.NetworkID = networkID
// 	exporter.Headers = headers
// 	exporter.Values = data
// 	// generate excel
// 	if fileType == FileXls {
// 		url, err = exporter.ExportXLSX()
// 	} else {
// 		url, err = exporter.ExportCSV()
// 	}
// 	return url, err
// }

// // GetExportPath get export file path. Has default export path.
// func (e *BaseExporter) GetExportPath() string {
// 	return ""
// }

// // EscapeSlash escape slash in file name/path
// func EscapeSlash(str string) string {
// 	return strings.Replace(str, "/", "_", -1)
// }
