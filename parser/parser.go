package parser

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"reflect"
	
	"github.com/daarlabs/hirokit/errs"
	"github.com/daarlabs/hirokit/form"
	"github.com/daarlabs/hirokit/util/convertx"
	"github.com/daarlabs/hirokit/util/filex"
	"github.com/daarlabs/hirokit/util/reqx"
)

type Parse interface {
	Query(key string, target any) error
	PathValue(key string, target any) error
	File(filename string) (form.Multipart, error)
	Files(filesnames ...string) ([]form.Multipart, error)
	Json(target any) error
	Text() (string, error)
	Xml(target any) error
	Url(target any) error
	Multiple() Parse
	
	MustQuery(key string, target any)
	MustPathValue(key string, target any)
	MustFile(filename string) form.Multipart
	MustFiles(filesnames ...string) []form.Multipart
	MustJson(target any)
	MustText() string
	MustXml(target any)
	MustUrl(target any)
}

type Parser struct {
	r        *http.Request
	limit    int64
	multiple bool
}

func New(r *http.Request, limit int64) *Parser {
	return &Parser{
		r:     r,
		limit: limit,
	}
}

func (p *Parser) Multiple() Parse {
	p.multiple = true
	return p
}

func (p *Parser) Query(key string, target any) error {
	q := p.r.URL.Query()
	qv, ok := q[key]
	if !ok {
		return nil
	}
	n := len(qv)
	if !p.multiple && n == 1 {
		return convertx.ConvertValue(qv[0], target)
	}
	if p.multiple || n > 1 {
		return convertx.ConvertSlice(qv, target)
	}
	return nil
}

func (p *Parser) MustQuery(key string, target any) {
	err := p.Query(key, target)
	if err != nil {
		panic(err)
	}
}

func (p *Parser) PathValue(key string, target any) error {
	pathValue := p.r.PathValue(key)
	if len(pathValue) == 0 {
		return ErrorPathValueMissing
	}
	if err := convertx.ConvertValue(pathValue, target); err != nil {
		return err
	}
	return nil
}

func (p *Parser) MustPathValue(key string, target any) {
	err := p.PathValue(key, target)
	if err != nil {
		panic(err)
	}
}

func (p *Parser) Url(target any) error {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Ptr {
		return errs.ErrorPointerTarget
	}
	v := reflect.ValueOf(target).Elem()
	for i := 0; i < t.Elem().NumField(); i++ {
		fieldInfo := t.Elem().Field(i)
		fieldValue := v.Field(i).Addr().Interface()
		if err := p.processQuery(fieldInfo, fieldValue); err != nil {
			return err
		}
		if err := p.processPathValue(fieldInfo, fieldValue); err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) MustUrl(target any) {
	err := p.Url(target)
	if err != nil {
		panic(err)
	}
}

func (p *Parser) Text() (string, error) {
	if p.r.Body == nil {
		return "", nil
	}
	bytes, err := io.ReadAll(p.r.Body)
	return string(bytes), err
}

func (p *Parser) MustText() string {
	r, err := p.Text()
	if err != nil {
		panic(err)
	}
	return r
}

func (p *Parser) Json(target any) error {
	if p.r.Body == nil {
		return nil
	}
	err := json.NewDecoder(p.r.Body).Decode(target)
	if err == io.EOF {
		return nil
	}
	return err
}

func (p *Parser) MustJson(target any) {
	err := p.Json(target)
	if err != nil {
		panic(err)
	}
}

func (p *Parser) Xml(value any) error {
	if p.r.Body == nil {
		return nil
	}
	return xml.NewDecoder(p.r.Body).Decode(value)
}

func (p *Parser) MustXml(target any) {
	err := p.Xml(target)
	if err != nil {
		panic(err)
	}
}

func (p *Parser) File(filename string) (form.Multipart, error) {
	err := p.parseMultipartForm()
	if err != nil {
		return form.Multipart{}, err
	}
	multiparts, err := p.createMultiparts(filename)
	if err != nil {
		return form.Multipart{}, err
	}
	if len(multiparts) == 0 {
		return form.Multipart{}, nil
	}
	return multiparts[0], nil
}

func (p *Parser) MustFile(filename string) form.Multipart {
	file, err := p.File(filename)
	if err != nil {
		panic(err)
	}
	return file
}

func (p *Parser) Files(filesname ...string) ([]form.Multipart, error) {
	err := p.parseMultipartForm()
	if err != nil {
		return []form.Multipart{}, err
	}
	multiparts, err := p.createMultiparts(filesname...)
	if err != nil {
		return []form.Multipart{}, err
	}
	return multiparts, nil
}

func (p *Parser) MustFiles(filesnames ...string) []form.Multipart {
	files, err := p.Files(filesnames...)
	if err != nil {
		panic(err)
	}
	return files
}

func (p *Parser) createMultiparts(filename ...string) ([]form.Multipart, error) {
	var fn string
	if len(filename) > 0 {
		fn = filename[0]
	}
	fnLen := len(fn)
	result := make([]form.Multipart, 0)
	for name, files := range p.r.MultipartForm.File {
		if fnLen > 0 && name != fn {
			continue
		}
		for _, file := range files {
			f, err := file.Open()
			if err != nil {
				return result, errors.Join(ErrorOpenFile, err)
			}
			data, err := io.ReadAll(f)
			if err != nil {
				return result, errors.Join(ErrorReadData, err)
			}
			result = append(
				result, form.Multipart{
					Key:    name,
					Name:   file.Filename,
					Type:   http.DetectContentType(data),
					Suffix: filex.GetSuffix(file.Filename),
					Data:   data,
				},
			)
		}
	}
	return result, nil
}

func (p *Parser) parseMultipartForm() error {
	if !reqx.IsMultipart(p.r) {
		return ErrorInvalidMultipart
	}
	return p.r.ParseMultipartForm(p.limit << 20)
}

func (p *Parser) processQuery(fieldInfo reflect.StructField, fieldValue any) error {
	queryKey := fieldInfo.Tag.Get("query")
	q, exists := p.r.URL.Query()[queryKey]
	if !exists || len(q) == 0 {
		return nil
	}
	if len(q) == 1 {
		return convertx.ConvertValue(q[0], fieldValue)
	}
	return convertx.ConvertSlice(q, fieldValue)
}

func (p *Parser) processPathValue(fieldInfo reflect.StructField, fieldValue any) error {
	pathKey := fieldInfo.Tag.Get("path")
	pathValue := p.r.PathValue(pathKey)
	if pathValue == "" {
		return nil
	}
	return convertx.ConvertValue(pathValue, fieldValue)
}
