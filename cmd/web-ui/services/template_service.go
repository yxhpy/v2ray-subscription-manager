package services

import (
	"bytes"
	"html/template"
	"path/filepath"
)

// TemplateServiceImpl 模板服务实现
type TemplateServiceImpl struct {
	templateDir string
}

// NewTemplateService 创建模板服务
func NewTemplateService(templateDir string) TemplateService {
	return &TemplateServiceImpl{
		templateDir: templateDir,
	}
}

// RenderIndex 渲染主页
func (t *TemplateServiceImpl) RenderIndex() (string, error) {
	templatePath := filepath.Join(t.templateDir, "index.html")

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
