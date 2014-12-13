package milo

import (
	"bytes"
	"errors"
	html "html/template"
	"path/filepath"
	"text/template"
)

type MsgRender struct {
	tplDir   string
	tplFuncs map[string]interface{}
}

func NewMsgRender(tplDir string) *MsgRender {
	m := &MsgRender{tplDir: tplDir}
	m.tplFuncs = make(map[string]interface{})
	return m
}

func (m *MsgRender) RegisterTemplateFunc(key string, fn interface{}) {
	m.tplFuncs[key] = fn
}

func (m *MsgRender) Render(data map[string]interface{}, tpls ...string) (string, error) {
	if len(tpls) < 1 {
		return "", errors.New("Template identifiers required to render.")
	}
	list := make([]string, 0)
	for _, elem := range tpls {
		list = append(list, filepath.Join(m.tplDir, elem))
	}

	if tpl, tplErr := template.New(filepath.Base(tpls[0])).ParseFiles(list...); tplErr != nil {
		return "", tplErr
	} else {
		output := bytes.NewBufferString("")
		if execErr := tpl.Execute(output, data); execErr != nil {
			return "", execErr
		}
		return output.String(), nil
	}
}

func (m *MsgRender) RenderHtml(data map[string]interface{}, tpls ...string) (string, error) {
	if len(tpls) < 1 {
		return "", errors.New("Template identifiers required to render.")
	}
	list := make([]string, 0)
	for _, elem := range tpls {
		list = append(list, filepath.Join(m.tplDir, elem))
	}

	if tpl, tplErr := html.New(filepath.Base(tpls[0])).ParseFiles(list...); tplErr != nil {
		return "", tplErr
	} else {
		output := bytes.NewBufferString("")
		if execErr := tpl.Execute(output, data); execErr != nil {
			return "", execErr
		}
		return output.String(), nil
	}
}
